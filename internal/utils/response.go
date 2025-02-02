package utils

import (
	"github.com/gin-gonic/gin"
)

// APIResponse defines the standard API response format
type APIResponse struct {
	Status    int         `json:"status"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	ErrorCode string      `json:"error_code"`
}

// ResponseBuilder helps construct API responses
type ResponseBuilder struct {
	response APIResponse
}

// NewResponse creates a new response builder
func NewResponse() *ResponseBuilder {
	return &ResponseBuilder{
		response: APIResponse{},
	}
}

func (rb *ResponseBuilder) SetStatus(status int) *ResponseBuilder {
	rb.response.Status = status
	return rb
}

func (rb *ResponseBuilder) SetMessage(message string) *ResponseBuilder {
	rb.response.Message = message
	return rb
}

func (rb *ResponseBuilder) SetData(data interface{}) *ResponseBuilder {
	rb.response.Data = data
	return rb
}

func (rb *ResponseBuilder) SetErrorCode(code string) *ResponseBuilder {
	rb.response.ErrorCode = code
	return rb
}

func (rb *ResponseBuilder) Build(ctx *gin.Context) {
	ctx.JSON(rb.response.Status, rb.response)
}
