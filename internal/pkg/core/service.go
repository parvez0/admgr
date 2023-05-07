package core

import (
	"fmt"
	"github.com/kiran-anand14/admgr/internal/pkg/api"
	"github.com/kiran-anand14/admgr/internal/pkg/models"
	"github.com/kiran-anand14/admgr/internal/pkg/storage/mysql"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

// Service provides User adding operations.
type Service interface {
	CreateSlots(slots []*api.CreateSlotRequestBody) error
	PatchSlots(slots *api.CreateSlotRequestBody) error
	GetSlots(filters map[string]interface{}) ([]*api.GetSlotsResponse, error)
}

// Repository provides access to User repository.
type Repository interface {
	Create(interface{}) (int, error)
	UpdateSlots(slot []mysql.Slot) (int, error)
	SearchSlots(filters map[string]interface{}) ([]*mysql.Slot, error)
}

type service struct {
	log *logrus.Logger
	r   Repository
}

// NewService creates an adding service with the necessary dependencies
func NewService(r Repository, log *logrus.Logger) Service {
	s := service{
		log: log,
		r:   r,
	}
	return &s
}

func (s *service) CreateSlots(createReqBody []*api.CreateSlotRequestBody) error {
	// any validation can be done here
	var slotsToCreate []mysql.Slot
	for _, req := range createReqBody {
		startDate := time.Time(req.StartDate)
		endDate := time.Time(req.EndDate)
		if startDate.After(endDate) {
			return models.NewError(
				fmt.Sprintf("BadParameterValue: start_date[%s] should be less than or equal to end_date[%s]", startDate.Format(time.DateOnly), endDate.Format(time.DateOnly)),
				models.DecodeFailureError,
			)
		}
		slots, err := s.fetchSlotsFromReqBody(req)
		if err != nil {
			return err
		}
		slotsToCreate = append(slotsToCreate, slots...)
	}
	s.log.Debugf("CreateSlots:: Adding %v to Repository", slotsToCreate)
	_, er := s.r.Create(slotsToCreate)
	if er != nil {
		return er
	}
	return nil
}

func (s *service) PatchSlots(req *api.CreateSlotRequestBody) error {
	slots, err := s.fetchSlotsFromReqBody(req)
	if err != nil {
		return err
	}
	_, err = s.r.UpdateSlots(slots)
	return err
}

func (s *service) GetSlots(filters map[string]interface{}) ([]*api.GetSlotsResponse, error) {
	allSlots := make([]*api.GetSlotsResponse, 0)

	if uid, e := filters["uid"]; e {
		filters["booked_by"] = uid
		delete(filters, "uid")
	}

	startDate, err := time.Parse(time.DateOnly, filters["start_date"].(string))
	if err != nil {
		return nil, models.NewError(fmt.Sprintf("start_date: %s decode failed", startDate), models.DecodeFailureError)
	}
	endDate, err := time.Parse(time.DateOnly, filters["end_date"].(string))
	if err != nil {
		return nil, models.NewError(fmt.Sprintf("end_date: %s decode failed", startDate), models.DecodeFailureError)
	}
	delete(filters, "start_date")
	delete(filters, "end_date")

	// Channels for passing dates and receiving slots
	resCh := make(chan *api.GetSlotsResponse)
	errCh := make(chan error)

	// Start a goroutine for each date
	var wg sync.WaitGroup
	for date := startDate; date.Before(endDate) || date.Equal(endDate); date = date.AddDate(0, 0, 1) {
		wg.Add(1)
		go s.fetchSlot(filters, date, errCh, resCh, &wg)
	}

	go func() {
		wg.Wait()
		close(resCh)
		close(errCh)
	}()

	for slot := range resCh {
		allSlots = append(allSlots, slot)
	}

	if err := <-errCh; err != nil {
		return nil, err
	}

	return allSlots, nil
}

func (s *service) fetchSlot(filters map[string]interface{}, date time.Time, errCh chan error, resCh chan *api.GetSlotsResponse, wg *sync.WaitGroup) {

	defer wg.Done()

	filters["date"] = date.Format(time.DateOnly)
	slots, err := s.r.SearchSlots(filters)
	if err != nil {
		errCh <- err
		return
	}
	// Process fetched slots
	var fetchedSlots []api.SlotResponse
	slotStatus := models.SLOT_STATUS_CLOSED
	for _, slot := range slots {
		if filters["uid"] != "" && slot.BookedBy != nil {
			if *slot.BookedBy == filters["uid"] {
				*slot.BookedBy = "me"
			} else {
				*slot.BookedBy = "others"
			}
		}
		var bookedDate models.JSONDate
		if slot.BookedDate != nil {
			bookedDate = models.JSONDate(*slot.BookedDate)
		}
		apiSlot := api.SlotResponse{
			Position:   *slot.Position,
			Cost:       *slot.Cost,
			Status:     *slot.Status,
			BookedBy:   slot.BookedBy,
			BookedDate: bookedDate,
		}
		if apiSlot.Status == models.SLOT_STATUS_OPEN {
			slotStatus = models.SLOT_STATUS_OPEN
		}
		fetchedSlots = append(fetchedSlots, apiSlot)
	}
	slot := &api.GetSlotsResponse{
		Date:   date,
		Status: slotStatus,
		Slots:  &fetchedSlots,
	}
	resCh <- slot
}

func (s *service) fetchSlotsFromReqBody(req *api.CreateSlotRequestBody) ([]mysql.Slot, error) {
	var slots []mysql.Slot
	for date := time.Time(req.StartDate); date.Before(time.Time(req.EndDate)) || date.Equal(time.Time(req.EndDate)); date = date.AddDate(0, 0, 1) {
		if req.Position[0] > 1 {
			preSlots, err := s.r.SearchSlots(
				map[string]interface{}{
					"date":     date.Format(time.DateOnly),
					"position": req.Position[0] - 1,
				},
			)
			if err != nil || len(preSlots) == 0 {
				return nil, models.NewError(
					fmt.Sprintf("Invalid position given it should be a sequence, record with position '%d' doesn't exits", req.Position[0]-1),
					models.DecodeFailureError,
				)
			}
		}
		for pos := req.Position[0]; pos <= req.Position[1]; pos++ {
			slotDate, slotPos := date, pos
			slot := mysql.Slot{
				Date:     &slotDate,
				Position: &slotPos,
				Cost:     req.Cost,
				Status:   req.Status,
			}
			slots = append(slots, slot)
		}
	}
	return slots, nil
}
