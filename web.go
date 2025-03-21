package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"github.com/mymmrac/telego"
	"github.com/redis/rueidis"
)

// OAuthHTTPServer implements a simple HTTP server for OAuth callbacks
type OAuthHTTPServer struct {
	port int
	s    *chi.Mux
	h    *handler
}

type handler struct {
	pg          *sqlx.DB
	bot         *telego.Bot
	redis       rueidis.Client
	client      *resty.Client
	redirectURL string
}

// NewOAuthHTTPServer creates a new OAuth HTTP server
func NewOAuthHTTPServer(pg *sqlx.DB, redis rueidis.Client, bot *telego.Bot, port int) *OAuthHTTPServer {
	var url = EXTERNAL_HTTP_ADDRESS
	if url == "" {
		url = "http://127.0.0.1:4562"
	}

	var h = &handler{
		pg:          pg,
		bot:         bot,
		redis:       redis,
		redirectURL: strings.TrimRight(url, "/") + "/callback",
	}

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
func (h *handler) handleOAuthCallback(w http.ResponseWriter, req *http.Request) {
	// Get query parameters
	q := req.URL.Query()
	code := q.Get("code")
	state := q.Get("state")

	if code == "" || state == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	var r OAuthAccessTokenResponse

	resp, err := h.client.R().
		SetFormData(map[string]string{
			"client_id":     BANGUMI_APP_ID,
			"client_secret": BANGUMI_APP_SECRET,
			"grant_type":    "authorization_code",
			"code":          code,
			"redirect_uri":  h.redirectURL,
		}).
		SetResult(&r).
		Post("https://next.bgm.tv/oauth/access_token")

	if err != nil {
		panic(err)
	}

	if resp.StatusCode() >= 300 {
		http.Error(w, "请求错误", http.StatusBadRequest)
		return
	}

	v, err := h.redis.Do(req.Context(), h.redis.B().Get().Key("tg-bot-oauth:"+state).Build()).AsBytes()
	if err != nil {
		return
	}
	var redisState RedisOAuthState
	json.Unmarshal(v, &state)

	_, err = h.pg.ExecContext(req.Context(), `
	INSERT INTO telegram_notify_chat(chat_id, user_id, disabled) VALUES ($1, $2, 0)`,
		redisState.ChatID,
		r.UserID,
	)

	_, _ = h.bot.SendMessage(req.Context(), &telego.SendMessageParams{
		BusinessConnectionID: "",
		ChatID:               telego.ChatID{ID: redisState.ChatID},
		Text:                 fmt.Sprintf("已经成功关联用户 %s", r.UserID),
	})

	// Redirect or display success message
	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write([]byte("你已经成功认证，请关闭页面返回 telegram"))
}

type OAuthAccessTokenResponse struct {
	// AccessToken  string `json:"access_token"`
	// TokenType    string `json:"token_type"`
	// ExpiresIn    int    `json:"expires_in"`
	// RefreshToken string `json:"refresh_token"`
	// Scope        string `json:"scope"`
	UserID string `json:"user_id"`
}

// RedisOAuthState represents the OAuth state stored in Redis
type RedisOAuthState struct {
	ChatID int64 `json:"chat_id" db:"chat_id"`
}
