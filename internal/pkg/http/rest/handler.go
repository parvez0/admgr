package rest

import (
	"net/http"

	"github.com/kiran-anand14/admgr/internal/pkg/api"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/sirupsen/logrus"

	"github.com/kiran-anand14/admgr/internal/pkg/core"
	"github.com/kiran-anand14/admgr/internal/pkg/models"
)

var (
	logger  *logrus.Logger
	service core.Service
)

func Handler(log *logrus.Logger, s core.Service) (*gin.Engine, error) {
	logger = log
	service = s

	r := gin.Default()

	// Add all HTTP routes here.
	r.POST("/prod", createProd)

	return r, nil
}

func createProd(c *gin.Context) {
	var prod api.Product

	if err := c.ShouldBindBodyWith(&prod, binding.JSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": DecodeFailureErrorMsg})
		return
	}

	msgBody, er := service.CreateProd(&prod)
	if er != nil {
		httpCode, msg := getHttpCodeAndMessage(er)
		if msg == "" {
			msg = DefaultErrorMsg
		}
		c.JSON(httpCode, gin.H{"error": msg})
		return
	}

	c.JSON(http.StatusCreated, msgBody)

	return
}

func getHttpCodeAndMessage(er *models.Error) (int, string) {
	var httpCode int

	switch er.Type {
	case models.DecodeFailureError:
		httpCode = http.StatusBadRequest
	case models.InternalProcessingError:
		httpCode = http.StatusInternalServerError
	case models.DuplicateResourceCreationError:
		httpCode = http.StatusConflict
	case models.ResourceNotFoundError:
		httpCode = http.StatusNotFound
	case models.ActionForbidden:
		httpCode = http.StatusForbidden
		if er.Message == "" {
			er.Message = ActionForbiddenMsg
		}
	case models.DetailedResourceInfoNotFound:
		httpCode = http.StatusNotFound
	default:
		logger.Errorf("Invalid Error Type, so returning 500")
		httpCode = http.StatusInternalServerError
	}

	return httpCode, er.Message
}
