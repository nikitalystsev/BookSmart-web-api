package handlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nikitalystsev/BookSmart-services/core/models"
	"github.com/nikitalystsev/BookSmart-services/errs"
	jsondto "github.com/nikitalystsev/BookSmart-web-api/core/dto"
	jsonmodels "github.com/nikitalystsev/BookSmart-web-api/core/models"
	"net/http"
)

// @Summary Метод создания читательского билета
// @Security ApiKeyAuth
// @Tags reader_lib_card
// @ID createLibCard
// @Accept  json
// @Produce  json
// @Param id path string true "Идентификатор читателя"
// @Success 201 "Успешное создание читательского билета"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 401 {object} dto.ErrorResponse "Неавторизованный пользователь"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 409 {object} dto.ErrorResponse "Читательский билет уже существует"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/readers/{id}/lib_cards [post]
func (h *Handler) createLibCard(c *gin.Context) {
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

	err = h.libCardService.Create(c.Request.Context(), readerID)
	if err != nil && errors.Is(err, errs.ErrLibCardAlreadyExist) {
		c.AbortWithStatusJSON(http.StatusConflict, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}

// @Summary Метод обновления читательского билета
// @Security ApiKeyAuth
// @Tags reader_lib_card
// @ID updateLibCard
// @Accept  json
// @Produce  json
// @Param id path string true "Идентификатор читателя"
// @Success 200 "Успешное обновление читательского билета"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 401 {object} dto.ErrorResponse "Неавторизованный пользователь"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Читательского билета не существует"
// @Failure 409 {object} dto.ErrorResponse "Читательский билет актуален"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/readers/{id}/lib_cards [put]
func (h *Handler) updateLibCard(c *gin.Context) {
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

	libCard, err := h.libCardService.GetByReaderID(c.Request.Context(), readerID)
	if err != nil && !errors.Is(err, errs.ErrLibCardDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusInternalServerError, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if err != nil && errors.Is(err, errs.ErrLibCardDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	err = h.libCardService.Update(c.Request.Context(), libCard)
	if err != nil && errors.Is(err, errs.ErrLibCardDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if err != nil && errors.Is(err, errs.ErrLibCardIsValid) {
		c.AbortWithStatusJSON(http.StatusConflict, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if err != nil && errors.Is(err, errs.ErrLibCardObjectIsNil) {
		c.AbortWithStatusJSON(http.StatusBadRequest, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// @Summary Метод получения читательского билета
// @Security ApiKeyAuth
// @Tags reader_lib_card
// @ID getLibCardByReaderID
// @Accept  json
// @Produce  json
// @Param id path string true "Идентификатор читателя"
// @Success 200 {array} models.JSONLibCardModel "Успешное получение читательского билета"
// @Failure 400 {object} dto.ErrorResponse "Некорректный идентификатор пользователя"
// @Failure 401 {object} dto.ErrorResponse "Неавторизованный пользователь"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Читательский билет не найден"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/readers/{id}/lib_cards [get]
func (h *Handler) getLibCardByReaderID(c *gin.Context) {
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

	libCard, err := h.libCardService.GetByReaderID(c.Request.Context(), readerID)
	if err != nil && !errors.Is(err, errs.ErrLibCardDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusInternalServerError, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if err != nil && errors.Is(err, errs.ErrLibCardDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	libCards := []*models.LibCardModel{libCard} // один чит билет упаковываю в массив

	c.JSON(http.StatusOK, h.convertArrayToJSONLibCardModels(libCards))
}

func (h *Handler) convertArrayToJSONLibCardModels(libCards []*models.LibCardModel) []*jsonmodels.JSONLibCardModel {
	jsonLibCardModels := make([]*jsonmodels.JSONLibCardModel, len(libCards))
	for i, libCard := range libCards {
		jsonLibCardModels[i] = h.convertToJSONLibCardModel(libCard)
	}

	return jsonLibCardModels
}

func (h *Handler) convertToJSONLibCardModel(libCard *models.LibCardModel) *jsonmodels.JSONLibCardModel {
	return &jsonmodels.JSONLibCardModel{
		ID:           libCard.ID,
		LibCardNum:   libCard.LibCardNum,
		Validity:     libCard.Validity,
		IssueDate:    libCard.IssueDate,
		ActionStatus: libCard.ActionStatus,
	}
}
