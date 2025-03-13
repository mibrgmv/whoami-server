package question_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"reflect"
	"testing"
	"whoami-server/cmd/internal/models"
	"whoami-server/cmd/internal/services/question"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Add(ctx context.Context, questions []models.Question) ([]models.Question, error) {
	args := m.Called(ctx, questions)
	return args.Get(0).([]models.Question), args.Error(1)
}

func (m *MockRepository) Query(ctx context.Context, query question.Query) ([]models.Question, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]models.Question), args.Error(1)
}

func TestEvaluateAnswers(t *testing.T) {
	mockRepo := new(MockRepository)
	service := question.NewService(mockRepo)

	quizID := int64(1)
	quiz := models.Quiz{
		ID:      quizID,
		Title:   "GTA V Character Quiz",
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
				"Yes": {0.0, 1.0, 0.0},
				"No":  {0.5, 0.0, 0.5},
			},
		},
	}

	mockRepo.On("Query", mock.Anything, question.Query{QuizIds: []int64{quizID}}).Return(questions, nil)

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
				{QuizID: quizID, QuestionID: 101, Body: "Yes"}, // Likes drinking gasoline -> +1.0 Trevor
				{QuizID: quizID, QuestionID: 102, Body: "No"},  // Doesn't like betraying friends -> +0.5 Franklin/Trevor
				{QuizID: quizID, QuestionID: 103, Body: "No"},  // Not good at math -> No effect
				{QuizID: quizID, QuestionID: 104, Body: "No"},  // Not fond of hip hop -> +0.5 Michael/Trevor
			},
			expected: []float32{0.5, 0.5, 2.0}, // Trevor has the highest score
			wantErr:  false,
			errMsg:   "",
		},
		{
			name: "All Michael answers",
			answers: []models.Answer{
				{QuizID: quizID, QuestionID: 101, Body: "No"},  // Doesn't like drinking gasoline -> +0.5 Michael/Franklin
				{QuizID: quizID, QuestionID: 102, Body: "Yes"}, // Likes betraying friends -> +1.0 Michael
				{QuizID: quizID, QuestionID: 103, Body: "Yes"}, // Good at math -> No effect
				{QuizID: quizID, QuestionID: 104, Body: "No"},  // Not fond of hip hop -> +0.5 Michael/Trevor
			},
			expected: []float32{2.0, 0.5, 0.5}, // Michael has the highest score
			wantErr:  false,
			errMsg:   "",
		},
		{
			name: "All Franklin answers",
			answers: []models.Answer{
				{QuizID: quizID, QuestionID: 101, Body: "No"},  // Doesn't like drinking gasoline -> +0.5 Michael/Franklin
				{QuizID: quizID, QuestionID: 102, Body: "No"},  // Doesn't like betraying friends -> +0.5 Franklin/Trevor
				{QuizID: quizID, QuestionID: 103, Body: "Yes"}, // Good at math -> No effect
				{QuizID: quizID, QuestionID: 104, Body: "Yes"}, // Fond of hip hop -> +1.0 Franklin
			},
			expected: []float32{0.5, 2.0, 0.5}, // Franklin has the highest score
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
			ctx := context.Background()
			results, err := service.EvaluateAnswers(ctx, tt.answers, quiz)

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

// Additional test for question with mismatched quiz ID
func TestEvaluateAnswers_QuestionQuizIDMismatch(t *testing.T) {
	mockRepo := new(MockRepository)
	service := question.NewService(mockRepo)

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

	mockRepo.On("Query", mock.Anything, question.Query{QuizIds: []int64{1}}).Return(questions, nil)

	answers := []models.Answer{
		{QuizID: 1, QuestionID: 101, Body: "Yes"},
	}

	ctx := context.Background()
	results, err := service.EvaluateAnswers(ctx, answers, quiz)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "question quiz ID does not match quiz ID")
	assert.Nil(t, results)
}

// Test for weight length mismatch
func TestEvaluateAnswers_WeightLengthMismatch(t *testing.T) {
	mockRepo := new(MockRepository)
	service := question.NewService(mockRepo)

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

	mockRepo.On("Query", mock.Anything, question.Query{QuizIds: []int64{1}}).Return(questions, nil)

	answers := []models.Answer{
		{QuizID: 1, QuestionID: 101, Body: "Yes"},
	}

	ctx := context.Background()
	results, err := service.EvaluateAnswers(ctx, answers, quiz)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "weights length for option 'Yes' does not match number of results")
	assert.Nil(t, results)
}
