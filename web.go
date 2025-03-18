package main

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-resty/resty/v2"
)

// OAuthHTTPServer implements a simple HTTP server for OAuth callbacks
type OAuthHTTPServer struct {
	port int
	s    *chi.Mux
	h    *handler
}

type handler struct {
	client      *resty.Client
	redirectURL string
}

// NewOAuthHTTPServer creates a new OAuth HTTP server
func NewOAuthHTTPServer(port int) *OAuthHTTPServer {
	var h = &handler{}

	mux := chi.NewRouter()
	mux.Use(middleware.Recoverer)
	mux.Get("/callback", h.handleOAuthCallback)
	mux.Get("/redirect", h.oauthRedirect)

	return &OAuthHTTPServer{
		port: port,
		s:    mux,
		h:    h,
	}
}

// Start initializes and starts the HTTP server
func (s *OAuthHTTPServer) Start() error {
	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), s.s)
}

func (h *handler) oauthRedirect(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	if state == "" {
		http.Error(w, "请求无效，请在 telegram 中重新认证", http.StatusBadRequest)
		return
	}

	redirectURL := "https://next.bgm.tv/oauth/authorize"

	query := url.Values{}
	query.Add("client_id", BANGUMI_APP_ID)
	query.Add("response_type", "code")
	query.Add("redirect_uri", "your_redirect_url")
	query.Add("state", state)

	http.Redirect(w, r, redirectURL+"?"+query.Encode(), http.StatusFound)
}

// handleOAuthCallback processes the OAuth callback request
func (h *handler) handleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" || state == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	resp, err := h.client.R().
		SetFormData(map[string]string{
			"client_id":     BANGUMI_APP_ID,
			"client_secret": BANGUMI_APP_SECRET,
			"grant_type":    "authorization_code",
			"code":          code,
			"redirect_uri":  h.redirectURL,
		}).
		Post("https://next.bgm.tv/oauth/access_token")

	if err != nil {
		panic(err)
	}

	fmt.Println(resp.Body())

	// Redirect or display success message
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(""))
}
