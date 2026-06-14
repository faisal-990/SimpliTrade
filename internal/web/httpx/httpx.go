// Package httpx defines the uniform HTTP response contract for the API: a single
// success envelope ({"data": ...}) and a single error envelope
// ({"error": {"code", "message"}}), plus typed application errors that map
// cleanly to HTTP status codes. Controllers return these instead of hand-rolling
// gin.H bodies, so the frontend can rely on one predictable shape everywhere.
package httpx

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Code is a stable, machine-readable error identifier the client can switch on
// without parsing human-facing messages.
type Code string

const (
	CodeBadRequest   Code = "bad_request"
	CodeValidation   Code = "validation_error"
	CodeUnauthorized Code = "unauthorized"
	CodeForbidden    Code = "forbidden"
	CodeNotFound     Code = "not_found"
	CodeConflict     Code = "conflict"
	CodeInternal     Code = "internal_error"
)

// AppError is a domain/transport error carrying everything needed to render a
// consistent HTTP response. Services return AppErrors; controllers pass them to
// Fail, which renders the envelope and status.
type AppError struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	status  int
	// err is the wrapped underlying cause, kept server-side for logging and
	// never serialized to the client.
	err error
}

func (e *AppError) Error() string {
	if e.err != nil {
		return string(e.Code) + ": " + e.Message + ": " + e.err.Error()
	}
	return string(e.Code) + ": " + e.Message
}

// Unwrap exposes the underlying cause for errors.Is/As.
func (e *AppError) Unwrap() error { return e.err }

// Status returns the HTTP status this error maps to.
func (e *AppError) Status() int { return e.status }

// WithCause attaches an underlying error for server-side logging without
// changing the client-facing message.
func (e *AppError) WithCause(err error) *AppError {
	clone := *e
	clone.err = err
	return &clone
}

func newAppError(code Code, status int, msg string) *AppError {
	return &AppError{Code: code, Message: msg, status: status}
}

// Constructors for the common cases. Message is client-safe (no internals).
func BadRequest(msg string) *AppError { return newAppError(CodeBadRequest, http.StatusBadRequest, msg) }
func Validation(msg string) *AppError { return newAppError(CodeValidation, http.StatusBadRequest, msg) }
func Unauthorized(msg string) *AppError {
	return newAppError(CodeUnauthorized, http.StatusUnauthorized, msg)
}
func Forbidden(msg string) *AppError { return newAppError(CodeForbidden, http.StatusForbidden, msg) }
func NotFound(msg string) *AppError  { return newAppError(CodeNotFound, http.StatusNotFound, msg) }
func Conflict(msg string) *AppError  { return newAppError(CodeConflict, http.StatusConflict, msg) }
func Internal(msg string) *AppError {
	return newAppError(CodeInternal, http.StatusInternalServerError, msg)
}

// successEnvelope and errorEnvelope are the only two response shapes the API emits.
type successEnvelope struct {
	Data any `json:"data"`
}

type errorEnvelope struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
}

// OK writes a 200 success envelope.
func OK(c *gin.Context, data any) { c.JSON(http.StatusOK, successEnvelope{Data: data}) }

// Created writes a 201 success envelope.
func Created(c *gin.Context, data any) { c.JSON(http.StatusCreated, successEnvelope{Data: data}) }

// Fail renders any error as the error envelope. AppErrors map to their declared
// status and code; any other error is treated as an opaque 500 so internal
// details never leak to clients.
func Fail(c *gin.Context, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		c.AbortWithStatusJSON(appErr.status, errorEnvelope{Error: errorBody{
			Code:    appErr.Code,
			Message: appErr.Message,
		}})
		return
	}
	c.AbortWithStatusJSON(http.StatusInternalServerError, errorEnvelope{Error: errorBody{
		Code:    CodeInternal,
		Message: "an unexpected error occurred",
	}})
}
