package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nikhil/url-shortner-backend/internal/dto"
	"github.com/nikhil/url-shortner-backend/internal/service"
	"github.com/nikhil/url-shortner-backend/internal/utils"
)

type URLHandler struct {
	urlService *service.URLService
}

func NewURLHandler(urlService *service.URLService) *URLHandler {
	return &URLHandler{
		urlService: urlService,
	}
}

func (h *URLHandler) CreateShortURL(ctx *gin.Context) {
	var createShortURLRequest dto.CreateShortURLRequest

	if err := ctx.ShouldBindJSON(&createShortURLRequest); err != nil {
		utils.NewResponse().
			SetStatus(http.StatusBadRequest).
			SetMessage("Invalid request payload").
			SetErrorCode("BAD_REQUEST").
			SetData(nil).
			Build(ctx)
		return
	}

	if createShortURLRequest.ExpiresDays == 0 {
		createShortURLRequest.ExpiresDays = 30
	}

	userID := ctx.GetUint("user_id")
	url, err := h.urlService.CreateShortURL(userID, createShortURLRequest.LongURL, createShortURLRequest.ExpiresDays)
	if err != nil {
		utils.NewResponse().
			SetStatus(http.StatusInternalServerError).
			SetMessage("Failed to create short URL").
			SetErrorCode("INTERNAL_ERROR").
			SetData(nil).
			Build(ctx)
		return
	}

	utils.NewResponse().
		SetStatus(http.StatusCreated).
		SetMessage("Short URL created successfully").
		SetErrorCode("").
		SetData(url).
		Build(ctx)
}

func (h *URLHandler) RedirectToLongURL(ctx *gin.Context) {
	shortCode := ctx.Param("shortCode")
	longURL, err := h.urlService.GetLongURL(ctx, shortCode)
	if err != nil {
		utils.NewResponse().
			SetStatus(http.StatusNotFound).
			SetMessage("URL not found or expired").
			SetErrorCode("NOT_FOUND").
			SetData(nil).
			Build(ctx)
		return
	}

	ctx.Redirect(http.StatusMovedPermanently, longURL)
}

func (h *URLHandler) GetUserURLs(c *gin.Context) {
	userID := c.GetUint("user_id")
	urls, err := h.urlService.GetUserURLs(userID)
	if err != nil {
		utils.NewResponse().
			SetStatus(http.StatusInternalServerError).
			SetMessage("Failed to fetch user URLs").
			SetErrorCode("INTERNAL_ERROR").
			SetData(nil).
			Build(c)
		return
	}

	utils.NewResponse().
		SetStatus(http.StatusOK).
		SetMessage("User URLs fetched successfully").
		SetErrorCode("").
		SetData(urls).
		Build(c)
}
