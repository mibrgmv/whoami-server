package question

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"whoami-server/cmd/gateway/internal/models"
	pb "whoami-server/protogen/golang/question"
)

type Handler struct {
	client pb.QuestionServiceClient
	conn   *grpc.ClientConn
}

func NewHandler(questionServiceAddr string) (*Handler, error) {
	conn, err := grpc.NewClient(questionServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to question service: %w", err)
	}

	client := pb.NewQuestionServiceClient(conn)

	return &Handler{
		client: client,
	}, nil
}

func (h *Handler) Close() error {
	// todo seems very useless since h.conn is never assigned
	if h.conn != nil {
		return h.conn.Close()
	}
	return nil
}

// Add godoc
// @Summary Add multiple questions
// @Description Add multiple questions to the database
// @Tags questions
// @Accept json
// @Produce json
// @Param questions body []models.Question true "Array of question objects to add"
// @Success 201 {array} models.Question
// @Failure 400
// @Failure 500
// @Router /questions/add [post]
func (h *Handler) Add(c *gin.Context) {
	var questions []models.Question
	if err := c.ShouldBindJSON(&questions); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stream, err := h.client.Add(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add questions"})
		return
	}

	var createdQuestions []models.Question
	done := make(chan bool)

	go func() {
		for _, q := range questions {
			in := &pb.Question{QuizID: q.QuizID, Body: q.Body, Options: q.Options}
			if err := stream.Send(in); err != nil {
				log.Fatalf("can not send %v", err)
			}
		}
		if err := stream.CloseSend(); err != nil {
			log.Fatalf("can not close %v", err)
		}
	}()

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
			createdQuestions = append(createdQuestions, models.Question{ID: resp.ID, QuizID: resp.QuizID, Body: resp.Body, Options: resp.Options})
		}
	}()

	<-done
	c.JSON(http.StatusCreated, createdQuestions)
}

// All godoc
// @Summary Query questions with parameters
// @Description Retrieve all quizzes.
// @Tags questions
// @Produce json
// @Success 200 {array} question
// @Failure 400
// @Failure 500
// @Router /questions [get]
func (h *Handler) All(c *gin.Context) {
	stream, err := h.client.GetAll(c.Request.Context(), &pb.Empty{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch questions"})
		return
	}

	var result []models.Question
	done := make(chan bool)

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

			// todo
			result = append(result, models.ToModel(&pb.Question{
				ID:             resp.ID,
				QuizID:         resp.QuizID,
				Body:           resp.Body,
				OptionsWeights: resp.OptionsWeights,
			}))
		}
	}()

	<-done
	c.JSON(http.StatusOK, result)
}

// GetByQuizID godoc
// @Summary Get questions by quiz ID
// @Description Retrieve questions associated with a specific quiz ID.
// @Tags questions
// @Produce json
// @Param id path int64 true "Quiz ID"
// @Success 200 {array} question
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /quizzes/{id}/questions [get]
func (h *Handler) GetByQuizID(c *gin.Context) {
	quizIDStr := c.Param("id")
	quizID, err := strconv.ParseInt(quizIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid quiz ID"})
		return
	}

	in := &pb.GetByQuizIDRequest{QuizID: quizID}
	stream, err := h.client.GetByQuizID(c.Request.Context(), in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch questions"})
		return
	}

	var result []models.Question
	done := make(chan bool)

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
			result = append(result, models.Question{ID: resp.ID, QuizID: resp.QuizID, Body: resp.Body, Options: resp.Options})
		}
	}()

	<-done
	c.JSON(http.StatusOK, result)
}

// EvaluateAnswers godoc
// @Summary Evaluate answers for a quiz
// @Description Evaluate submitted answers against a quiz and return result weights
// @Tags questions
// @Accept json
// @Produce json
// @Param id path int64 true "Quiz ID"
// @Param answers body []models.Answer true "Array of answer objects to evaluate"
// @Success 200 {object} map[string]interface{}
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /quizzes/{id}/evaluate [post]
func (h *Handler) EvaluateAnswers(c *gin.Context) {
	quizIDStr := c.Param("id")
	quizID, err := strconv.ParseInt(quizIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid quiz ID"})
		return
	}

	var answers []models.Answer
	if err := c.ShouldBindJSON(&answers); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid answers format", "error": err.Error()})
		return
	}

	if len(answers) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "No answers provided"})
		return
	}

	in := &pb.GetByIDRequest{ID: quizID}
	quiz, err := h.client.GetByID(c.Request.Context(), in)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Quiz not found"})
		return
	}

	in = &pb.EvaluateAnswersRequest{Answers: answers, Quiz: quiz}
	result, err := h.client.EvaluateAnswers(c.Request.Context(), in)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	response := gin.H{
		"result": result,
	}

	c.JSON(http.StatusOK, response)
}
