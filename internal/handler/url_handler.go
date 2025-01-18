package handler

import (
	"github.com/gin-gonic/gin"
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

func (h *URLHandler) CreateShortURL(c *gin.Context) {
	var req struct {
		LongURL     string `json:"long_url" binding:"required,url"`
		ExpiresDays int    `json:"expires_days" binding:"omitempty,min=0,max=365"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	url, err := h.urlService.CreateShortURL(userID, req.LongURL, req.ExpiresDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, url)
}

func (h *URLHandler) RedirectToLongURL(c *gin.Context) {
	shortCode := c.Param("shortCode")
	longURL, err := h.urlService.GetLongURL(shortCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found or expired"})
		return
	}

	c.Redirect(http.StatusMovedPermanently, longURL)
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
