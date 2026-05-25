package http

import nethttp "net/http"

func NewRouter(handler *SubscriptionHandler) nethttp.Handler {
	mux := nethttp.NewServeMux()

	mux.HandleFunc("/subscriptions/total", handler.HandleTotal)
	mux.HandleFunc("/subscriptions/", handler.HandleSubscriptionByID)
	mux.HandleFunc("/subscriptions", handler.HandleSubscriptions)

	return mux
}
