package handlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nikitalystsev/BookSmart-services/core/models"
	"github.com/nikitalystsev/BookSmart-services/errs"
	jsonmodels "github.com/nikitalystsev/BookSmart-web-api/core/models"
	"net/http"
)

func (h *Handler) reserveBook(c *gin.Context) {
	var bookID uuid.UUID
	if err := c.BindJSON(&bookID); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	readerIDStr, _, err := getReaderData(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	readerID, err := uuid.Parse(readerIDStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	err = h.reservationService.Create(c.Request.Context(), readerID, bookID)
	if err != nil && errors.Is(err, errs.ErrReaderDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, err.Error())
		return
	}
	if err != nil && errors.Is(err, errs.ErrReaderHasExpiredBooks) {
		c.AbortWithStatusJSON(http.StatusConflict, err.Error())
		return
	}
	if err != nil && errors.Is(err, errs.ErrReservationsLimitExceeded) {
		c.AbortWithStatusJSON(http.StatusConflict, err.Error())
		return
	}
	if err != nil && errors.Is(err, errs.ErrLibCardDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, err.Error())
		return
	}
	if err != nil && errors.Is(err, errs.ErrLibCardIsInvalid) {
		c.AbortWithStatusJSON(http.StatusConflict, err.Error())
		return
	}
	if err != nil && errors.Is(err, errs.ErrBookDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, err.Error())
		return
	}
	if err != nil && errors.Is(err, errs.ErrBookNoCopiesNum) {
		c.AbortWithStatusJSON(http.StatusConflict, err.Error())
		return
	}
	if err != nil && errors.Is(err, errs.ErrUniqueBookNotReserved) {
		c.AbortWithStatusJSON(http.StatusConflict, err.Error())
		return
	}
	if err != nil && errors.Is(err, errs.ErrReservationAgeLimit) {
		c.AbortWithStatusJSON(http.StatusConflict, err.Error())
		return
	}

	if err != nil && errors.Is(err, errs.ErrReservationAlreadyExists) {
		c.AbortWithStatusJSON(http.StatusConflict, err.Error())
		return
	}

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusCreated)
}

func (h *Handler) updateReservation(c *gin.Context) {
	var reservationID uuid.UUID
	if err := c.BindJSON(&reservationID); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	var reservation *models.ReservationModel
	reservation, err := h.reservationService.GetByID(c.Request.Context(), reservationID)
	if err != nil && errors.Is(err, errs.ErrReservationDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	err = h.reservationService.Update(c.Request.Context(), reservation)
	if err != nil && errors.Is(err, errs.ErrReservationObjectIsNil) {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	if err != nil && errors.Is(err, errs.ErrLibCardDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, err.Error())
		return
	}
	if err != nil && errors.Is(err, errs.ErrLibCardIsInvalid) {
		c.AbortWithStatusJSON(http.StatusConflict, err.Error())
		return
	}
	if err != nil && errors.Is(err, errs.ErrReaderHasExpiredBooks) {
		c.AbortWithStatusJSON(http.StatusConflict, err.Error())
		return
	}
	if err != nil && errors.Is(err, errs.ErrReservationIsAlreadyClosed) {
		c.AbortWithStatusJSON(http.StatusConflict, err.Error())
		return
	}
	if err != nil && errors.Is(err, errs.ErrReservationIsAlreadyExpired) {
		c.AbortWithStatusJSON(http.StatusConflict, err.Error())
		return
	}
	if err != nil && errors.Is(err, errs.ErrReservationIsAlreadyExtended) {
		c.AbortWithStatusJSON(http.StatusConflict, err.Error())
		return
	}
	if err != nil && errors.Is(err, errs.ErrRareAndUniqueBookNotExtended) {
		c.AbortWithStatusJSON(http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) getReservationsByReaderID(c *gin.Context) {
	readerIDStr, _, err := getReaderData(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	readerID, err := uuid.Parse(readerIDStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	var reservations []*models.ReservationModel
	reservations, err = h.reservationService.GetAllReservationsByReaderID(c.Request.Context(), readerID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}
	if len(reservations) == 0 {
		c.AbortWithStatusJSON(http.StatusNotFound, "reservation not found")
		return
	}

	_reservations := make([]*jsonmodels.ReservationModel, len(reservations))
	for i, reservation := range reservations {
		_reservations[i] = h.convertToJSONReservationModel(reservation)
	}

	c.JSON(http.StatusOK, _reservations)
}

func (h *Handler) getReservationsByID(c *gin.Context) {
	_, _, err := getReaderData(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	reservationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	var reservations *models.ReservationModel
	reservations, err = h.reservationService.GetByID(c.Request.Context(), reservationID)
	if err != nil && errors.Is(err, errs.ErrReservationDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	_reservation := h.convertToJSONReservationModel(reservations)

	c.JSON(http.StatusOK, _reservation)
}

func (h *Handler) convertToJSONReservationModel(reservation *models.ReservationModel) *jsonmodels.ReservationModel {
	return &jsonmodels.ReservationModel{
		ID:         reservation.ID,
		ReaderID:   reservation.ReaderID,
		BookID:     reservation.BookID,
		IssueDate:  reservation.IssueDate,
		ReturnDate: reservation.ReturnDate,
		State:      reservation.State,
	}
}
