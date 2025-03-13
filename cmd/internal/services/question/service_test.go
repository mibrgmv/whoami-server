package question

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"whoami-server/cmd/internal/models"
)

func TestEvaluateAnswers(t *testing.T) {
	quizID := int64(1)
	quiz := models.Quiz{
		ID:      quizID,
		Title:   "GTA Character Quiz",
		Results: []string{"Michael", "Franklin", "Trevor"},
	}

	questions := []models.Question{
		{
			ID:     101,
			QuizID: quizID,
			Body:   "Do you like drinking gasoline?",
			OptionsWeights: map[string][]float32{
				"Yes": {0.0, 0.0, 1.0},
				"No":  {0.5, 0.5, 0.0},
			},
		},
		{
			ID:     102,
			QuizID: quizID,
			Body:   "Do you like betraying your friends?",
			OptionsWeights: map[string][]float32{
				"Yes": {1.0, 0.0, 0.0},
				"No":  {0.0, 0.5, 0.5},
			},
		},
		{
			ID:     103,
			QuizID: quizID,
			Body:   "Are you good at math?",
			OptionsWeights: map[string][]float32{
				"Yes": {0.0, 0.0, 0.0},
				"No":  {0.0, 0.0, 0.0},
			},
		},
		{
			ID:     104,
			QuizID: quizID,
			Body:   "Are you fond of hip hop?",
			OptionsWeights: map[string][]float32{
				"Yes": {0.0, 0.1, 0.0},
				"No":  {0.5, 0.0, 0.5},
			},
		},
	}

	tests := []struct {
		name     string
		answers  []models.Answer
		expected []float32
		wantErr  bool
		errMsg   string
	}{
		{
			name: "All Trevor answers",
			answers: []models.Answer{
				{QuizID: quizID, QuestionID: 101, Body: "Yes"}, // Likes gasoline -> Trevor
				{QuizID: quizID, QuestionID: 102, Body: "No"},  // Doesn't betray friends -> Franklin/Trevor
				{QuizID: quizID, QuestionID: 103, Body: "No"},  // Not good at math -> No effect
				{QuizID: quizID, QuestionID: 104, Body: "No"},  // Not fond of hip hop -> Michael/Trevor
			},
			expected: []float32{0.5, 0.5, 2.0}, // Trevor has highest score
			wantErr:  false,
			errMsg:   "",
		},
		{
			name: "All Michael answers",
			answers: []models.Answer{
				{QuizID: quizID, QuestionID: 101, Body: "No"},  // Doesn't like gasoline -> Michael/Franklin
				{QuizID: quizID, QuestionID: 102, Body: "Yes"}, // Betrays friends -> Michael
				{QuizID: quizID, QuestionID: 103, Body: "Yes"}, // Good at math -> No effect
				{QuizID: quizID, QuestionID: 104, Body: "No"},  // Not fond of hip hop -> Michael/Trevor
			},
			expected: []float32{2.0, 0.5, 0.5}, // Michael has highest score
			wantErr:  false,
			errMsg:   "",
		},
		{
			name: "All Franklin answers",
			answers: []models.Answer{
				{QuizID: quizID, QuestionID: 101, Body: "No"},  // Doesn't like gasoline -> Michael/Franklin
				{QuizID: quizID, QuestionID: 102, Body: "No"},  // Doesn't betray friends -> Franklin/Trevor
				{QuizID: quizID, QuestionID: 103, Body: "Yes"}, // Good at math -> No effect
				{QuizID: quizID, QuestionID: 104, Body: "Yes"}, // Fond of hip hop -> Franklin
			},
			expected: []float32{0.5, 1.1, 0.5}, // Franklin has highest score
			wantErr:  false,
			errMsg:   "",
		},
		{
			name: "Missing question ID",
			answers: []models.Answer{
				{QuizID: quizID, QuestionID: 999, Body: "Yes"}, // Non-existent question ID
			},
			expected: nil,
			wantErr:  true,
			errMsg:   "question with ID 999 not found",
		},
		{
			name: "Invalid option",
			answers: []models.Answer{
				{QuizID: quizID, QuestionID: 101, Body: "Maybe"}, // Invalid option
			},
			expected: nil,
			wantErr:  true,
			errMsg:   "option 'Maybe' not found for question 101",
		},
		{
			name: "Quiz ID mismatch in answer",
			answers: []models.Answer{
				{QuizID: 999, QuestionID: 101, Body: "Yes"}, // Wrong quiz ID
			},
			expected: nil,
			wantErr:  true,
			errMsg:   "answer quiz ID does not match quiz ID",
		},
		{
			name:     "Empty answers",
			answers:  []models.Answer{},
			expected: []float32{0, 0, 0}, // No answers means no weights added
			wantErr:  false,
			errMsg:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := EvaluateAnswers(tt.answers, questions, quiz)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			assert.NoError(t, err)
			assert.True(t, reflect.DeepEqual(results, tt.expected),
				"Expected %v, got %v", tt.expected, results)
		})
	}
}

func TestEvaluateAnswers_QuestionQuizIDMismatch(t *testing.T) {
	quiz := models.Quiz{
		ID:      1,
		Title:   "GTA Character Quiz",
		Results: []string{"Michael", "Franklin", "Trevor"},
	}

	questions := []models.Question{
		{
			ID:     101,
			QuizID: 999, // Mismatched quiz ID
			Body:   "Question with wrong quiz ID",
			OptionsWeights: map[string][]float32{
				"Yes": {1.0, 0.0, 0.0},
			},
		},
	}

	answers := []models.Answer{
		{QuizID: 1, QuestionID: 101, Body: "Yes"},
	}

	results, err := EvaluateAnswers(answers, questions, quiz)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "question quiz ID does not match quiz ID")
	assert.Nil(t, results)
}

func TestEvaluateAnswers_WeightLengthMismatch(t *testing.T) {
	quiz := models.Quiz{
		ID:      1,
		Title:   "GTA Character Quiz",
		Results: []string{"Michael", "Franklin", "Trevor"},
	}

	questions := []models.Question{
		{
			ID:     101,
			QuizID: 1,
			Body:   "Question with wrong weights length",
			OptionsWeights: map[string][]float32{
				"Yes": {1.0, 0.0},
			},
		},
	}

	answers := []models.Answer{
		{QuizID: 1, QuestionID: 101, Body: "Yes"},
	}

	results, err := EvaluateAnswers(answers, questions, quiz)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "weights length for option 'Yes' does not match number of results")
	assert.Nil(t, results)
}
