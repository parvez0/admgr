package main

import (
	"fmt"
	"github.com/google/uuid"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type AccountingRequestBody struct {
	Source   string             `json:"source"`
	Uid      string             `json:"uid"`
	Amount   float64            `json:"amount"`
	Txnid    string             `json:"txnid"`
	Metadata AccountingMetadata `json:"metadata"`
}

type AccountingMetadata struct {
	Slots []AccountingMetadataSlot `json:"slots"`
}

type AccountingMetadataSlot struct {
	Date     time.Time `json:"date"`
	Position int32     `json:"position"`
	Cost     float64   `json:"cost"`
}

type AccountingStatusResponse struct {
	Txnid    string             `json:"txnid,omitempty"`
	UID      string             `json:"uid,omitempty"`
	Created  time.Time          `json:"created,omitempty"`
	Metadata AccountingMetadata `json:"metadata,omitempty"`
}

type AccountingStatusRequest []string

func main() {
	router := gin.Default()

	router.GET("/health-check", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	router.POST("/debit", func(c *gin.Context) {
		var requestBody AccountingRequestBody

		// Bind request body to struct
		if err := c.BindJSON(&requestBody); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Failed to parse request body"})
			rand.New(rand.NewSource(time.Now().UnixNano()))
			time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
			return
		}
		fmt.Println(requestBody)
		// Do something with the request data...

		// Return "ok" response
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	router.POST("/status", func(c *gin.Context) {
		var requestBody AccountingStatusRequest
		if err := c.BindJSON(&requestBody); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Failed to parse request body"})
			return
		}
		fmt.Println(requestBody)
		// Do something with the request data...
		rand.New(rand.NewSource(time.Now().UnixNano()))
		status := []int{http.StatusNotFound}
		var res []*AccountingStatusResponse
		for _, req := range requestBody {
			// Return "ok" response
			sts := status[rand.Intn(len(status))]
			if sts != http.StatusOK {
				continue
			}
			fmt.Println("ID: ", res)
			data := &AccountingStatusResponse{
				Txnid:   req,
				UID:     uuid.New().String(),
				Created: time.Now(),
				Metadata: AccountingMetadata{
					Slots: []AccountingMetadataSlot{
						{
							Date:     time.Now(),
							Position: int32(rand.Intn(20)),
							Cost:     10.2,
						},
					},
				},
			}
			res = append(res, data)
		}
		c.JSON(http.StatusOK, res)
	})

	if err := router.Run(":10002"); err != nil {
		panic(err)
	}
}
