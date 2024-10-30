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
	"github.com/nikitalystsev/BookSmart-services/impl"
	jsondto "github.com/nikitalystsev/BookSmart-web-api/core/dto"
	jsonmodels "github.com/nikitalystsev/BookSmart-web-api/core/models"
	weberrs "github.com/nikitalystsev/BookSmart-web-api/errs"
	"net/http"
	"strconv"
)

// @Summary Метод получения книг по параметрам
// @Tags book
// @ID getPageBooks
// @Accept json
// @Produce json
// @Param title query string false "Название книги"
// @Param author query string false "Автор книги"
// @Param publisher query string false "Издательство книги"
// @Param rarity query string false "Редкость книги"
// @Param genre query string false "Жанр книги"
// @Param language query string false "Язык книги"
// @Param copies_number query uint false "Количество копий"
// @Param publishing_year query uint false "Год издания"
// @Param age_limit query uint false "Возрастное ограничение"
// @Param page_number query uint true "Номер страницы для пагинации"
// @Success 200 {array} models.JSONBookModel "Список книг"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 404 {object} dto.ErrorResponse "Книги не найдены"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/books [get]
func (h *Handler) getPageBooks(c *gin.Context) {
	fmt.Println("call getPageBooks")
	var (
		params     dto.BookParamsDTO
		err        error
		pageNumber uint
	)

	params.Title = c.Query("title")
	params.Author = c.Query("author")
	params.Publisher = c.Query("publisher")
	params.Rarity = c.Query("rarity")
	params.Genre = c.Query("genre")
	params.Language = c.Query("language")
	if h.isNoEmptyField(c.Query("copies_number")) {
		if params.CopiesNumber, err = h.getUintFromStr(c.Query("copies_number")); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, jsondto.ErrorResponse{ErrorMsg: err.Error()})
			return
		}
	}
	if h.isNoEmptyField(c.Query("publishing_year")) {
		if params.PublishingYear, err = h.getUintFromStr(c.Query("publishing_year")); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, jsondto.ErrorResponse{ErrorMsg: err.Error()})
			return
		}
	}

	if h.isNoEmptyField(c.Query("age_limit")) {
		if params.AgeLimit, err = h.getUintFromStr(c.Query("age_limit")); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, jsondto.ErrorResponse{ErrorMsg: err.Error()})
			return
		}
	}

	if h.isNoEmptyField(c.Query("page_number")) {
		pageNumber, err = h.getUintFromStr(c.Query("page_number"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, jsondto.ErrorResponse{ErrorMsg: err.Error()})
			return
		}
		params.Limit = impl.PageLimit
		params.Offset = int((pageNumber - 1) * impl.PageLimit)
	}

	books, err := h.bookService.GetByParams(c.Request.Context(), &params)
	if err != nil && errors.Is(err, errs.ErrBookDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	c.JSON(http.StatusOK, h.convertArrayBooksToJSONBookModels(books))
}

// @Summary Метод получения книги по идентификатору
// @Tags book
// @ID getBookByID
// @Accept  json
// @Produce  json
// @Param id path string true "Идентификатор книги"
// @Success 200 {object} models.JSONBookModel "Успешное получение книги по идентификатору"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 404 {object} dto.ErrorResponse "Книги нет"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/books/{id} [get]
func (h *Handler) getBookByID(c *gin.Context) {
	bookID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	book, err := h.bookService.GetByID(c.Request.Context(), bookID)
	if err != nil && errors.Is(err, errs.ErrBookDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	c.JSON(http.StatusOK, h.convertToJSONBookModel(book))
}

// @Summary Метод добавления книги в избранное
// @Security ApiKeyAuth
// @Tags reader
// @ID addBookToFavorites
// @Accept  json
// @Produce  json
// @Param id path string true "Идентификатор читателя"
// @Param input body dto.FavoriteBookInputDTO true "Идентификатор книги"
// @Success 201 "Успешное добавление книги в избранное"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 401 {object} dto.ErrorResponse "Неавторизованный пользователь"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Читатель не найден"
// @Failure 409 {object} dto.ErrorResponse " Книга уже добавлена в избранное"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/readers/{id}/favorite_books [post]
func (h *Handler) addToFavorites(c *gin.Context) {
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

	var inp dto.FavoriteBookInputDTO
	if err = c.BindJSON(&inp); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	err = h.readerService.AddToFavorites(c.Request.Context(), readerID, inp.BookID)
	if err != nil && errors.Is(err, errs.ErrReaderDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if err != nil && errors.Is(err, errs.ErrBookAlreadyIsFavorite) {
		c.AbortWithStatusJSON(http.StatusConflict, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}

// @Summary Метод получения рейтингов книги
// @Tags book_ratings
// @ID getRatingsByBookID
// @Accept  json
// @Produce  json
// @Param id path string true "Идентификатор книги"
// @Success 200 {array} dto.RatingOutputDTO "Успешное получение отзывов на книгу"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 404 {object} dto.ErrorResponse "У книги нет отзывов"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/books/{id}/ratings [get]
func (h *Handler) getRatingsByBookID(c *gin.Context) {
	bookID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	ratings, err := h.ratingService.GetByBookID(c.Request.Context(), bookID)
	if err != nil && errors.Is(err, errs.ErrRatingDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	ratingOutputDTOs, err := h.copyRatingModelsToRatingOutputDTOs(c.Request.Context(), ratings)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	c.JSON(http.StatusOK, ratingOutputDTOs)
}

// @Summary Метод добавления отзыва на книгу
// @Security ApiKeyAuth
// @Tags book_ratings
// @ID addNewRating
// @Accept  json
// @Produce  json
// @Param id path string true "Идентификатор книги"
// @Param input body dto.RatingInputDTO true "DTO с данными отзыва"
// @Success 201 "Успешное добавление отзыва"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 401 {object} dto.ErrorResponse "Неавторизованный пользователь"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Пользователь никогда не бронировал книгу"
// @Failure 409 {object} dto.ErrorResponse "Пользователь уже оценил книгу"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/books/{id}/ratings [post]
func (h *Handler) addNewRating(c *gin.Context) {
	bookID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	var ratingDTO dto.RatingInputDTO
	if err = c.BindJSON(&ratingDTO); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	isReader, err := isReaderID(c, ratingDTO.ReaderID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if !isReader {
		c.AbortWithStatusJSON(http.StatusForbidden, jsondto.ErrorResponse{ErrorMsg: "access denied"})
		return
	}

	if err = h.checkRatingCanBeAdded(&ratingDTO); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	if ratingDTO.Rating < 0 || ratingDTO.Rating > 5 {
		c.AbortWithStatusJSON(http.StatusBadRequest, jsondto.ErrorResponse{ErrorMsg: "invalid rating"})
		return
	}

	rating := &models.RatingModel{
		ID:       uuid.New(),
		ReaderID: ratingDTO.ReaderID,
		BookID:   bookID,
		Review:   ratingDTO.Review,
		Rating:   ratingDTO.Rating,
	}

	err = h.ratingService.Create(c.Request.Context(), rating)
	if errors.Is(err, errs.ErrRatingAlreadyExist) {
		c.AbortWithStatusJSON(http.StatusConflict, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if errors.Is(err, errs.ErrReservationDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}

// @Summary Метод получения среднего рейтинга книги
// @Tags book_ratings
// @ID getAvgRatingByBookID
// @Accept  json
// @Produce  json
// @Param id path string true "Идентификатор книги"
// @Success 200 {object} dto.AvgRatingOutputDTO "Успешное получение среднего рейтинга книги"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 404 {object} dto.ErrorResponse "У книги нет отзывов"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/books/{id}/ratings/avg [get]
func (h *Handler) getAvgRatingByBookID(c *gin.Context) {
	bookID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	avgRating, err := h.ratingService.GetAvgRatingByBookID(c.Request.Context(), bookID)
	if err != nil && errors.Is(err, errs.ErrRatingDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, jsondto.ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.AvgRatingOutputDTO{AvgRating: avgRating})
}

func (h *Handler) convertArrayBooksToJSONBookModels(books []*models.BookModel) []*jsonmodels.JSONBookModel {
	jsonBooks := make([]*jsonmodels.JSONBookModel, len(books))
	for i, book := range books {
		jsonBooks[i] = h.convertToJSONBookModel(book)
	}

	return jsonBooks
}

func (h *Handler) convertToJSONBookModel(book *models.BookModel) *jsonmodels.JSONBookModel {
	return &jsonmodels.JSONBookModel{
		ID:             book.ID,
		Title:          book.Title,
		Author:         book.Author,
		Publisher:      book.Publisher,
		CopiesNumber:   book.CopiesNumber,
		Rarity:         book.Rarity,
		Genre:          book.Genre,
		PublishingYear: book.PublishingYear,
		Language:       book.Language,
		AgeLimit:       book.AgeLimit,
	}
}

func (h *Handler) isNoEmptyField(field string) bool {
	return field != "" && field != "NaN" && field != "null"
}

func (h *Handler) getUintFromStr(str string) (uint, error) {
	number, err := strconv.ParseUint(str, 10, 0)
	if err != nil {
		return 0, err
	}

	return uint(number), nil
}

func (h *Handler) convertToJSONRatingModel(rating *models.RatingModel) *jsonmodels.JSONRatingModel {
	return &jsonmodels.JSONRatingModel{
		ID:       rating.ID,
		ReaderID: rating.ReaderID,
		BookID:   rating.BookID,
		Review:   rating.Review,
		Rating:   rating.Rating,
	}
}

func (h *Handler) copyRatingModelToRatingOutputDTO(ctx context.Context, rating *models.RatingModel) (*dto.RatingOutputDTO, error) {
	reader, err := h.readerService.GetByID(ctx, rating.ReaderID)
	if err != nil && !errors.Is(err, errs.ErrReaderDoesNotExists) {
		return nil, err
	}
	if reader == nil {
		return nil, errs.ErrReaderDoesNotExists
	}

	ratingOutputDTO := dto.RatingOutputDTO{
		ReaderFio: reader.Fio,
		Rating:    rating.Rating,
		Review:    rating.Review,
	}

	return &ratingOutputDTO, nil
}

func (h *Handler) copyRatingModelsToRatingOutputDTOs(ctx context.Context, ratings []*models.RatingModel) ([]*dto.RatingOutputDTO, error) {
	ratingOutputDTOs := make([]*dto.RatingOutputDTO, len(ratings))

	for i, rating := range ratings {
		outputDTO, err := h.copyRatingModelToRatingOutputDTO(ctx, rating)
		if err != nil {
			return nil, err
		}
		ratingOutputDTOs[i] = outputDTO
	}

	return ratingOutputDTOs, nil
}

func (h *Handler) checkRatingCanBeAdded(ratingDTO *dto.RatingInputDTO) error {
	if ratingDTO.Rating < 0 || ratingDTO.Rating > 5 {
		return weberrs.ErrRatingOutOfBounds
	}
	return nil
}
