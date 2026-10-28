package main

// Fixtures only. Production starts from the saved SQLite review data.

// This explicit mapping covers only the demo vocabulary, not a full dictionary.
func demoKanjiMapping() map[string][]string {
	return map[string][]string{
		"図書館": {"図", "書", "館"},
		"書く":  {"書"},
		"読書":  {"読", "書"},
		"食べる": {"食"},
		"日本語": {"日", "本", "語"},
	}
}

// Whole-word demo readings are curated, not inferred from individual kanji.
func demoVocabularyReadings() map[string]string {
	return map[string]string{
		"図書館": "としょかん",
		"書く":  "かく",
		"読書":  "どくしょ",
		"食べる": "たべる",
		"日本語": "にほんご",
	}
}

// demoReviews returns fresh sample data each time the backend starts.
func demoReviews() []review {
	return []review{
		// Mixed results: difficulty 63/100 across four reviews.
		{Vocabulary: "図書館", Rating: "again"},
		{Vocabulary: "図書館", Rating: "hard"},
		{Vocabulary: "図書館", Rating: "good"},
		{Vocabulary: "図書館", Rating: "again"},
		// Frequent failures: difficulty 80/100 across five reviews.
		{Vocabulary: "書く", Rating: "again"},
		{Vocabulary: "書く", Rating: "again"},
		{Vocabulary: "書く", Rating: "hard"},
		{Vocabulary: "書く", Rating: "again"},
		{Vocabulary: "書く", Rating: "hard"},
		// Exactly at the practice threshold: difficulty 50/100.
		{Vocabulary: "読書", Rating: "hard"},
		{Vocabulary: "読書", Rating: "hard"},
		{Vocabulary: "読書", Rating: "hard"},
		{Vocabulary: "読書", Rating: "hard"},
		// Mostly successful: difficulty 13/100 across four reviews.
		{Vocabulary: "食べる", Rating: "good"},
		{Vocabulary: "食べる", Rating: "easy"},
		{Vocabulary: "食べる", Rating: "hard"},
		{Vocabulary: "食べる", Rating: "good"},
		// One success is still limited evidence, not proof of mastery.
		{Vocabulary: "日本語", Rating: "easy"},
	}
}
