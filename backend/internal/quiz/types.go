package quiz

type quizExample struct {
	Written string `json:"written"`
	Reading string `json:"reading"`
	Meaning string `json:"meaning"`
}

type quizQuestion struct {
	ID             string               `json:"id"`
	Mode           string               `json:"mode"`
	Prompt         string               `json:"prompt"`
	Answer         string               `json:"answer"`
	Choices        []string             `json:"choices"`
	Example        quizExample          `json:"example"`
	EstimatedLevel string               `json:"estimatedLevel,omitempty"`
	ExerciseID     string               `json:"exerciseId,omitempty"`
	Task           string               `json:"task,omitempty"`
	Context        *quizSentenceContext `json:"context,omitempty"`
}

type quizSentenceContext struct {
	Before string `json:"before"`
	After  string `json:"after"`
}

type quizAnswer struct {
	QuestionID     string `json:"questionId"`
	SelectedAnswer string `json:"selectedAnswer"`
}

type quizSubmission struct {
	ID              string         `json:"id"`
	Mode            string         `json:"mode"`
	Kind            string         `json:"kind"`
	ParentAttemptID *string        `json:"parentAttemptId,omitempty"`
	StartedAt       string         `json:"startedAt"`
	CompletedAt     string         `json:"completedAt"`
	Questions       []quizQuestion `json:"questions"`
	Answers         []quizAnswer   `json:"answers"`
}

type quizResultItem struct {
	Question       quizQuestion `json:"question"`
	SelectedAnswer string       `json:"selectedAnswer"`
	Correct        bool         `json:"correct"`
}

type quizResult struct {
	CorrectCount int              `json:"correctCount"`
	Total        int              `json:"total"`
	Items        []quizResultItem `json:"items"`
}

type quizAttemptSummary struct {
	ID           string `json:"id"`
	Mode         string `json:"mode"`
	Kind         string `json:"kind"`
	CompletedAt  string `json:"completedAt"`
	CorrectCount int    `json:"correctCount"`
	Total        int    `json:"total"`
	ExerciseID   string `json:"exerciseId,omitempty"`
}

type savedQuizAttempt struct {
	quizAttemptSummary
	StartedAt       string     `json:"startedAt"`
	ParentAttemptID *string    `json:"parentAttemptId,omitempty"`
	Result          quizResult `json:"result"`
}
