package rest

import (
	"encoding/json"
	"fmt"
	"github.com/kiran-anand14/admgr/internal/pkg/api"
	"github.com/kiran-anand14/admgr/internal/pkg/models"
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

func Handler(log *logrus.Logger, s core.Service) (*gin.Engine, error) {
	logger = log
	service = s

	r := gin.Default()

	// Add all HTTP routes here.
	r.POST("/adslots", createSlotHandler)
	r.GET("/adslots", getSlotHandler)
	r.PATCH("/adslots", updateSlotHandler)
	r.PATCH("/adslots/reserve", reserveSlotHandler)

	return r, nil
}

func createSlotHandler(c *gin.Context) {
	var requestBody []*api.CreateSlotRequestBody
	jsonData, err := c.GetRawData()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}
	err = json.Unmarshal(jsonData, &requestBody)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": DecodeFailureErrorMsg + err.Error()})
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
		httpCode, msg := getHttpCodeAndMessage(er.(*models.Error))
		if msg == "" {
			msg = DefaultErrorMsg
		}
		c.JSON(httpCode, gin.H{"error": msg})
		return
	}

	c.JSON(http.StatusCreated, "ok")
	return
}

func getSlotHandler(c *gin.Context) {
	reqParams, requiredParams := c.Request.URL.Query(), map[string]bool{"start_date": true, "end_date": true}
	params := make(map[string]interface{})
	for k, v := range reqParams {
		if requiredParams[k] && len(v) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s is required", k)})
			return
		}
		params[k] = strings.Join(v, "")
	}
	res, er := service.GetSlots(params)
	if er != nil {
		httpCode, msg := getHttpCodeAndMessage(er.(*models.Error))
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
	jsonData, err := c.GetRawData()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}
	err = json.Unmarshal(jsonData, &requestBody)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": DecodeFailureErrorMsg + err.Error()})
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
		httpCode, erMsg := getHttpCodeAndMessage(err.(*models.Error))
		c.AbortWithStatusJSON(httpCode, gin.H{"error": erMsg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Total %d records updated", affected)})
}

func reserveSlotHandler(c *gin.Context) {}

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
