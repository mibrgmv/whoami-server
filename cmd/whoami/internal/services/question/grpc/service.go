package grpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"whoami-server/cmd/whoami/internal/models"
	"whoami-server/cmd/whoami/internal/services/question"
	pbHistory "whoami-server/protogen/golang/history"
	pb "whoami-server/protogen/golang/question"
)

type QuestionService struct {
	service       *question.Service
	historyClient pbHistory.QuizCompletionHistoryServiceClient
	conn          *grpc.ClientConn
	pb.UnimplementedQuestionServiceServer
}

func NewService(service *question.Service, historyServiceAddr string) (*QuestionService, error) {
	conn, err := grpc.NewClient(historyServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to history service: %w", err)
	}

	historyClient := pbHistory.NewQuizCompletionHistoryServiceClient(conn)

	return &QuestionService{
		service:       service,
		historyClient: historyClient,
	}, nil
}

func (s *QuestionService) Add(stream pb.QuestionService_AddServer) error {
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
			ID:             req.ID,
			QuizID:         req.QuizID,
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
		pbQuestion := models.ToProto(q)
		if err := stream.Send(pbQuestion); err != nil {
			return err
		}
	}

	return nil
}

func (s *QuestionService) GetAll(empty *pb.Empty, stream pb.QuestionService_GetAllServer) error {
	questions, err := s.service.Query(stream.Context(), question.Query{})
	if err != nil {
		return err
	}

	done := make(chan bool)

	go func() {
		defer close(done)
		for _, q := range questions {
			if err := stream.Send(models.ToProto(q)); err != nil {
				log.Fatalf("can not send %v", err)
			}
		}
	}()

	<-done
	return nil
}

func (s *QuestionService) GetByQuizID(request *pb.GetByQuizIDRequest, stream pb.QuestionService_GetByQuizIDServer) error {
	questions, err := s.service.GetByQuizID(stream.Context(), request.QuizID)
	if err != nil {
		return err
	}

	for _, q := range questions {
		if err := stream.Send(models.ToProto(q)); err != nil {
			return err
		}
	}

	return nil
}

func (s *QuestionService) EvaluateAnswers(ctx context.Context, request *pb.EvaluateAnswersRequest) (*pb.EvaluateAnswersResponse, error) {
	var answers []models.Answer
	for _, answer := range request.Answers {
		answers = append(answers, models.Answer{
			QuizID:     answer.QuizID,
			QuestionID: answer.QuestionID,
			Body:       answer.Body,
		})
	}

	quiz := models.Quiz{
		ID:      request.Quiz.ID,
		Title:   request.Quiz.Title,
		Results: request.Quiz.Results,
	}

	result, err := s.service.EvaluateAnswers(ctx, answers, quiz)
	if err != nil {
		return nil, err
	}

	// todo user ID
	historyError := s.addToQuizCompletionHistory(ctx, 1, quiz.ID, result)
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
			if err == io.EOF {
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
