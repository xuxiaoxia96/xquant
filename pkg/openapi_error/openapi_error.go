package openapi_error

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// copied from model-proxy protocol
type OpenAPIError struct {
	Code    string `json:"code" yaml:"code" mapstructure:"code"`
	Message string `json:"message" yaml:"message" mapstructure:"message"`
	Param   string `json:"param" yaml:"param" mapstructure:"param"`
	Type    string `json:"type" yaml:"type" mapstructure:"type"`

	// for apig use
	StatusCode int `json:"status_code" yaml:"status_code" mapstructure:"http_status_code"`
}

type ErrorResponse struct {
	// Error corresponds to the JSON schema field "error".
	Error OpenAPIError `json:"error" yaml:"error" mapstructure:"error"`
}

func (e OpenAPIError) Error() string {
	return fmt.Sprintf("Code: %s, Message: %s, Param: %s, Type: %s", e.Code, e.Message, e.Param, e.Type)
}

func (e OpenAPIError) HTTPStatusCode() int {
	return e.StatusCode
}

func NewInvalidParameterError(ctx context.Context, field, reason string, a ...any) OpenAPIError {
	return OpenAPIError{
		Code:    "InvalidParameter",
		Message: fmt.Sprintf("The parameter `%s` specified in the request are not valid: %s. Request id: %s", field, fmt.Sprintf(reason, a...), api_utils.MustFromCtxRequestID(ctx)),
		Param:   field,
		Type:    strings.ReplaceAll(http.StatusText(http.StatusBadRequest), " ", ""),

		StatusCode: http.StatusBadRequest,
	}
}

func NewAuthenticationError(ctx context.Context) OpenAPIError {
	return OpenAPIError{
		Code:    "AuthenticationError",
		Message: "The API key in the request is missing or invalid. Request id: " + api_utils.MustFromCtxRequestID(ctx),
		Type:    strings.ReplaceAll(http.StatusText(http.StatusUnauthorized), " ", ""),

		StatusCode: http.StatusUnauthorized,
	}
}

func NewInternalServiceError(ctx context.Context) OpenAPIError {
	return OpenAPIError{
		Code:    "InternalServiceError",
		Message: "The service encountered an unexpected internal error. Request id: " + api_utils.MustFromCtxRequestID(ctx),
		Type:    strings.ReplaceAll(http.StatusText(http.StatusInternalServerError), " ", ""),

		StatusCode: http.StatusInternalServerError,
	}
}

func NewAccessDenied(ctx context.Context) OpenAPIError {
	return OpenAPIError{
		Code:    "AccessDenied",
		Message: "The request failed because you do not have access to the requested resource. Request id: " + api_utils.MustFromCtxRequestID(ctx),
		Type:    strings.ReplaceAll(http.StatusText(http.StatusForbidden), " ", ""),

		StatusCode: http.StatusForbidden,
	}
}

func NewInvalidEndpoint(ctx context.Context) OpenAPIError {
	return OpenAPIError{
		Code:    "InvalidEndpoint.NotFound",
		Message: "The request targeted an endpoint that does not exist or is invalid. Request id: " + api_utils.MustFromCtxRequestID(ctx),
		Type:    strings.ReplaceAll(http.StatusText(http.StatusBadRequest), " ", ""),

		StatusCode: http.StatusNotFound,
	}
}

func NewInvalidBot(ctx context.Context) OpenAPIError {
	return OpenAPIError{
		Code:    "InvalidBot",
		Message: "The request targeted bot that does not exist or is invalid. Request id: " + api_utils.MustFromCtxRequestID(ctx),
		Type:    strings.ReplaceAll(http.StatusText(http.StatusBadRequest), " ", ""),

		StatusCode: http.StatusBadRequest,
	}
}
