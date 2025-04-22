package question_test

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"whoami-server/cmd/whoami/internal/models"
	"whoami-server/cmd/whoami/internal/services/question"
	"whoami-server/cmd/whoami/internal/services/question/mocks"
)

func TestEvaluateAnswers(t *testing.T) {
	mockRepo := new(mocks.MockRepository)
	mockCache := new(mocks.MockCache)
	service := question.NewService(mockRepo, mockCache)

	quizID := uuid.New()
	quiz := models.Quiz{
		ID:      quizID,
		Title:   "GTA V Character Quiz",
		Results: []string{"Michael", "Franklin", "Trevor"},
	}

	questionIDs := []uuid.UUID{
		uuid.New(),
		uuid.New(),
		uuid.New(),
		uuid.New(),
	}

	questions := []models.Question{
		{
			ID:     questionIDs[0],
			QuizID: quizID,
			Body:   "Do you like drinking gasoline?",
			OptionsWeights: map[string][]float32{
				"Yes": {0.0, 0.0, 1.0},
				"No":  {0.5, 0.5, 0.0},
			},
		},
		{
			ID:     questionIDs[1],
			QuizID: quizID,
			Body:   "Do you like betraying your friends?",
			OptionsWeights: map[string][]float32{
				"Yes": {1.0, 0.0, 0.0},
				"No":  {0.0, 0.5, 0.5},
			},
		},
		{
			ID:     questionIDs[2],
			QuizID: quizID,
			Body:   "Are you good at math?",
			OptionsWeights: map[string][]float32{
				"Yes": {0.0, 0.0, 0.0},
				"No":  {0.0, 0.0, 0.0},
			},
		},
		{
			ID:     questionIDs[3],
			QuizID: quizID,
			Body:   "Are you fond of hip hop?",
			OptionsWeights: map[string][]float32{
				"Yes": {0.0, 1.0, 0.0},
				"No":  {0.5, 0.0, 0.5},
			},
		},
	}

	cacheKey := "questions:quiz:" + quizID.String()

	mockCache.On("Get", mock.Anything, cacheKey, mock.AnythingOfType("*[]models.Question")).Return(errors.New("cache miss"))
	mockRepo.On("Query", mock.Anything, question.Query{QuizIds: []uuid.UUID{quizID}}).Return(questions, nil)
	mockCache.On("Set", mock.Anything, cacheKey, mock.AnythingOfType("*[]models.Question")).Return(nil)

	tests := []struct {
		name     string
		answers  []models.Answer
		expected string
		wantErr  bool
		errMsg   string
	}{
		{
			name: "All Trevor answers",
			answers: []models.Answer{
				{QuizID: quizID, QuestionID: questionIDs[0], Body: "Yes"}, // Likes drinking gasoline -> +1.0 Trevor
				{QuizID: quizID, QuestionID: questionIDs[1], Body: "No"},  // Doesn't like betraying friends -> +0.5 Franklin/Trevor
				{QuizID: quizID, QuestionID: questionIDs[2], Body: "No"},  // Not good at math -> No effect
				{QuizID: quizID, QuestionID: questionIDs[3], Body: "No"},  // Not fond of hip hop -> +0.5 Michael/Trevor
			},
			expected: "Trevor", // Trevor has the highest score
			wantErr:  false,
			errMsg:   "",
		},
		{
			name: "All Michael answers",
			answers: []models.Answer{
				{QuizID: quizID, QuestionID: questionIDs[0], Body: "No"},  // Doesn't like drinking gasoline -> +0.5 Michael/Franklin
				{QuizID: quizID, QuestionID: questionIDs[1], Body: "Yes"}, // Likes betraying friends -> +1.0 Michael
				{QuizID: quizID, QuestionID: questionIDs[2], Body: "Yes"}, // Good at math -> No effect
				{QuizID: quizID, QuestionID: questionIDs[3], Body: "No"},  // Not fond of hip hop -> +0.5 Michael/Trevor
			},
			expected: "Michael", // Michael has the highest score
			wantErr:  false,
			errMsg:   "",
		},
		{
			name: "All Franklin answers",
			answers: []models.Answer{
				{QuizID: quizID, QuestionID: questionIDs[0], Body: "No"},  // Doesn't like drinking gasoline -> +0.5 Michael/Franklin
				{QuizID: quizID, QuestionID: questionIDs[1], Body: "No"},  // Doesn't like betraying friends -> +0.5 Franklin/Trevor
				{QuizID: quizID, QuestionID: questionIDs[2], Body: "Yes"}, // Good at math -> No effect
				{QuizID: quizID, QuestionID: questionIDs[3], Body: "Yes"}, // Fond of hip hop -> +1.0 Franklin
			},
			expected: "Franklin", // Franklin has the highest score
			wantErr:  false,
			errMsg:   "",
		},
		{
			name: "Missing question ID",
			answers: []models.Answer{
				{QuizID: quizID, QuestionID: uuid.New(), Body: "Yes"}, // Non-existent question ID
			},
			expected: "",
			wantErr:  true,
			// Fix 3: Update error message to match implementation format
			errMsg: "question with ID",
		},
		{
			name: "Invalid option",
			answers: []models.Answer{
				{QuizID: quizID, QuestionID: questionIDs[0], Body: "Maybe"}, // Invalid option
			},
			expected: "",
			wantErr:  true,
			errMsg:   "option 'Maybe' not found for question",
		},
		{
			name: "Quiz ID mismatch in answer",
			answers: []models.Answer{
				{QuizID: uuid.New(), QuestionID: questionIDs[0], Body: "Yes"}, // Wrong quiz ID
			},
			expected: "",
			wantErr:  true,
			errMsg:   "answer quiz ID does not match quiz ID",
		},
		{
			name:     "Empty answers",
			answers:  []models.Answer{},
			expected: "Michael", // No answers means weights are all 0, so first result is returned
			wantErr:  false,
			errMsg:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := service.EvaluateAnswers(ctx, tt.answers, &quiz)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result,
				"Expected %v, got %v", tt.expected, result)
		})
	}
}

// Additional test for question with mismatched quiz ID
func TestEvaluateAnswers_QuestionQuizIDMismatch(t *testing.T) {
	mockRepo := new(mocks.MockRepository)
	mockCache := new(mocks.MockCache)
	service := question.NewService(mockRepo, mockCache)

	quizID := uuid.New()
	wrongQuizID := uuid.New()
	questionID := uuid.New()

	quiz := models.Quiz{
		ID:      quizID,
		Title:   "GTA Character Quiz",
		Results: []string{"Michael", "Franklin", "Trevor"},
	}

	questions := []models.Question{
		{
			ID:     questionID,
			QuizID: wrongQuizID, // Mismatched quiz ID
			Body:   "Question with wrong quiz ID",
			OptionsWeights: map[string][]float32{
				"Yes": {1.0, 0.0, 0.0},
			},
		},
	}

	cacheKey := "questions:quiz:" + quizID.String()

	mockCache.On("Get", mock.Anything, cacheKey, mock.AnythingOfType("*[]models.Question")).Return(errors.New("cache miss"))

	mockRepo.On("Query", mock.Anything, question.Query{QuizIds: []uuid.UUID{quizID}}).Return(questions, nil)

	mockCache.On("Set", mock.Anything, cacheKey, mock.AnythingOfType("*[]models.Question")).Return(nil)

	answers := []models.Answer{
		{QuizID: quizID, QuestionID: questionID, Body: "Yes"},
	}

	ctx := context.Background()
	result, err := service.EvaluateAnswers(ctx, answers, &quiz)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "question quiz ID does not match quiz ID")
	assert.Empty(t, result)
}

// Test for weight length mismatch
func TestEvaluateAnswers_WeightLengthMismatch(t *testing.T) {
	mockRepo := new(mocks.MockRepository)
	mockCache := new(mocks.MockCache)
	service := question.NewService(mockRepo, mockCache)

	quizID := uuid.New()
	questionID := uuid.New()

	quiz := models.Quiz{
		ID:      quizID,
		Title:   "GTA Character Quiz",
		Results: []string{"Michael", "Franklin", "Trevor"},
	}

	questions := []models.Question{
		{
			ID:     questionID,
			QuizID: quizID,
			Body:   "Question with wrong weights length",
			OptionsWeights: map[string][]float32{
				"Yes": {1.0, 0.0}, // Only two weights for three results
			},
		},
	}

	cacheKey := "questions:quiz:" + quizID.String()

	mockCache.On("Get", mock.Anything, cacheKey, mock.AnythingOfType("*[]models.Question")).Return(errors.New("cache miss"))

	mockRepo.On("Query", mock.Anything, question.Query{QuizIds: []uuid.UUID{quizID}}).Return(questions, nil)

	mockCache.On("Set", mock.Anything, cacheKey, mock.AnythingOfType("*[]models.Question")).Return(nil)

	answers := []models.Answer{
		{QuizID: quizID, QuestionID: questionID, Body: "Yes"},
	}

	ctx := context.Background()
	result, err := service.EvaluateAnswers(ctx, answers, &quiz)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "weights length for option 'Yes' does not match number of results")
	assert.Empty(t, result)
}

func TestEvaluateAnswers_CacheHit(t *testing.T) {
	mockRepo := new(mocks.MockRepository)
	mockCache := new(mocks.MockCache)
	service := question.NewService(mockRepo, mockCache)

	quizID := uuid.New()
	questionID := uuid.New()

	quiz := models.Quiz{
		ID:      quizID,
		Title:   "GTA Character Quiz",
		Results: []string{"Michael", "Franklin", "Trevor"},
	}

	cachedQuestions := []models.Question{
		{
			ID:     questionID,
			QuizID: quizID,
			Body:   "Do you like drinking gasoline?",
			OptionsWeights: map[string][]float32{
				"Yes": {0.0, 0.0, 1.0},
				"No":  {0.5, 0.5, 0.0},
			},
		},
	}

	cacheKey := "questions:quiz:" + quizID.String()

	mockCache.On("Get", mock.Anything, cacheKey, mock.AnythingOfType("*[]models.Question")).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]models.Question)
		*dest = cachedQuestions
	}).Return(nil)

	answers := []models.Answer{
		{QuizID: quizID, QuestionID: questionID, Body: "Yes"},
	}

	ctx := context.Background()
	result, err := service.EvaluateAnswers(ctx, answers, &quiz)

	assert.NoError(t, err)
	assert.Equal(t, "Trevor", result)
	mockRepo.AssertNotCalled(t, "Query")
}
