package core

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/kiran-anand14/admgr/internal/pkg/accounting"
	"github.com/kiran-anand14/admgr/internal/pkg/api"
	"github.com/kiran-anand14/admgr/internal/pkg/models"
	"github.com/kiran-anand14/admgr/internal/pkg/storage/mysql"
	"github.com/sirupsen/logrus"
	"time"
)

// Service provides User adding operations.
type Service interface {
	CreateSlots(slots []*api.CreateSlotRequestBody) error
	PatchSlots(slots []*api.CreateSlotRequestBody) (int, error)
	GetSlots(filters map[string]string) ([]*api.GetSlotsResponse, error)
	ReserveSlots(request []*api.ReserveSlotRequestBody, uid string) error
	DeleteSlots(reqBody []*api.DeleteSlotRequestBody) error
}

// Repository provides access to User repository.
type Repository interface {
	Create(interface{}) (int, error)
	UpdateSlots(slots []*mysql.Slot) (int, error)
	SearchSlotsInRange(options *mysql.GetOptions) ([]*mysql.Slot, error)
	SearchSlotsByStatus(options *mysql.GetOptions) ([]*mysql.Slot, error)
	UpdateSlotsStatus(slots []*mysql.Slot, lastStatus, newStatus string) error
	Delete(records interface{}) (int, error)
}

type service struct {
	log *logrus.Logger
	acc accounting.AccountingService
	rep Repository
}

// NewService creates an adding service with the necessary dependencies
func NewService(r Repository, a accounting.AccountingService, log *logrus.Logger) Service {
	s := service{
		log: log,
		rep: r,
		acc: a,
	}
	retry := 1
	for {
		if err := s.revertFailedReservations(); err != nil {
			s.log.Errorf("CoreServiceInitialization: Failed to revert reservations [Error: %s, Retrying: %d]", err, retry)
			if retry > 10 {
				panic("ReachedMaximumRetires: Failed CoreServiceInitialization")
			}
			time.Sleep(time.Second * 10)
			retry++
			continue
		}
		break
	}
	return &s
}

func (s *service) revertFailedReservations() error {
	s.log.Info("Finding all slots on hold status")
	opts := mysql.GetOptions{
		Status:             models.SlotStatusHold,
		PreloadTransaction: true,
	}
	slots, err := s.rep.SearchSlotsByStatus(&opts)
	if err != nil {
		return err
	}
	s.log.Infof("Total %d slots found to be on hold status", len(slots))
	var (
		slotsToUpdate []*mysql.Slot
		txnIds        []string
		slotMap       = make(map[string]*mysql.Slot)
	)
	for _, slot := range slots {
		if slot.Transaction == nil {
			s.log.Warnf("Transaction not found for [Slot: %s, Status: %v], Reverting status to '%s'", slotIdFromSlot([]*mysql.Slot{slot}), slot.Status, models.SlotStatusOpen)
			slot.Status = models.PtrString(models.SlotStatusOpen)
			slotsToUpdate = append(slotsToUpdate, slot)
			continue
		}
		txnIds = append(txnIds, slot.Transaction.Txnid)
		slotMap[slot.Transaction.Txnid] = slot
	}
	resp, err := s.acc.Status(txnIds)
	if err != nil {
		s.log.Errorf("Error while communicating with accounting service: %s", err.Error())
		return err
	}
	for _, txn := range resp {
		slot, ok := slotMap[txn.Txnid]
		if !ok {
			s.log.Errorf("Id not found in slot map: %s", txn.Txnid)
			continue
		}
		slot.BookedDate = models.PtrDate(txn.Created)
		slot.BookedBy = models.PtrString(txn.UID)
		slot.Status = models.PtrString(models.SlotStatusBooked)
		slotsToUpdate = append(slotsToUpdate, slot)
		delete(slotMap, txn.Txnid)
	}
	for _, slot := range slotMap {
		slot.Status = models.PtrString(models.SlotStatusOpen)
		slotsToUpdate = append(slotsToUpdate, slot)
	}
	s.log.Infof("Transaction was successful. Changing status to booked for %s", slotIdFromSlot(slotsToUpdate))
	updateCount, err := s.rep.UpdateSlots(slotsToUpdate)
	if err != nil {
		s.log.Errorf("Reverting changes failed [Error: %s]", err.Error())
		return err
	}
	s.log.Infof("Total %d slots status updated", updateCount)
	return nil
}

func slotIdFromSlot(slots []*mysql.Slot) string {
	res := ""
	for _, s := range slots {
		res += fmt.Sprintf("[Slots: %s-%d, Status: %s],", s.Date.Format(time.DateOnly), *s.Position, *s.Status)
	}
	return res
}

func (s *service) CreateSlots(createReqBody []*api.CreateSlotRequestBody) error {
	// any validation can be done here
	var slotsToCreate []*mysql.Slot
	for _, req := range createReqBody {
		startDate := time.Time(req.StartDate)
		endDate := time.Time(req.EndDate)
		if startDate.After(endDate) {
			return models.NewError(
				fmt.Sprintf("BadParameterValue: start_date[%s] should be less than or equal to end_date[%s]", startDate.Format(time.DateOnly), endDate.Format(time.DateOnly)),
				models.DecodeFailureError,
			)
		}
		slots, err := s.fetchSlotsFromReqBody(req, models.PtrString(models.SlotStatusOpen))
		if err != nil {
			return err
		}
		slotsToCreate = append(slotsToCreate, slots...)
	}
	s.log.Debugf("CreateSlots:: Adding %v to Repository", slotsToCreate)
	_, er := s.rep.Create(slotsToCreate)
	if er != nil {
		return er
	}
	return nil
}

func (s *service) PatchSlots(patchReqBody []*api.CreateSlotRequestBody) (int, error) {
	var slotsToUpdate []*mysql.Slot
	for _, req := range patchReqBody {
		startDate := time.Time(req.StartDate)
		endDate := time.Time(req.EndDate)
		if startDate.After(endDate) {
			return 0, models.NewError(
				fmt.Sprintf("BadParameterValue: start_date[%s] should be less than or equal to end_date[%s]", startDate.Format(time.DateOnly), endDate.Format(time.DateOnly)),
				models.DecodeFailureError,
			)
		}
		slots, err := s.fetchSlotsFromReqBody(req, nil)
		if err != nil {
			return 0, err
		}
		slotsToUpdate = append(slotsToUpdate, slots...)
	}
	s.log.Debugf("CreateSlots:: Adding %v to Repository", slotsToUpdate)
	return s.rep.UpdateSlots(slotsToUpdate)
}

func (s *service) GetSlots(filters map[string]string) ([]*api.GetSlotsResponse, error) {
	startDate, err := time.Parse(time.DateOnly, filters["start_date"])
	if err != nil {
		return nil, models.NewError(fmt.Sprintf("start_date: %s decode failed", startDate), models.DecodeFailureError)
	}
	endDate, err := time.Parse(time.DateOnly, filters["end_date"])
	if err != nil {
		return nil, models.NewError(fmt.Sprintf("end_date: %s decode failed", startDate), models.DecodeFailureError)
	}
	if startDate.After(endDate) {
		return nil, models.NewError(fmt.Sprintf("start_date[%s] cannot be greater than end_date[%s]", startDate.Format(time.DateOnly), endDate.Format(time.DateOnly)), models.DecodeFailureError)
	}

	position, _ := filters["position"]
	status, _ := filters["status"]
	uid, _ := filters["uid"]
	getOptions := &mysql.GetOptions{
		StartDate:     startDate,
		EndDate:       endDate,
		PositionStart: position,
		PositionEnd:   position,
		Status:        status,
		Uid:           uid,
	}
	slots, err := s.rep.SearchSlotsInRange(getOptions)
	if err != nil {
		return nil, err
	}
	return ConvertSlotsToJSON(slots)
}

func ConvertSlotsToJSON(slots []*mysql.Slot) ([]*api.GetSlotsResponse, error) {
	groups := make(map[string]*api.GetSlotsResponse)
	for _, s := range slots {
		date := s.Date.Format(time.DateOnly)
		if _, ok := groups[date]; !ok {
			groups[date] = &api.GetSlotsResponse{
				Date:  date,
				Slots: make([]*api.SlotResponse, 0),
			}
		}

		slot := &api.SlotResponse{
			Position: *s.Position,
			Cost:     *s.Cost,
			Status:   *s.Status,
		}
		if s.BookedDate != nil {
			slot.BookedDate = models.JsonDatePtr(models.JsonDate(*s.BookedDate))
			slot.BookedBy = s.BookedBy
		}
		groups[date].Slots = append(groups[date].Slots, slot)
	}
	result := make([]*api.GetSlotsResponse, 0, len(groups))
	for _, g := range groups {
		result = append(result, g)
	}
	return result, nil
}

func (s *service) ReserveSlots(reserveRequest []*api.ReserveSlotRequestBody, uid string) (err error) {
	var (
		slots        []*mysql.Slot
		transactions []*mysql.Transaction
	)
	txnid, err := uuid.NewUUID()
	if err != nil {
		return models.NewError(
			"failed to create new transaction id",
			models.InternalProcessingError,
		)
	}

	// prepare slots and transactions
	for _, r := range reserveRequest {
		date := time.Time(r.Date)
		pos := models.Int32ToString(*r.Position)
		getOptions := &mysql.GetOptions{
			StartDate:     date,
			EndDate:       date,
			PositionStart: pos,
			PositionEnd:   pos,
			Status:        models.SlotStatusOpen,
			Uid:           "",
		}
		slot, err := s.rep.SearchSlotsInRange(getOptions)
		if err != nil || len(slot) == 0 {
			return models.NewError(
				fmt.Sprintf("Slot with [date: %s, position: %d] not open", models.DateToString(date), *r.Position),
				models.ActionForbidden,
			)
		}
		txn := &mysql.Transaction{
			Txnid:    txnid.String(),
			Date:     models.PtrDate(date),
			Position: r.Position,
		}
		slot[0].BookedBy = models.PtrString(uid)
		slot[0].BookedDate = models.PtrDate(time.Now())
		slot[0].Status = models.PtrString(models.SlotStatusBooked)
		slots = append(slots, slot[0])
		transactions = append(transactions, txn)
	}

	// create transactions
	if _, err = s.rep.Create(transactions); err != nil {
		if mErr, ok := err.(*models.Error); ok {
			if mErr.Type == models.DuplicateResourceCreationError {
				err = models.NewError(
					fmt.Sprintf("Cannot reserve the slots, either they are already booked or it's on hold"),
					models.ActionForbidden,
				)
			}
		}
		return err
	}
	defer func() {
		if ok := recover(); ok != nil || err != nil {
			s.log.Errorf("Encountered error while reserving slots [PanicError: %+v, Error: %v] reverting changes", ok, err)
			s.rep.Delete(transactions)
			if err == nil {
				err = models.NewError("Failed to reserve slots, internal server error", models.InternalProcessingError)
			}
		}
	}()

	// debit transaction
	if err = s.acc.Debit(slots, uid, txnid.String()); err != nil {
		s.log.Debugf("DebitTransactionFailed:: reverting changes to db with [Status: %s, Slots: %+v]", models.SlotStatusOpen, slots)
		return err
	}

	// retry update slots on error
	for i := 0; i < 3; i++ {
		d, dbErr := s.rep.UpdateSlots(slots)
		if dbErr == nil {
			s.log.Infof("Total %d slots reserved successfully", d)
			break
		}
		s.log.Errorf("UpdatedReservedData:: Failed to update slots [Retried: %d, Error: %s]", i+1, dbErr.Error())
	}
	return nil
}

func (s *service) DeleteSlots(deleteReqBody []*api.DeleteSlotRequestBody) error {
	for _, reqBody := range deleteReqBody {
		startDate := time.Time(reqBody.StartDate)
		endDate := time.Time(reqBody.EndDate)
		if startDate.After(endDate) {
			return models.NewError(
				fmt.Sprintf("start_date[%s] cannot be greater than end_date[%s]", models.DateToString(startDate), models.DateToString(endDate)),
				models.DecodeFailureError,
			)
		}
		getOptions := &mysql.GetOptions{
			StartDate:          startDate,
			EndDate:            endDate,
			PositionStart:      models.Int32ToString(reqBody.Position[0]),
			PositionEnd:        models.Int32ToString(reqBody.Position[1]),
			PreloadTransaction: true,
		}
		getOptions.Query = fmt.Sprintf("status = '%s' OR status = '%s'", models.SlotStatusBooked, models.SlotStatusHold)
		bookedSlots, err := s.rep.SearchSlotsInRange(getOptions)
		if err != nil {
			return err
		}
		if len(bookedSlots) > 0 {
			return models.NewError(
				fmt.Sprintf("Attempting to delete booked or on-hold slots [start_date: %s, end_date: %s]", models.DateToString(startDate), models.DateToString(endDate)),
				models.ActionForbidden,
			)
		}
		getOptions.Status = models.SlotStatusOpen
		getOptions.Query = ""
		openSlots, err := s.rep.SearchSlotsInRange(getOptions)
		if err != nil {
			return err
		}
		if len(openSlots) == 0 {
			return models.NewError(
				fmt.Sprintf("records not found with [start_date: %s, end_date: %s] status open", models.DateToString(startDate), models.DateToString(endDate)),
				models.DetailedResourceInfoNotFound,
			)
		}
		getOptions.PositionStart = models.Int32ToString(reqBody.Position[1] + 1)
		getOptions.PositionEnd = models.Int32ToString(reqBody.Position[1] + 1)
		outOfSequenceSlots, err := s.rep.SearchSlotsInRange(getOptions)
		if err != nil {
			return err
		}
		if len(outOfSequenceSlots) > 0 {
			return models.NewError(
				fmt.Sprintf("attempt to delete out of sequence records not authorized, DeletePosition -> [%d, %d] but Position -> %d exits",
					reqBody.Position[0], reqBody.Position[1], reqBody.Position[1]+1),
				models.ActionForbidden,
			)
		}
		s.log.Debugf("Deleting %d records: %+v", len(openSlots), openSlots)
		_, err = s.rep.Delete(openSlots)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *service) fetchSlotsFromReqBody(req *api.CreateSlotRequestBody, status *string) ([]*mysql.Slot, error) {
	var slots []*mysql.Slot
	for date := time.Time(req.StartDate); date.Before(time.Time(req.EndDate)) || date.Equal(time.Time(req.EndDate)); date = date.AddDate(0, 0, 1) {
		if req.Position[0] > 1 {
			pos := models.Int32ToString(req.Position[0] - 1)
			getOptions := &mysql.GetOptions{
				StartDate:     date,
				EndDate:       date,
				PositionStart: pos,
				PositionEnd:   pos,
				Status:        "",
				Uid:           "",
			}
			preSlots, err := s.rep.SearchSlotsInRange(getOptions)
			if err != nil || len(preSlots) == 0 {
				return nil, models.NewError(
					fmt.Sprintf("Invalid date[%s] or record with position '%d' doesn't exits", models.DateToString(date), req.Position[0]-1),
					models.DecodeFailureError,
				)
			}
		}
		for pos := req.Position[0]; pos <= req.Position[1]; pos++ {
			slotDate, slotPos := date, pos
			slot := &mysql.Slot{
				Date:     &slotDate,
				Position: &slotPos,
				Cost:     req.Cost,
			}
			if status != nil {
				slot.Status = models.PtrString(models.SlotStatusOpen)
			}
			slots = append(slots, slot)
		}
	}
	return slots, nil
}
