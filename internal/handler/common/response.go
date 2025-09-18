package common

import (
	"context"
	"errors"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/pkg/kernel"
)

// RespondError writes a JSON error payload with a status derived from known domain errors.
func RespondError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		httpx.OkJsonCtx(r.Context(), w, map[string]string{"message": "ok"})
		return
	}

	status := http.StatusInternalServerError

	switch {
	case errors.Is(err, repository.ErrNotFound), errors.Is(err, kernel.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, repository.ErrInvalidArgument):
		status = http.StatusBadRequest
	case errors.Is(err, repository.ErrConflict):
		status = http.StatusConflict
	case errors.Is(err, repository.ErrForbidden):
		status = http.StatusForbidden
	case errors.Is(err, repository.ErrUnauthorized):
		status = http.StatusUnauthorized
	case errors.Is(err, kernel.ErrProviderNotFound):
		status = http.StatusBadRequest
	case errors.Is(err, kernel.ErrNotImplemented):
		status = http.StatusNotImplemented
	case errors.Is(err, context.Canceled):
		status = http.StatusRequestTimeout
	}

	if status == http.StatusInternalServerError {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	httpx.WriteJsonCtx(r.Context(), w, status, map[string]any{
		"message": err.Error(),
	})
}
