package grpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"log"
	"whoami-server/cmd/whoami/internal/models"
	"whoami-server/cmd/whoami/internal/services/question"
	"whoami-server/cmd/whoami/internal/services/quiz"
	pbHistory "whoami-server/protogen/golang/history"
	pb "whoami-server/protogen/golang/question"
)

type QuestionService struct {
	service       *question.Service
	quizService   *quiz.Service
	historyClient pbHistory.QuizCompletionHistoryServiceClient
	pb.UnimplementedQuestionServiceServer
}

func NewService(service *question.Service, quizService *quiz.Service, historyServiceAddr string) (*QuestionService, error) {
	conn, err := grpc.NewClient(historyServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to history service: %w", err)
	}

	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			fmt.Printf("failed to close connection to history service: %v", err)
		}
	}(conn)

	historyClient := pbHistory.NewQuizCompletionHistoryServiceClient(conn)

	return &QuestionService{
		service:       service,
		quizService:   quizService,
		historyClient: historyClient,
	}, nil
}

func (s *QuestionService) BatchCreateQuestions(ctx context.Context, request *pb.BatchCreateQuestionsRequest) (*pb.BatchCreateQuestionsResponse, error) {
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
		if errors.Is(err, quiz.ErrQuizNotFound) {
			return nil, status.Errorf(codes.NotFound, "quiz not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to get quiz: %v", err)
	}

	createdQuestions, err := s.service.Add(ctx, quizID, questionsToCreate)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error creating questions: %v", err)
	}

	var pbCreatedQuestions []*pb.Question
	for _, q := range createdQuestions {
		pbCreatedQuestions = append(pbCreatedQuestions, q.ToProto())
	}

	return &pb.BatchCreateQuestionsResponse{
		Questions: pbCreatedQuestions,
	}, nil
}

// todo not paginated -> fix api
func (s *QuestionService) BatchGetQuestions(ctx context.Context, request *pb.BatchGetQuestionsRequest) (*pb.BatchGetQuestionsResponse, error) {
	quizID, err := uuid.Parse(request.QuizId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid quiz ID format: %v", err)
	}

	questions, err := s.service.GetByQuizID(ctx, quizID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get questions by quiz id: %v", err)
	}

	var pbQuestions []*pb.QuestionResponse
	for _, q := range questions {
		pbQuestions = append(pbQuestions, q.ToProtoWithoutWeights())
	}

	return &pb.BatchGetQuestionsResponse{
		Questions: pbQuestions,
	}, nil
}

func (s *QuestionService) EvaluateAnswers(ctx context.Context, request *pb.EvaluateAnswersRequest) (*pb.EvaluateAnswersResponse, error) {
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
		if errors.Is(err, quiz.ErrQuizNotFound) {
			return nil, status.Errorf(codes.NotFound, "quiz not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to get quiz: %v", err)
	}

	result, err := s.service.EvaluateAnswers(ctx, answers, q)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to evaluate answers: %v", err)
	}

	userIDStr, ok := ctx.Value("user_id").(string)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "user not authenticated")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID format: %v", err)
	}

	err = s.addToQuizCompletionHistory(ctx, userID, q.ID, result)
	if err != nil {
		log.Printf("failed to add to quiz completion history: %v", err)
	}

	return &pb.EvaluateAnswersResponse{Result: result}, nil
}

func (s *QuestionService) addToQuizCompletionHistory(ctx context.Context, userID, quizID uuid.UUID, result string) error {
	historyItem := &pbHistory.QuizCompletionHistoryItem{
		UserId:     userID.String(),
		QuizId:     quizID.String(),
		QuizResult: result,
	}

	request := &pbHistory.CreateItemRequest{
		Item: historyItem,
	}

	_, err := s.historyClient.CreateItem(ctx, request)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to add quiz history: %v", err)
	}

	return nil
}
