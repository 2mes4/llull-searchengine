package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func NewRouter(h *Handlers) http.Handler {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Use(loggingMiddleware)

	r.Get("/v1/health", h.Health)
	r.Get("/v1/indices", h.Indices)

	r.Post("/v1/index", h.IndexHandlerLegacy)
	r.Get("/v1/search", h.Search)

	r.Post("/v1/{index}/index", h.Index)
	r.Get("/v1/{index}/search", h.Search)
	r.Delete("/v1/{index}/documents/{id}", h.DeleteDocument)

	return r
}
