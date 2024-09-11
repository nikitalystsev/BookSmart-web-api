package handlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nikitalystsev/BookSmart-services/core/dto"
	"github.com/nikitalystsev/BookSmart-services/core/models"
	"github.com/nikitalystsev/BookSmart-services/errs"
	jsonmodels "github.com/nikitalystsev/BookSmart-web-api/core/models"
	"net/http"
	"time"
)

func (h *Handler) signInAsAdmin(c *gin.Context) {
	var inp dto.ReaderSignInDTO
	if err := c.BindJSON(&inp); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	var (
		res *models.Tokens
		err error
	)

	res, err = h.readerService.SignIn(c.Request.Context(), inp.PhoneNumber, inp.Password)
	if err != nil && errors.Is(err, errs.ErrReaderDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, err.Error())
		return
	}
	if err != nil && errors.Is(err, errors.New("wrong password")) {
		c.AbortWithStatusJSON(http.StatusConflict, err.Error())
		return
	}
	if err != nil && errors.Is(err, errs.ErrReaderObjectIsNil) {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	_, role, err := h.tokenManager.Parse(res.AccessToken)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	if role == "Reader" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, "you are not authorized to perform this action")
		return
	}

	c.JSON(http.StatusOK, dto.ReaderTokensDTO{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
		ExpiredAt:    time.Now().Add(h.accessTokenTTL).UnixMilli(),
	})
}

func (h *Handler) deleteBook(c *gin.Context) {
	_, readerRole, err := getReaderData(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	if readerRole == "Reader" {
		c.AbortWithStatusJSON(http.StatusForbidden, "reader not delete book")
		return
	}

	bookID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	err = h.bookService.Delete(c.Request.Context(), bookID)
	if err != nil && errors.Is(err, errs.ErrBookObjectIsNil) {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	if err != nil && errors.Is(err, errs.ErrBookDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) addNewBook(c *gin.Context) {
	_, readerRole, err := getReaderData(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	if readerRole == "Reader" {
		c.AbortWithStatusJSON(http.StatusForbidden, "reader not delete book")
		return
	}

	var newBook dto.BookDTO
	if err = c.BindJSON(&newBook); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	book := &models.BookModel{
		ID:             uuid.New(),
		Title:          newBook.Title,
		Author:         newBook.Author,
		Publisher:      newBook.Publisher,
		CopiesNumber:   newBook.CopiesNumber,
		Rarity:         newBook.Rarity,
		Genre:          newBook.Genre,
		PublishingYear: newBook.PublishingYear,
		Language:       newBook.Language,
		AgeLimit:       newBook.AgeLimit,
	}

	err = h.bookService.Create(c.Request.Context(), (*models.BookModel)(book))
	if err != nil && errors.Is(err, errs.ErrBookObjectIsNil) {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusCreated)
}

func (h *Handler) getReservationsByBookID(c *gin.Context) {
	bookID, err := uuid.Parse(c.Query("book_id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	reservations, err := h.reservationService.GetByBookID(c.Request.Context(), bookID)
	if err != nil && errors.Is(err, errs.ErrReservationDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	_reservations := make([]*jsonmodels.ReservationModel, len(reservations))
	for i, reservation := range reservations {
		_reservations[i] = h.convertToJSONReservationModel(reservation)
	}

	c.JSON(http.StatusOK, _reservations)
}
