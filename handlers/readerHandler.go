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

func (h *Handler) signUp(c *gin.Context) {
	var inp dto.ReaderSignUpDTO
	if err := c.BindJSON(&inp); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	reader := models.ReaderModel{
		ID:          uuid.New(),
		Fio:         inp.Fio,
		PhoneNumber: inp.PhoneNumber,
		Age:         inp.Age,
		Password:    inp.Password,
		Role:        "Reader",
	}

	err := h.readerService.SignUp(c.Request.Context(), &reader)
	if err != nil && errors.Is(err, errs.ErrReaderObjectIsNil) {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	if err != nil && errors.Is(err, errs.ErrReaderAlreadyExist) {
		c.AbortWithStatusJSON(http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusCreated)
}

func (h *Handler) signIn(c *gin.Context) {
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

	c.JSON(http.StatusOK, dto.ReaderTokensDTO{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
		ExpiredAt:    time.Now().Add(h.accessTokenTTL).UnixMilli(),
	})
}

func (h *Handler) refresh(c *gin.Context) {
	var inp string
	if err := c.BindJSON(&inp); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, "invalid input body")
		return
	}

	var (
		res *models.Tokens
		err error
	)

	res, err = h.readerService.RefreshTokens(c.Request.Context(), inp)
	if err != nil && errors.Is(err, errs.ErrReaderDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, dto.ReaderTokensDTO{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
		ExpiredAt:    time.Now().Add(h.accessTokenTTL).UnixMilli(),
	})
}

func (h *Handler) getReaderByPhoneNumber(c *gin.Context) {
	phoneNumber := c.Query("phone_number")

	reader, err := h.readerService.GetByPhoneNumber(c.Request.Context(), phoneNumber)
	if err != nil && errors.Is(err, errs.ErrReaderDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	_reader := h.convertToJSONReaderModel(reader)

	c.JSON(http.StatusOK, _reader)
}

func (h *Handler) convertToJSONReaderModel(reader *models.ReaderModel) *jsonmodels.ReaderModel {
	return &jsonmodels.ReaderModel{
		ID:          reader.ID,
		Fio:         reader.Fio,
		PhoneNumber: reader.PhoneNumber,
		Age:         reader.Age,
		Password:    reader.Password,
		Role:        reader.Role,
	}
}
