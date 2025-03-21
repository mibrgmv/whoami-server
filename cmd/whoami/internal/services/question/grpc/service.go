package grpc

import (
	"context"
	"whoami-server/cmd/whoami/internal/models"
	"whoami-server/cmd/whoami/internal/services/question"
	pb "whoami-server/protogen/golang/question"
)

type QuestionService struct {
	service *question.Service
	pb.UnimplementedQuestionServiceServer
}

func NewService(service *question.Service) *QuestionService {
	return &QuestionService{service: service}
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

	for _, q := range questions {
		if err := stream.Send(models.ToProto(q)); err != nil {
			return err
		}
	}

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

	return &pb.EvaluateAnswersResponse{
		Result:      result,
		Description: "", // todo
	}, nil
}
