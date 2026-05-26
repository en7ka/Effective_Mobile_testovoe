package http

import (
	"log"
	nethttp "net/http"
)

func NewRouter(handler *SubscriptionHandler, logger *log.Logger) nethttp.Handler {
	mux := nethttp.NewServeMux()

	mux.HandleFunc("/subscriptions/total", handler.HandleTotal)
	mux.HandleFunc("/subscriptions/", handler.HandleSubscriptionByID)
	mux.HandleFunc("/subscriptions", handler.HandleSubscriptions)

	return loggingMiddleware(logger)(rateLimitMiddleware(10, 20)(mux))
}
