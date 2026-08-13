package controller

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gmorini/inge-soft-3/backend/internal/identity/service"
	"github.com/gmorini/inge-soft-3/backend/internal/platform/requestctx"
)

func authenticate(
	tokens *service.TokenManager,
	logger *slog.Logger,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		scheme, tokenValue, found := strings.Cut(header, " ")
		if !found || scheme != "Bearer" || tokenValue == "" || strings.Contains(tokenValue, " ") {
			writeInvalidToken(w)
			return
		}

		identity, err := tokens.Validate(tokenValue)
		if err != nil {
			logger.Warn("access token rejected", "path", r.URL.Path)
			writeInvalidToken(w)
			return
		}

		ctx := requestctx.WithIdentity(r.Context(), requestctx.Identity{
			UserID:   identity.UserID,
			Username: identity.Username,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func writeInvalidToken(w http.ResponseWriter) {
	writeError(
		w,
		http.StatusUnauthorized,
		"invalid_token",
		"Token de acceso inválido o vencido.",
		nil,
	)
}
