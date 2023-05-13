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

type AccountingDebitResponse struct {
	Txnid    string             `json:"txnid,omitempty"`
	UID      string             `json:"uid,omitempty"`
	Created  time.Time          `json:"created,omitempty"`
	Metadata AccountingMetadata `json:"metadata,omitempty"`
}

func main() {
	router := gin.Default()

	router.POST("/debit", func(c *gin.Context) {
		var requestBody AccountingRequestBody

		// Bind request body to struct
		if err := c.BindJSON(&requestBody); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Failed to parse request body"})
			fmt.Println(requestBody)
			rand.New(rand.NewSource(time.Now().UnixNano()))
			time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
			return
		}

		// Do something with the request data...

		// Return "ok" response
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	router.GET("/debit", func(c *gin.Context) {
		txnid, _ := c.GetQuery("txnid")
		if txnid == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Query params txnid is required"})
			return
		}
		// Do something with the request data...
		rand.New(rand.NewSource(time.Now().UnixNano()))
		// Return "ok" response
		status := []int{http.StatusOK, http.StatusNotFound}
		sts := status[rand.Intn(len(status))]
		if sts == http.StatusNotFound {
			c.AbortWithStatus(sts)
			return
		}
		data := &AccountingDebitResponse{
			Txnid:   txnid,
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
		c.JSON(sts, data)
	})

	if err := router.Run(":10002"); err != nil {
		panic(err)
	}
}

// POST
// status: reject if not open
// Error message reformat
// remove ok response
// Add required Info or Debug logs

// GET
// 	define a SQL range query for start and end date

// Patch
// status = [Close] delete on close -> on db level (tx) -> 1,2,3 update(date, position and status) -> error,4->booked,5
// don't allow booked
//
// Delete ->

// tx.DB.atomic -> transaction

// Req -> Create(Transactions) -> VerifyAndUpdate(Slots.Hold) -> Do -> Update(slots.Done)
