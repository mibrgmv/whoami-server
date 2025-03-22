package quiz

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
	"whoami-server/cmd/gateway/internal/models"
	pb "whoami-server/protogen/golang/quiz"
)

type Handler struct {
	client pb.QuizServiceClient
	conn   *grpc.ClientConn
}

func NewHandler(quizServiceAddr string) (*Handler, error) {
	quizConn, err := grpc.NewClient(quizServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to quiz service: %w", err)
	}

	quizClient := pb.NewQuizServiceClient(quizConn)

	return &Handler{
		client: quizClient,
	}, nil
}

// Add godoc
// @Summary Add multiple quizzes
// @Description Add multiple quizzes to the database
// @Tags quizzes
// @Accept json
// @Produce json
// @Param quizzes body []models.Quiz true "Array of quiz objects to add"
// @Success 201 {array} models.Quiz
// @Failure 400
// @Failure 500
// @Router /quizzes/add [post]
func (h *Handler) Add(c *gin.Context) {
	var quizzes []models.Quiz
	if err := c.ShouldBindJSON(&quizzes); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stream, err := h.client.Add(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add quizzes"})
		return
	}

	var createdQuizzes []models.Quiz
	done := make(chan bool)

	go func() {
		for _, q := range quizzes {
			in := &pb.Quiz{Title: q.Title, Results: q.Results}
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

			quiz := models.Quiz{ID: resp.ID, Title: resp.Title, Results: resp.Results}
			createdQuizzes = append(createdQuizzes, quiz)
		}
	}()

	<-done
	c.JSON(http.StatusCreated, createdQuizzes)
}

// All godoc
// @Summary Get all quizzes
// @Description Retrieve all quizzes.
// @Tags quizzes
// @Produce json
// @Success 200 {array} models.Quiz
// @Failure 400
// @Failure 500
// @Router /quizzes [get]
func (h *Handler) All(c *gin.Context) {
	stream, err := h.client.GetAll(c.Request.Context(), &pb.Empty{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch quizzes"})
		return
	}

	var quizzes []models.Quiz
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
			quizzes = append(quizzes, models.Quiz{ID: resp.ID, Title: resp.Title, Results: resp.Results})
		}
	}()

	<-done
	c.JSON(http.StatusOK, quizzes)
}

// GetByID godoc
// @Summary Get quiz by ID
// @Description Retrieve a quiz by its ID.
// @Tags quizzes
// @Produce json
// @Param id path int64 true "Quiz ID"
// @Success 200 {object} models.Quiz
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /quizzes/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	quizIDStr := c.Param("id")
	quizID, err := strconv.ParseInt(quizIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid quiz ID"})
		return
	}

	in := &pb.GetByIDRequest{ID: quizID}
	q, err := h.client.GetByID(c.Request.Context(), in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch quiz"})
		return
	}

	c.JSON(http.StatusOK, q)
}
