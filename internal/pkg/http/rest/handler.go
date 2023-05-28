package rest

import (
	"encoding/json"
	"fmt"
	"github.com/kiran-anand14/admgr/internal/pkg/api"
	"github.com/kiran-anand14/admgr/internal/pkg/models"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/kiran-anand14/admgr/internal/pkg/core"
)

var (
	logger  *logrus.Logger
	service core.Service
)

func Handler(log *logrus.Logger, s core.Service, writer io.Writer) (*gin.Engine, error) {
	logger = log
	service = s

	r := gin.Default()
	gin.DefaultWriter = writer
	// Add all HTTP routes here.
	r.POST("/adslots", createSlotHandler)
	r.GET("/adslots", getSlotHandler)
	r.PATCH("/adslots", updateSlotHandler)
	r.DELETE("/adslots", deleteSlotHandler)
	r.PATCH("/adslots/reserve", reserveSlotHandler)
	r.GET("/health-check", healthCheck)

	return r, nil
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, nil)
}

func createSlotHandler(c *gin.Context) {
	var requestBody []*api.CreateSlotRequestBody
	err := json.NewDecoder(c.Request.Body).Decode(&requestBody)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "ParsingError: Invalid request body provided"})
		return
	}
	for i, slotRequest := range requestBody {
		if err := api.ValidateWithTags(slotRequest, fmt.Sprintf(".[%d].", i)); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("BadRequest:: [Error: %s]", err.Error())})
			return
		}
	}
	er := service.CreateSlots(requestBody)
	if er != nil {
		httpCode, msg := getHttpCodeAndMessage(er)
		if msg == "" {
			msg = DefaultErrorMsg
		}
		c.JSON(httpCode, gin.H{"error": msg})
		return
	}

	c.Status(http.StatusCreated)
	return
}

func getSlotHandler(c *gin.Context) {
	reqParams, requiredParams := c.Request.URL.Query(), map[string]bool{"start_date": true, "end_date": true}
	params := make(map[string]string)
	for k, v := range reqParams {
		params[k] = strings.Join(v, "")
	}
	for k := range requiredParams {
		if v, e := params[k]; !e || v == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s is required", k)})
			return
		}
	}
	res, er := service.GetSlots(params)
	if er != nil {
		httpCode, msg := getHttpCodeAndMessage(er)
		if msg == "" {
			msg = DefaultErrorMsg
		}
		c.JSON(httpCode, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, res)
	return
}

func updateSlotHandler(c *gin.Context) {
	var requestBody []*api.CreateSlotRequestBody
	err := json.NewDecoder(c.Request.Body).Decode(&requestBody)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "ParsingError: Invalid request body provided"})
		return
	}
	for i, req := range requestBody {
		if time.Time(req.StartDate).IsZero() || time.Time(req.StartDate).IsZero() || len(req.Position) == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf(".[%d].start_date, .[%d].end_date or .[%d].position is missing in request body", i, i, i)})
			return
		}
	}
	affected, err := service.PatchSlots(requestBody)
	if err != nil {
		httpCode, erMsg := getHttpCodeAndMessage(err)
		c.AbortWithStatusJSON(httpCode, gin.H{"error": erMsg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Total %d records updated", affected)})
}

func deleteSlotHandler(c *gin.Context) {
	var requestBody []*api.DeleteSlotRequestBody
	err := json.NewDecoder(c.Request.Body).Decode(&requestBody)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "ParsingError: Invalid request body provided"})
		return
	}
	for i, req := range requestBody {
		if time.Time(req.StartDate).IsZero() || time.Time(req.StartDate).IsZero() || len(req.Position) == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf(".[%d].start_date, .[%d].end_date or .[%d].position is missing in request body", i, i, i)})
			return
		}
	}
	err = service.DeleteSlots(requestBody)
	if err != nil {
		httpCode, erMsg := getHttpCodeAndMessage(err)
		c.AbortWithStatusJSON(httpCode, gin.H{"error": erMsg})
		return
	}
	c.Status(http.StatusOK)
}

func reserveSlotHandler(c *gin.Context) {
	var requestBody []*api.ReserveSlotRequestBody
	err := json.NewDecoder(c.Request.Body).Decode(&requestBody)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "ParsingError: Invalid request body provided"})
		return
	}
	for i, slotRequest := range requestBody {
		if err := api.ValidateWithTags(slotRequest, fmt.Sprintf(".[%d].", i)); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("BadRequest:: [Error: %s]", err.Error())})
			return
		}
	}
	uid := c.Query("uid")
	if uid == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Query param 'uid' cannot be empty"})
		return
	}
	err = service.ReserveSlots(requestBody, uid)
	if err != nil {
		httpCode, erMsg := getHttpCodeAndMessage(err)
		c.AbortWithStatusJSON(httpCode, gin.H{"error": erMsg})
		return
	}
	c.JSON(http.StatusOK, "ok")
}

func getHttpCodeAndMessage(err error) (int, string) {
	httpCode := http.StatusInternalServerError
	if _, ok := err.(*models.Error); !ok {
		return httpCode, err.Error()
	}
	er := err.(*models.Error)
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
	case models.DependentServiceRequestFailed:
		httpCode = http.StatusFailedDependency
	default:
		logger.Errorf("Invalid Error Type, so returning 500")
		httpCode = http.StatusInternalServerError
	}

	return httpCode, er.Message
}
