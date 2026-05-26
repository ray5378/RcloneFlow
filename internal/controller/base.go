package controller

import (
	"encoding/json"
	"errors"
	"net/http"

	"rcloneflow/internal/service"
)

const maxRequestBodyBytes = 10 << 20

type ResponseWriter struct {
	http.ResponseWriter
}

func WriteJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func DecodeRequest(w http.ResponseWriter, r *http.Request, dst any) error {
	return json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)).Decode(dst)
}

func mapServiceError(err error) (string, int) {
	switch {
	case errors.Is(err, service.ErrTaskNotFound):
		return "task not found", 404
	case errors.Is(err, service.ErrTaskNameExists):
		return "task name already exists", 409
	case errors.Is(err, service.ErrScheduleNotFound):
		return "schedule not found", 404
	case errors.Is(err, service.ErrRunNotFound):
		return "run not found", 404
	case errors.Is(err, service.ErrBisyncCurrentVersion):
		return "cannot delete current version", 400
	default:
		return "internal error", 500
	}
}
