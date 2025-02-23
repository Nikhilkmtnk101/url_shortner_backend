package handler

import (
	"golang.org/x/crypto/bcrypt"
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
	url, err := h.urlService.CreateShortURL(
		ctx,
		userID,
		&createShortURLRequest,
	)
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

func (h *URLHandler) CreateBulkShortURLs(ctx *gin.Context) {
	var createBulkShortURLsRequest []dto.CreateShortURLRequest
	if err := ctx.ShouldBindJSON(&createBulkShortURLsRequest); err != nil {
		utils.NewResponse().
			SetStatus(http.StatusBadRequest).
			SetMessage("Invalid request payload").
			SetErrorCode("BAD_REQUEST").
			SetData(nil).
			Build(ctx)
		return
	}
	userID := ctx.GetUint("user_id")
	for _, request := range createBulkShortURLsRequest {
		if request.ExpiresDays == 0 {
			request.ExpiresDays = 30
		}
	}
	urls, err := h.urlService.CreateShortURLs(ctx, userID, createBulkShortURLsRequest)
	if err != nil {
		utils.NewResponse().
			SetStatus(http.StatusInternalServerError).
			SetMessage("Failed to create short URLs").
			SetErrorCode("INTERNAL_ERROR").
			SetData(nil).
			Build(ctx)
		return
	}
	utils.NewResponse().
		SetStatus(http.StatusCreated).
		SetMessage("Short URLs created successfully").
		SetErrorCode("").
		SetData(urls).
		Build(ctx)
}

func (h *URLHandler) RedirectToLongURL(ctx *gin.Context) {
	shortCode := ctx.Param("shortCode")
	password := ctx.DefaultQuery("password", "")

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
	err1 := bcrypt.CompareHashAndPassword([]byte(longURL.Password), []byte(""))
	err2 := bcrypt.CompareHashAndPassword([]byte(longURL.Password), []byte(password))
	if err1 != nil && err2 != nil {
		ctx.HTML(http.StatusOK, "password_form.html", gin.H{"shortCode": shortCode})
		return
	}
	ctx.Redirect(http.StatusMovedPermanently, longURL.LongURL)
}

func (h *URLHandler) GenerateQRCode(ctx *gin.Context) {
	shortCode := ctx.Param("shortCode")
	if shortCode == "" {
		utils.NewResponse().
			SetStatus(http.StatusBadRequest).
			SetMessage("URL parameter is required").
			SetErrorCode("BAD_REQUEST").
			SetData(nil).
			Build(ctx)
		return
	}
	qrCodeBase64, err := h.urlService.GenerateQRCodeBase64(ctx, shortCode)
	if err != nil {
		utils.NewResponse().
			SetStatus(http.StatusInternalServerError).
			SetMessage("Failed to generate QRCode").
			SetErrorCode("INTERNAL_ERROR").
			SetData(nil).
			Build(ctx)
		return
	}
	data := map[string]interface{}{
		"qrCodeBase64": qrCodeBase64,
	}
	utils.NewResponse().
		SetStatus(http.StatusOK).
		SetMessage("QRCode generated successfully").
		SetErrorCode("").
		SetData(data).
		Build(ctx)
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
