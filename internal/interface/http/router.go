package httpapi

import (
	"net/http"

	"github.com/iiimomoniii/backend-challenge/internal/interface/http/handler"
	"github.com/iiimomoniii/backend-challenge/internal/interface/http/middleware"
)

func NewRouter(userHandler *handler.UserHandler, authHandler *handler.AuthHandler, auth *middleware.Auth) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /register", userHandler.Create)
	mux.HandleFunc("POST /login", authHandler.Login)

	mux.Handle("GET /users", auth.Middleware(http.HandlerFunc(userHandler.SearchAll)))
	mux.Handle("GET /users/{id}", auth.Middleware(http.HandlerFunc(userHandler.Search)))
	mux.Handle("PATCH /users/{id}", auth.Middleware(http.HandlerFunc(userHandler.Update)))
	mux.Handle("DELETE /users/{id}", auth.Middleware(http.HandlerFunc(userHandler.Delete)))

	return middleware.Logger(mux)
}
