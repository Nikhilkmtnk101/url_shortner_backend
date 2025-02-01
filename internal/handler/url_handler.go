package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/nikhil/url-shortner-backend/internal/dto"
	"github.com/nikhil/url-shortner-backend/internal/service"
	"net/http"
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
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if createShortURLRequest.ExpiresDays == 0 {
		createShortURLRequest.ExpiresDays = 30
	}

	userID := ctx.GetUint("user_id")
	url, err := h.urlService.CreateShortURL(userID, createShortURLRequest.LongURL, createShortURLRequest.ExpiresDays)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, url)
}

func (h *URLHandler) RedirectToLongURL(ctx *gin.Context) {
	shortCode := ctx.Param("shortCode")
	longURL, err := h.urlService.GetLongURL(ctx, shortCode)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "URL not found or expired"})
		return
	}

	ctx.Redirect(http.StatusMovedPermanently, longURL)
}

func (h *URLHandler) GetUserURLs(c *gin.Context) {
	userID := c.GetUint("user_id")
	urls, err := h.urlService.GetUserURLs(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, urls)
}
