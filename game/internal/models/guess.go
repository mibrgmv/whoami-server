package models

import (
	"strings"

	"github.com/google/uuid"
	gamev1 "gordle/game/pkg/protogen/game/v1"
)

type LetterResult rune

const (
	LetterCorrect LetterResult = 'G'
	LetterPresent LetterResult = 'Y'
	LetterAbsent  LetterResult = 'B'
)

type Guess struct {
	ID            uuid.UUID `json:"guess_id"`
	SessionID     uuid.UUID `json:"session_id"`
	GuessWord     string    `json:"guess_word"`
	Result        string    `json:"result"`
	AttemptNumber int       `json:"attempt_number"`
}

func (g *Guess) ToProto() *gamev1.Guess {
	return &gamev1.Guess{
		Word:          g.GuessWord,
		Results:       parseResults(g.Result),
		AttemptNumber: int32(g.AttemptNumber),
	}
}

func parseResults(result string) []gamev1.LetterResult {
	results := make([]gamev1.LetterResult, len(result))
	for i, r := range result {
		switch r {
		case 'G':
			results[i] = gamev1.LetterResult_LETTER_RESULT_CORRECT
		case 'Y':
			results[i] = gamev1.LetterResult_LETTER_RESULT_PRESENT
		case 'B':
			results[i] = gamev1.LetterResult_LETTER_RESULT_ABSENT
		default:
			results[i] = gamev1.LetterResult_LETTER_RESULT_UNSPECIFIED
		}
	}
	return results
}

func EvaluateGuess(guess, target string) string {
	guess = strings.ToLower(guess)
	target = strings.ToLower(target)

	guessRunes := []rune(guess)
	targetRunes := []rune(target)

	results := make([]LetterResult, len(guessRunes))
	available := make(map[rune]int)

	for _, r := range targetRunes {
		available[r]++
	}

	for i := 0; i < len(guessRunes) && i < len(targetRunes); i++ {
		if guessRunes[i] == targetRunes[i] {
			results[i] = LetterCorrect
			available[guessRunes[i]]--
		}
	}

	for i := 0; i < len(guessRunes); i++ {
		if results[i] == LetterCorrect {
			continue
		}
		if available[guessRunes[i]] > 0 {
			results[i] = LetterPresent
			available[guessRunes[i]]--
		} else {
			results[i] = LetterAbsent
		}
	}

	var sb strings.Builder
	for _, r := range results {
		sb.WriteRune(rune(r))
	}
	return sb.String()
}
