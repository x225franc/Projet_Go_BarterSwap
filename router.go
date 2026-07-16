package main

import "net/http"

func newRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /docs", handleSwaggerUI)
	mux.HandleFunc("GET /openapi.json", handleOpenAPISpec)

	mux.HandleFunc("POST /api/users", handleCreateUser)
	mux.HandleFunc("GET /api/users/{id}", handleGetUser)
	mux.HandleFunc("PUT /api/users/{id}", handleUpdateUser)
	mux.HandleFunc("GET /api/users/{id}/skills", handleGetSkills)
	mux.HandleFunc("PUT /api/users/{id}/skills", handleUpdateSkills)
	mux.HandleFunc("GET /api/users/{id}/reviews", handleGetUserReviews)
	mux.HandleFunc("GET /api/users/{id}/stats", handleGetUserStats)

	mux.HandleFunc("POST /api/services", handleCreateService)
	mux.HandleFunc("GET /api/services", handleGetServices)
	mux.HandleFunc("GET /api/services/{id}", handleGetService)
	mux.HandleFunc("PUT /api/services/{id}", handleUpdateService)
	mux.HandleFunc("DELETE /api/services/{id}", handleDeleteService)
	mux.HandleFunc("GET /api/services/{id}/reviews", handleGetServiceReviews)

	mux.HandleFunc("POST /api/exchanges", handleCreateExchange)
	mux.HandleFunc("GET /api/exchanges", handleGetExchanges)
	mux.HandleFunc("GET /api/exchanges/{id}", handleGetExchangeByID)
	mux.HandleFunc("PUT /api/exchanges/{id}/accept", handleAcceptExchange)
	mux.HandleFunc("PUT /api/exchanges/{id}/reject", handleRejectExchange)
	mux.HandleFunc("PUT /api/exchanges/{id}/complete", handleCompleteExchange)
	mux.HandleFunc("PUT /api/exchanges/{id}/cancel", handleCancelExchange)
	mux.HandleFunc("POST /api/exchanges/{id}/review", handleCreateReview)

	return withMiddlewares(mux, loggingMiddleware, recoveryMiddleware, corsMiddleware)
}
