package idemhttp

import (
	"errors"
	"net/http"

	"github.com/victorotene80/idemkit/idem"
)

func StatusFromError(err error) int {
	if err == nil {
		return http.StatusOK
	}

	switch {
	case errors.Is(err, idem.ErrConflict):
		return http.StatusConflict
	case errors.Is(err, idem.ErrInProgress):
		return http.StatusConflict
	case errors.Is(err, idem.ErrReplayWithFailure):
		return http.StatusConflict
	case errors.Is(err, idem.ErrStoreUnavailable):
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
