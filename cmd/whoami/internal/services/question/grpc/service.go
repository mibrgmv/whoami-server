package grpc

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
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

func (s *QuestionService) Add(stream pb.QuestionService_AddStreamServer) error {
	var questions []models.Question

	for {
		req, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return err
		}

		q := models.ToModel(&pb.Question{
			Id:             req.Id,
			QuizId:         req.QuizId,
			Body:           req.Body,
			OptionsWeights: req.OptionsWeights,
		})
		questions = append(questions, *q)
	}

	addedQuestions, err := s.service.Add(stream.Context(), questions)
	if err != nil {
		return err
	}

	for _, q := range addedQuestions {
		if err := stream.Send(q.ToProto()); err != nil {
			return err
		}
	}

	return nil
}

func (s *QuestionService) GetAll(empty *emptypb.Empty, stream pb.QuestionService_GetAllStreamServer) error {
	questions, err := s.service.Query(stream.Context(), question.Query{})
	if err != nil {
		return err
	}

	done := make(chan bool)

	go func() {
		defer close(done)
		for _, q := range questions {
			if err := stream.Send(q.ToProto()); err != nil {
				log.Fatalf("can not send %v", err)
			}
		}
	}()

	<-done
	return nil
}

func (s *QuestionService) GetByQuizID(request *pb.GetByQuizIDRequest, stream pb.QuestionService_GetByQuizIDServer) error {
	questions, err := s.service.GetByQuizID(stream.Context(), request.QuizId)
	if err != nil {
		return err
	}

	for _, q := range questions {
		if err := stream.Send(q.ToProto()); err != nil {
			return err
		}
	}

	return nil
}

func (s *QuestionService) EvaluateAnswers(ctx context.Context, request *pb.EvaluateAnswersRequest) (*pb.EvaluateAnswersResponse, error) {
	var answers []models.Answer
	for _, answer := range request.Answers {
		answers = append(answers, models.Answer{
			QuizID:     answer.QuizId,
			QuestionID: answer.QuestionId,
			Body:       answer.Body,
		})
	}

	q, err := s.quizService.GetByID(ctx, request.QuizId)
	if err != nil {
		return nil, err
	}

	result, err := s.service.EvaluateAnswers(ctx, answers, q)
	if err != nil {
		return nil, err
	}

	// todo use jwt service
	userID, ok := ctx.Value("user_id").(int64)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "user not authenticated")
	}

	historyError := s.addToQuizCompletionHistory(ctx, userID, q.ID, result)
	if historyError != nil {
		return nil, historyError
	}

	return &pb.EvaluateAnswersResponse{
		Result:      result,
		Description: "", // todo
	}, nil
}

func (s *QuestionService) addToQuizCompletionHistory(ctx context.Context, userID, quizID int64, result string) error {
	stream, err := s.historyClient.Add(ctx)
	if err != nil {
		return err
	}

	var createdItems []pbHistory.QuizCompletionHistoryItem
	done := make(chan bool)

	go func() {
		if err := stream.Send(&pbHistory.QuizCompletionHistoryItem{
			UserId:     userID,
			QuizId:     quizID,
			QuizResult: result,
		}); err != nil {
			log.Fatalf("can not send %v", err)
		}
		if err := stream.CloseSend(); err != nil {
			log.Fatalf("can not close %v", err)
		}
	}()

	// todo read result ignored
	go func() {
		for {
			resp, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				done <- true
				return
			}
			if err != nil {
				log.Fatalf("cannot receive %v", err)
			}
			createdItems = append(createdItems, pbHistory.QuizCompletionHistoryItem{
				Id:         resp.Id,
				UserId:     resp.UserId,
				QuizId:     resp.QuizId,
				QuizResult: resp.QuizResult,
			})
		}
	}()

	<-done
	return nil
}
