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
	jsonmodels "github.com/nikitalystsev/BookSmart-web-api/core/models"
	"net/http"
	"strconv"
)

func (h *Handler) getBooks(c *gin.Context) {
	fmt.Println("call getBooks")
	var params dto.BookParamsDTO
	params.Title = c.Query("title")
	params.Author = c.Query("author")
	params.Publisher = c.Query("publisher")
	params.Rarity = c.Query("rarity")
	params.Genre = c.Query("genre")
	params.Language = c.Query("language")
	if c.Query("copies_number") != "" && c.Query("copies_number") != "NaN" && c.Query("copies_number") != "null" {
		copiesNumber, err := strconv.ParseUint(c.Query("copies_number"), 10, 0)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
			return
		}
		params.CopiesNumber = uint(copiesNumber)
	}

	if c.Query("publishing_year") != "" && c.Query("publishing_year") != "NaN" && c.Query("publishing_year") != "null" {
		publishingYear, err := strconv.ParseUint(c.Query("publishing_year"), 10, 0)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
			return
		}
		params.PublishingYear = uint(publishingYear)
	}

	if c.Query("age_limit") != "" && c.Query("age_limit") != "NaN" && c.Query("age_limit") != "null" {
		ageLimit, err := strconv.ParseUint(c.Query("age_limit"), 10, 0)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
			return
		}
		params.AgeLimit = uint(ageLimit)
	}

	if c.Query("limit") != "" && c.Query("limit") != "NaN" && c.Query("limit") != "null" {
		limit, err := strconv.ParseUint(c.Query("limit"), 10, 0)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
			return
		}
		params.Limit = uint(limit)
	}
	var err error
	params.Offset, err = strconv.Atoi(c.Query("offset"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, "invalid offset")
		return
	}

	fmt.Println(params)

	books, err := h.bookService.GetByParams(c.Request.Context(), &params)
	if err != nil && errors.Is(err, errs.ErrBookDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	_books := make([]*jsonmodels.BookModel, len(books))
	for i, book := range books {
		_books[i] = h.convertToJSONBookModel(book)
	}

	c.JSON(http.StatusOK, _books)
}

func (h *Handler) getBookByID(c *gin.Context) {
	bookID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	book, err := h.bookService.GetByID(c.Request.Context(), bookID)
	if err != nil && errors.Is(err, errs.ErrBookDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	_book := h.convertToJSONBookModel(book)

	c.JSON(http.StatusOK, _book)
}

func (h *Handler) addToFavorites(c *gin.Context) {
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

	var bookID uuid.UUID
	if err = c.BindJSON(&bookID); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	err = h.readerService.AddToFavorites(c.Request.Context(), readerID, bookID)
	if err != nil && errors.Is(err, errs.ErrReaderDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, err.Error())
		return
	}
	if err != nil && errors.Is(err, errs.ErrBookDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, err.Error())
		return
	}
	if err != nil && errors.Is(err, errs.ErrBookAlreadyIsFavorite) {
		c.AbortWithStatusJSON(http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusCreated)
}

func (h *Handler) getRatingsByBookID(c *gin.Context) {
	fmt.Println("call getRatingsByBookID")

	bookID, err := uuid.Parse(c.Query("book_id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	ratings, err := h.ratingService.GetByBookID(c.Request.Context(), bookID)
	if err != nil && errors.Is(err, errs.ErrRatingDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	_ratings := make([]*dto.RatingOutputDTO, len(ratings))
	for i, rating := range ratings {
		_ratings[i], err = h.copyRatingModelToOutputDTO(c.Request.Context(), rating)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
			return
		}
	}

	c.JSON(http.StatusOK, _ratings)
}

func (h *Handler) addNewRating(c *gin.Context) {
	var ratingDTO dto.RatingInputDTO
	if err := c.BindJSON(&ratingDTO); err != nil {
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

	rating := &models.RatingModel{
		ID:       uuid.New(),
		ReaderID: readerID,
		BookID:   ratingDTO.BookID,
		Review:   ratingDTO.Review,
		Rating:   ratingDTO.Rating,
	}

	err = h.ratingService.Create(c.Request.Context(), rating)
	if errors.Is(err, errs.ErrRatingAlreadyExist) {
		c.AbortWithStatusJSON(http.StatusConflict, err.Error())
		return
	}
	if errors.Is(err, errs.ErrReservationDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusCreated)
}

func (h *Handler) getAvgRatingByBookID(c *gin.Context) {
	bookID, err := uuid.Parse(c.Query("book_id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	avgRating, err := h.ratingService.GetAvgRatingByBookID(c.Request.Context(), bookID)
	if err != nil && errors.Is(err, errs.ErrRatingDoesNotExists) {
		c.AbortWithStatusJSON(http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, dto.AvgRatingDTO{AvgRating: avgRating})
}

func (h *Handler) convertToJSONBookModel(book *models.BookModel) *jsonmodels.BookModel {
	return &jsonmodels.BookModel{
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

func (h *Handler) convertToJSONRatingModel(rating *models.RatingModel) *jsonmodels.RatingModel {
	return &jsonmodels.RatingModel{
		ID:       rating.ID,
		ReaderID: rating.ReaderID,
		BookID:   rating.BookID,
		Review:   rating.Review,
		Rating:   rating.Rating,
	}
}

func (h *Handler) copyRatingModelToOutputDTO(ctx context.Context, rating *models.RatingModel) (*dto.RatingOutputDTO, error) {
	reader, err := h.readerService.GetByID(ctx, rating.ReaderID)
	if err != nil && !errors.Is(err, errs.ErrReaderDoesNotExists) {
		return nil, err
	}
	if reader == nil {
		return nil, errs.ErrReaderDoesNotExists
	}

	return &dto.RatingOutputDTO{
		Reader: reader.Fio,
		Rating: rating.Rating,
		Review: rating.Review,
	}, nil
}
