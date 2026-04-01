/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package router registers all HTTP routes for the API gateway.
package router

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/origadmin/runtime/log"
	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/gateway/handler"
	"origadmin/application/origcms/internal/gateway/middleware"
)

// Router holds all handler dependencies for route registration.
type Router struct {
	feed    *handler.FeedHandler
	detail  *handler.DetailHandler
	search  *handler.SearchHandler
	profile *handler.ProfileHandler
	auth    *auth.Manager
	log     *log.Helper
}

// NewRouter creates a new Router with all required handlers.
func NewRouter(
	feed *handler.FeedHandler,
	detail *handler.DetailHandler,
	search *handler.SearchHandler,
	profile *handler.ProfileHandler,
	authMgr *auth.Manager,
	logger log.Logger,
) *Router {
	return &Router{
		feed:    feed,
		detail:  detail,
		search:  search,
		profile: profile,
		auth:    authMgr,
		log:     log.NewHelper(log.With(logger, "module", "gateway.router")),
	}
}

// Build returns an http.Handler with all routes registered.
func (rt *Router) Build(logger log.Logger) http.Handler {
	mux := http.NewServeMux()

	// Health check (no auth required)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "svc-api-gateway"})
	})

	// Public routes (no auth required)
	mux.HandleFunc("GET /api/v1/feed", rt.handleHomeFeed)
	mux.HandleFunc("GET /api/v1/search", rt.handleSearch)
	mux.HandleFunc("GET /api/v1/media/{id}", rt.handleVideoDetail)
	mux.HandleFunc("GET /api/v1/users/{id}/profile", rt.handleUserProfile)

	// Example of a protected route (requires M1 implementation of actual business logic)
	// mux.Handle("POST /api/v1/media", middleware.JWTAuth(rt.auth)(http.HandlerFunc(rt.handleUploadMedia)))

	// Apply global middleware chain
	var h http.Handler = mux
	h = middleware.Logger(logger)(h)
	h = middleware.CORS()(h)

	return h
}

func (rt *Router) handleHomeFeed(w http.ResponseWriter, r *http.Request) {
	resp, err := rt.feed.GetHomeFeed(r.Context(), &handler.HomeFeedRequest{PageSize: 10})
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, resp)
}

func (rt *Router) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	resp, err := rt.search.Search(r.Context(), &handler.SearchRequest{Query: q, PageSize: 20})
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, resp)
}

func (rt *Router) handleVideoDetail(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid media ID"}`, http.StatusBadRequest)
		return
	}

	resp, err := rt.detail.GetVideoDetail(r.Context(), &handler.VideoDetailRequest{ID: id})
	if err != nil {
		http.Error(w, `{"error":"media not found"}`, http.StatusNotFound)
		return
	}
	writeJSON(w, resp)
}

func (rt *Router) handleUserProfile(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid user ID"}`, http.StatusBadRequest)
		return
	}

	resp, err := rt.profile.GetUserProfile(r.Context(), &handler.UserProfileRequest{ID: id})
	if err != nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
		return
	}
	writeJSON(w, resp)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
