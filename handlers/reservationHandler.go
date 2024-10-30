package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nikitalystsev/BookSmart-services/core/dto"
	"github.com/nikitalystsev/BookSmart-services/core/models"
	"github.com/nikitalystsev/BookSmart-services/errs"
	jsondto "github.com/nikitalystsev/BookSmart-web-api/core/dto"
	jsonmodels "github.com/nikitalystsev/BookSmart-web-api/core/models"
	"net/http"
)

// @Summary Метод бронирования книги
// @Security ApiKeyAuth
// @Tags reader_reservations
// @ID reserveBook
// @Accept  json
// @Produce  json
// @Param id path string true "Идентификатор читателя"
// @Param input body dto.ReservationInputDTO true "Идентификатор книги"
// @Success 201 "Успешное бронирование книги"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 401 {object} dto.ErrorResponse "Неавторизованный пользователь"
// @Failure 403 {object} dto.ErrorResponse "доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Нет читательского билета или книги"
// @Failure 409 {object} dto.ErrorResponse "Бронирование невозможно из-за нарушения некоторых условий"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/readers/{id}/reservations [post]
func (h *Handler) reserveBook(c *gin.Context) {
	readerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	isReader, err := isReaderID(c, readerID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if !isReader {
		c.AbortWithStatusJSON(http.StatusForbidden, jsondto.ErrorResponse{ErrorMsg: "access denied"})
		return
	}

	var inp dto.ReservationInputDTO
	if err = c.BindJSON(&inp); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	err = h.reservationService.Create(c.Request.Context(), readerID, inp.BookID)
	if err != nil && errors.Is(err, errs.ErrReaderDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if err != nil && errors.Is(err, errs.ErrReaderHasExpiredBooks) {
		c.AbortWithStatusJSON(http.StatusConflict, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if err != nil && errors.Is(err, errs.ErrReservationsLimitExceeded) {
		c.AbortWithStatusJSON(http.StatusConflict, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if err != nil && errors.Is(err, errs.ErrLibCardDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if err != nil && errors.Is(err, errs.ErrLibCardIsInvalid) {
		c.AbortWithStatusJSON(http.StatusConflict, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if err != nil && errors.Is(err, errs.ErrBookDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if err != nil && errors.Is(err, errs.ErrBookNoCopiesNum) {
		c.AbortWithStatusJSON(http.StatusConflict, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if err != nil && errors.Is(err, errs.ErrUniqueBookNotReserved) {
		c.AbortWithStatusJSON(http.StatusConflict, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if err != nil && errors.Is(err, errs.ErrReservationAgeLimit) {
		c.AbortWithStatusJSON(http.StatusConflict, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	if err != nil && errors.Is(err, errs.ErrReservationAlreadyExists) {
		c.AbortWithStatusJSON(http.StatusConflict, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}

// @Summary Метод обновления брони читателя по идентификатору
// @Security ApiKeyAuth
// @Tags reader_reservations
// @ID updateReservation
// @Accept  json
// @Produce  json
// @Param id path string true "Идентификатор читателя"
// @Param reservation_id path string true "Идентификатор брони"
// @Param input body dto.ReservationExtentionPeriodDaysInputDTO true "Срок продления брони"
// @Success 200 "Успешное продление брони"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 401 {object} dto.ErrorResponse "Неавторизованный пользователь"
// @Failure 404 {object} dto.ErrorResponse "Бронь не найдена"
// @Failure 409 {object} dto.ErrorResponse "Нарушение каких либо условий для успешного продления брони"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/readers/{id}/reservations/{reservation_id} [patch]
func (h *Handler) updateReservation(c *gin.Context) {
	readerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	reservationID, err := uuid.Parse(c.Param("reservation_id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	isReader, err := isReaderID(c, readerID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if !isReader {
		c.AbortWithStatusJSON(http.StatusForbidden, jsondto.ErrorResponse{ErrorMsg: "access denied"})
		return
	}

	var inp dto.ReservationExtentionPeriodDaysInputDTO
	if err = c.BindJSON(&inp); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	reservation, err := h.reservationService.GetByID(c.Request.Context(), reservationID)
	if err != nil && errors.Is(err, errs.ErrReservationDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	err = h.reservationService.Update(c.Request.Context(), reservation, inp.ExtentionPeriodDays)
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

// @Summary Метод получения всех броней читателя
// @Security ApiKeyAuth
// @Tags reader_reservations
// @ID getReservationsByReaderID
// @Accept  json
// @Produce  json
// @Param id path string true "Идентификатор читателя"
// @Success 200 {array} dto.ReservationOutputDTO "Успешное получение броней"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 401 {object} dto.ErrorResponse "Неавторизованный пользователь"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Нет броней"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/readers/{id}/reservations [get]
func (h *Handler) getReservationsByReaderID(c *gin.Context) {
	readerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	isReader, err := isReaderID(c, readerID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if !isReader {
		c.AbortWithStatusJSON(http.StatusForbidden, jsondto.ErrorResponse{ErrorMsg: "access denied"})
		return
	}

	var reservations []*models.ReservationModel
	reservations, err = h.reservationService.GetAllReservationsByReaderID(c.Request.Context(), readerID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if len(reservations) == 0 {
		c.AbortWithStatusJSON(http.StatusNotFound, jsondto.ErrorResponse{ErrorMsg: "reservation not found"})
		return
	}

	reservationOutputDTOs, err := h.copyReservationModelsToReservationOutputDTOs(c.Request.Context(), reservations)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	c.JSON(http.StatusOK, reservationOutputDTOs)
}

// @Summary Метод получения брони читателя по идентификатору
// @Security ApiKeyAuth
// @Tags reader_reservations
// @ID getReservationByID
// @Accept  json
// @Produce  json
// @Param id path string true "Идентификатор читателя"
// @Param reservation_id path string true "Идентификатор брони"
// @Success 200 {object} dto.ReservationOutputDTO "Успешное получение брони"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 401 {object} dto.ErrorResponse "Неавторизованный пользователь"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse " Бронь не найдена"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/readers/{id}/reservations/{reservation_id} [get]
func (h *Handler) getReservationByID(c *gin.Context) {
	readerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	reservationID, err := uuid.Parse(c.Param("reservation_id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	isReader, err := isReaderID(c, readerID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if !isReader {
		c.AbortWithStatusJSON(http.StatusForbidden, jsondto.ErrorResponse{ErrorMsg: "access denied"})
		return
	}

	reservations, err := h.reservationService.GetByID(c.Request.Context(), reservationID)
	if err != nil && errors.Is(err, errs.ErrReservationDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	reservationOutputDTO, err := h.copyReservationModelToReservationOutputDTO(c.Request.Context(), reservations)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	c.JSON(http.StatusOK, reservationOutputDTO)
}

func (h *Handler) copyReservationModelToReservationOutputDTO(ctx context.Context, reservation *models.ReservationModel) (*dto.ReservationOutputDTO, error) {
	book, err := h.bookService.GetByID(ctx, reservation.BookID)
	if err != nil && !errors.Is(err, errs.ErrBookDoesNotExists) {
		return nil, err
	}
	if book == nil {
		return nil, errs.ErrBookDoesNotExists
	}

	reservationOutputDTO := dto.ReservationOutputDTO{
		ID:                 reservation.ID,
		BookTitleAndAuthor: fmt.Sprintf("%s; %s", book.Title, book.Author),
		IssueDate:          reservation.IssueDate,
		ReturnDate:         reservation.ReturnDate,
		State:              reservation.State,
	}

	return &reservationOutputDTO, nil
}

func (h *Handler) copyReservationModelsToReservationOutputDTOs(ctx context.Context, reservations []*models.ReservationModel) ([]*dto.ReservationOutputDTO, error) {
	reservationOutputDTOs := make([]*dto.ReservationOutputDTO, len(reservations))

	for i, reservation := range reservations {
		outputDTO, err := h.copyReservationModelToReservationOutputDTO(ctx, reservation)
		if err != nil {
			return nil, err
		}
		reservationOutputDTOs[i] = outputDTO
	}

	return reservationOutputDTOs, nil
}

func (h *Handler) convertArrayToJSONReservationModels(reservations []*models.ReservationModel) []*jsonmodels.JSONReservationModel {
	jsonReservations := make([]*jsonmodels.JSONReservationModel, len(reservations))
	for i, reservation := range reservations {
		jsonReservations[i] = h.convertToJSONReservationModel(reservation)
	}
	return jsonReservations
}

func (h *Handler) convertToJSONReservationModel(reservation *models.ReservationModel) *jsonmodels.JSONReservationModel {
	return &jsonmodels.JSONReservationModel{
		ID:         reservation.ID,
		ReaderID:   reservation.ReaderID,
		BookID:     reservation.BookID,
		IssueDate:  reservation.IssueDate,
		ReturnDate: reservation.ReturnDate,
		State:      reservation.State,
	}
}
