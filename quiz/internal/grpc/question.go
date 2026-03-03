package grpc

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	libsgrpc "whoami-server/libs/grpc"
	"whoami-server/libs/kafka"
	"whoami-server/quiz/internal/models"
	"whoami-server/quiz/internal/service"
	questionv1 "whoami-server/quiz/pkg/protogen/question/v1"
)

type questionServer struct {
	questionService    service.QuestionService
	quizService        service.QuizService
	producer           kafka.Producer
	quizCompletedTopic string
	questionv1.UnimplementedQuestionServiceServer
}

func NewQuestionServer(
	questionService service.QuestionService,
	quizService service.QuizService,
	producer kafka.Producer,
	quizCompletedTopic string,
) questionv1.QuestionServiceServer {
	return &questionServer{
		questionService:    questionService,
		quizService:        quizService,
		producer:           producer,
		quizCompletedTopic: quizCompletedTopic,
	}
}

func (s *questionServer) BatchCreateQuestions(ctx context.Context, request *questionv1.BatchCreateQuestionsRequest) (*questionv1.BatchCreateQuestionsResponse, error) {
	var questionsToCreate []*models.Question
	for _, req := range request.Requests {
		if req.QuizId != request.QuizId {
			return nil, status.Error(codes.InvalidArgument, "quiz id does not match request quiz id")
		}

		q, err := models.QuestionToModel(req)
		if err != nil {
			log.Fatalf("parse error %v", err)
		}

		questionsToCreate = append(questionsToCreate, q)
	}

	quizID, err := uuid.Parse(request.QuizId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid quiz ID format: %v", err)
	}

	_, err = s.quizService.GetByID(ctx, quizID)
	if err != nil {
		if errors.Is(err, service.ErrQuizNotFound) {
			return nil, status.Errorf(codes.NotFound, "quiz not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to get quiz: %v", err)
	}

	createdQuestions, err := s.questionService.Add(ctx, quizID, questionsToCreate)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error creating questions: %v", err)
	}

	var pbCreatedQuestions []*questionv1.Question
	for _, q := range createdQuestions {
		pbCreatedQuestions = append(pbCreatedQuestions, q.ToProto())
	}

	return &questionv1.BatchCreateQuestionsResponse{
		Questions: pbCreatedQuestions,
	}, nil
}

func (s *questionServer) BatchGetQuestions(ctx context.Context, request *questionv1.BatchGetQuestionsRequest) (*questionv1.BatchGetQuestionsResponse, error) {
	quizID, err := uuid.Parse(request.QuizId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid quiz ID format: %v", err)
	}

	questions, err := s.questionService.GetByQuizID(ctx, quizID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get questions by quiz id: %v", err)
	}

	var pbQuestions []*questionv1.QuestionResponse
	for _, q := range questions {
		pbQuestions = append(pbQuestions, q.ToProtoWithoutWeights())
	}

	return &questionv1.BatchGetQuestionsResponse{
		Questions: pbQuestions,
	}, nil
}

func (s *questionServer) EvaluateAnswers(ctx context.Context, request *questionv1.EvaluateAnswersRequest) (*questionv1.EvaluateAnswersResponse, error) {
	var answers []models.Answer
	for _, answer := range request.Answers {
		modelAnswer, err := models.AnswerToModel(answer)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid answer format: %v", err)
		}
		answers = append(answers, *modelAnswer)
	}

	quizID, err := uuid.Parse(request.QuizId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid quiz ID format: %v", err)
	}

	q, err := s.quizService.GetByID(ctx, quizID)
	if err != nil {
		if errors.Is(err, service.ErrQuizNotFound) {
			return nil, status.Errorf(codes.NotFound, "quiz not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to get quiz: %v", err)
	}

	result, err := s.questionService.EvaluateAnswers(ctx, answers, q)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to evaluate answers: %v", err)
	}

	userIDStr, ok := ctx.Value(libsgrpc.UserIDKey).(string)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "user not authenticated")
	}

	if err := s.publishQuizCompleted(ctx, userIDStr, q.ID.String(), result); err != nil {
		log.Printf("failed to publish quiz completed event: %v", err)
	}

	return &questionv1.EvaluateAnswersResponse{Result: result}, nil
}

func (s *questionServer) publishQuizCompleted(ctx context.Context, userID, quizID, result string) error {
	event := kafka.QuizCompletedEvent{
		UserID:     userID,
		QuizID:     quizID,
		QuizResult: result,
	}

	return s.producer.Produce(ctx, s.quizCompletedTopic, userID, event)
}
