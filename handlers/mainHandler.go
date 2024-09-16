package handlers

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nikitalystsev/BookSmart-services/intf"
	"github.com/nikitalystsev/BookSmart-services/pkg/auth"
	"github.com/nikitalystsev/BookSmart-services/pkg/hash"
	"io"
	"net/http"
	"time"
)

type Handler struct {
	bookService        intf.IBookService
	libCardService     intf.ILibCardService
	readerService      intf.IReaderService
	reservationService intf.IReservationService
	ratingService      intf.IRatingService
	tokenManager       auth.ITokenManager
	hasher             hash.IPasswordHasher
	accessTokenTTL     time.Duration
	refreshTokenTTL    time.Duration
}

func NewHandler(
	bookService intf.IBookService,
	libCardService intf.ILibCardService,
	readerService intf.IReaderService,
	reservationService intf.IReservationService,
	ratingService intf.IRatingService,
	tokenManager auth.ITokenManager,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) *Handler {
	return &Handler{
		bookService:        bookService,
		libCardService:     libCardService,
		readerService:      readerService,
		reservationService: reservationService,
		ratingService:      ratingService,
		tokenManager:       tokenManager,
		accessTokenTTL:     accessTokenTTL,
		refreshTokenTTL:    refreshTokenTTL,
	}
}

func (h *Handler) InitRoutes() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard

	router := gin.Default()

	router.Use(h.corsSettings())

	authenticate := router.Group("/auth")
	{
		authenticate.POST("/sign-up", h.signUp)
		authenticate.POST("/sign-in", h.signIn)
		authenticate.POST("/admin/sign-in", h.signInAsAdmin)
		authenticate.POST("/refresh", h.refresh)
	}

	general := router.Group("/")
	{
		general.GET("/books", h.getBooks)
		general.GET("/books/:id", h.getBookByID)
		general.GET("/ratings", h.getRatingsByBookID)
		general.GET("/ratings/avg", h.getAvgRatingByBookID)
	}

	api := router.Group("/api", h.readerIdentity)
	{
		api.POST("/favorites", h.addToFavorites)
		api.GET("/readers", h.getReaderByPhoneNumber)

		api.POST("/lib-cards", h.createLibCard)
		api.PUT("/lib-cards", h.updateLibCard)
		api.GET("/lib-cards", h.getLibCardByReaderID)

		api.POST("/reservations", h.reserveBook)
		api.GET("/reservations", h.getReservationsByReaderID)
		api.GET("/reservations/:id", h.getReservationsByID)
		api.PUT("/reservations/:id", h.updateReservation)

		api.POST("/ratings", h.addNewRating)

		admin := api.Group("/admin")
		{
			admin.DELETE("/books/:id", h.deleteBook)
			admin.POST("/books", h.addNewBook)
			admin.GET("/reservations", h.getReservationsByBookID)
		}
	}

	return router
}

func (h *Handler) corsSettings() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowMethods: []string{
			http.MethodPost,
			http.MethodGet,
			http.MethodPut,
			http.MethodDelete,
		},
		AllowOrigins: []string{
			"*",
		},
		AllowCredentials: true,
		AllowHeaders: []string{
			"Authorization",
			"Content-Type",
		},
		ExposeHeaders: []string{
			"Content-Type",
		},
	})
}
