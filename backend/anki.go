package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
)

type ankiClient struct {
	client   *http.Client
	endpoint string
}

type ankiDeckList struct {
	Decks []string `json:"decks"`
}

type ankiError struct {
	status  int
	message string
}

func (e *ankiError) Error() string { return e.message }

// call shares the read-only AnkiConnect request/response handling across actions.
func (a *ankiClient) call(ctx context.Context, action string, params any, result any) error {
	payload := struct {
		Action  string `json:"action"`
		Version int    `json:"version"`
		Params  any    `json:"params"`
	}{action, 6, params}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode AnkiConnect request: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, a.endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create AnkiConnect request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := a.client.Do(request)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		var networkError net.Error
		if errors.As(err, &networkError) && networkError.Timeout() {
			return &ankiError{504, "AnkiConnect timed out. Check that Anki is open and responsive, then retry."}
		}
		return &ankiError{503, "Cannot reach AnkiConnect. Open Anki Desktop with the add-on enabled, then retry."}
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return &ankiError{502, "AnkiConnect returned an HTTP error. Check Anki and retry."}
	}
	var envelope struct {
		Result json.RawMessage `json:"result"`
		Error  *string         `json:"error"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 16*1024*1024)).Decode(&envelope); err != nil {
		return &ankiError{502, "AnkiConnect returned unreadable data."}
	}
	if envelope.Error != nil {
		return &ankiError{502, "AnkiConnect: " + *envelope.Error}
	}
	if len(envelope.Result) == 0 || string(envelope.Result) == "null" {
		return &ankiError{502, "AnkiConnect returned incomplete data."}
	}
	if err := json.Unmarshal(envelope.Result, result); err != nil {
		return &ankiError{502, "AnkiConnect returned an unexpected data format."}
	}
	return nil
}

func writeAnkiError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	message := "Could not complete the Anki request. See the backend log for details."
	var upstream *ankiError
	if errors.As(err, &upstream) {
		status, message = upstream.status, upstream.message
	} else if errors.Is(err, context.DeadlineExceeded) {
		status, message = 504, "The import timed out. The previous saved data is still available; please retry."
	}
	log.Printf("Anki request: %v", err)
	http.Error(w, message, status)
}

func (a *ankiClient) listDecks(w http.ResponseWriter, r *http.Request) {
	var decks []string
	if err := a.call(r.Context(), "deckNames", map[string]any{}, &decks); err != nil {
		if r.Context().Err() == nil {
			writeAnkiError(w, err)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if err := json.NewEncoder(w).Encode(ankiDeckList{Decks: decks}); err != nil {
		log.Printf("write Anki deck list: %v", err)
	}
}
