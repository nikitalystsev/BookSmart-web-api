package handlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nikitalystsev/BookSmart-services/core/dto"
	"github.com/nikitalystsev/BookSmart-services/core/models"
	"github.com/nikitalystsev/BookSmart-services/errs"
	"github.com/nikitalystsev/BookSmart-services/pkg/hash"
	jsondto "github.com/nikitalystsev/BookSmart-web-api/core/dto"
	jsonmodels "github.com/nikitalystsev/BookSmart-web-api/core/models"
	"net/http"
	"time"
)

// @Summary Метод регистрации пользователя
// @Tags auth
// @ID signUp
// @Accept  json
// @Produce  json
// @Param input body dto.SignUpInputDTO true "DTO c данными пользователя"
// @Success 201 "Успешное создание пользователя"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 409 {object} dto.ErrorResponse "Пользователь уже существует"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/auth/sign-up [post]
func (h *Handler) signUp(c *gin.Context) {
	var inp dto.SignUpInputDTO
	if err := c.BindJSON(&inp); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, jsondto.ErrorResponse{ErrorMsg: err.Error()})
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
	if err != nil && errors.Is(err, errs.ErrReaderAlreadyExist) {
		c.AbortWithStatusJSON(http.StatusConflict, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}

// @Summary Метод аутентификации пользователя
// @Tags auth
// @ID signIn
// @Accept  json
// @Produce  json
// @Param input body dto.SignInInputDTO true "DTO c номером телефона и паролем пользователя"
// @Success 200 {object} dto.SignInOutputDTO "Успешный вход пользователя"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 404 {object} dto.ErrorResponse "Читателя не существует"
// @Failure 409 {object} dto.ErrorResponse "Неверный логин или пароль"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/auth/sign-in [post]
func (h *Handler) signIn(c *gin.Context) {
	var inp dto.SignInInputDTO
	if err := c.BindJSON(&inp); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	res, err := h.readerService.SignIn(c.Request.Context(), inp.PhoneNumber, inp.Password)
	if err != nil && errors.Is(err, errs.ErrReaderDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if err != nil && errors.Is(err, hash.ErrInvalidLoginOrPassword) {
		c.AbortWithStatusJSON(http.StatusConflict, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if err != nil && errors.Is(err, errs.ErrReaderObjectIsNil) {
		c.AbortWithStatusJSON(http.StatusBadRequest, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	readerID, err := h.getReaderIDFromAccessToken(res.AccessToken)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.SignInOutputDTO{
		ReaderID:     readerID,
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
		ExpiredAt:    time.Now().Add(h.accessTokenTTL).UnixMilli(),
	})
}

// @Summary Метод обновления токенов
// @Tags auth
// @ID refresh
// @Accept  json
// @Produce  json
// @Param input body dto.RefreshTokenInputDTO true "Токен обновления"
// @Success 200 {object} dto.RefreshTokenOutputDTO "Успешное обновление токенов"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 404 {object} dto.ErrorResponse "Читателя не существует"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/auth/refresh [post]
func (h *Handler) refresh(c *gin.Context) {
	var inp dto.RefreshTokenInputDTO
	if err := c.BindJSON(&inp); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	var (
		res *models.Tokens
		err error
	)

	res, err = h.readerService.RefreshTokens(c.Request.Context(), inp.RefreshToken)
	if err != nil && errors.Is(err, errs.ErrReaderDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.RefreshTokenOutputDTO{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
		ExpiredAt:    time.Now().Add(h.accessTokenTTL).UnixMilli(),
	})
}

// @Summary Метод получения читателя по идентификатору
// @Security ApiKeyAuth
// @Tags reader
// @ID getReaderByPhoneNumber
// @Accept  json
// @Produce  json
// @Param id path string true "Идентификатор читателя"
// @Success 200 {object} models.JSONReaderModel "Успешное получение читателя"
// @Failure 401 {object} dto.ErrorResponse "Неавторизованный пользователь"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Читатель не найден"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/readers/{id} [get]
func (h *Handler) getReaderByID(c *gin.Context) {
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

	reader, err := h.readerService.GetByID(c.Request.Context(), readerID)
	if err != nil && errors.Is(err, errs.ErrReaderDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	c.JSON(http.StatusOK, h.convertToJSONReaderModel(reader))
}

func (h *Handler) getReaderIDFromAccessToken(accessToken string) (uuid.UUID, error) {
	readerIDStr, _, err := h.tokenManager.Parse(accessToken)
	if err != nil {
		return uuid.Nil, err
	}

	readerID, err := uuid.Parse(readerIDStr)
	if err != nil {
		return uuid.Nil, err
	}

	return readerID, nil
}

func (h *Handler) convertToJSONReaderModel(reader *models.ReaderModel) *jsonmodels.JSONReaderModel {
	return &jsonmodels.JSONReaderModel{
		ID:          reader.ID,
		Fio:         reader.Fio,
		PhoneNumber: reader.PhoneNumber,
		Age:         reader.Age,
		Password:    reader.Password,
		Role:        reader.Role,
	}
}
