package main

import (
	"context"
	crand "crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	mrand "math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func init() {
	mrand.Seed(time.Now().UnixNano())
}

// registerAuthRoutes sets up auth-related routes and a sample protected route.
func registerAuthRoutes(r *gin.Engine, db *sql.DB) {
	auth := r.Group("/api/v1/auth")
	auth.GET("/google/login", googleLoginHandler)
	auth.GET("/google/callback", func(c *gin.Context) { googleCallbackHandler(c, db) })

	// Protected user endpoint
	r.GET("/api/v1/user/me", AuthMiddleware(db), func(c *gin.Context) { userMeHandler(c, db) })
}

func oauthConfigFromEnv() *oauth2.Config {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirect := os.Getenv("GOOGLE_REDIRECT_URL")
	if redirect == "" {
		redirect = "http://localhost:8080/api/v1/auth/google/callback"
	}
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirect,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

func generateState() string {
	b := make([]byte, 16)
	if _, err := crand.Read(b); err != nil {
		return fmt.Sprintf("st-%d", mrand.Int63())
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func googleLoginHandler(c *gin.Context) {
	conf := oauthConfigFromEnv()
	state := generateState()
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "oauthstate",
		Value:    state,
		Path:     "/",
		Expires:  time.Now().Add(5 * time.Minute),
		HttpOnly: true,
	})
	url := conf.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusFound, url)
}

// googleCallbackHandler exchanges code for userinfo, creates user if needed, and sets JWT cookie.
func googleCallbackHandler(c *gin.Context, db *sql.DB) {
	state := c.Query("state")
	if state == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing state"})
		return
	}
	cookie, err := c.Request.Cookie("oauthstate")
	if err != nil || cookie.Value != state {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid state"})
		return
	}

	code := c.Query("code")
	if code == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing code"})
		return
	}

	conf := oauthConfigFromEnv()
	token, err := conf.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("oauth exchange error: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "token exchange failed"})
		return
	}

	client := conf.Client(context.Background(), token)
	resp, err := client.Get("https://openidconnect.googleapis.com/v1/userinfo")
	if err != nil {
		log.Printf("userinfo request failed: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to get userinfo"})
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("userinfo status %d: %s", resp.StatusCode, string(body))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "userinfo error"})
		return
	}

	var ui struct {
		Sub     string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ui); err != nil {
		log.Printf("decode userinfo: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "invalid userinfo"})
		return
	}

	if ui.Sub == "" {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "missing user id"})
		return
	}

	userID, username, err := findOrCreateUser(db, ui.Sub, ui.Name, ui.Email)
	if err != nil {
		log.Printf("create user: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "user create failed"})
		return
	}

	tokenStr, err := createJWT(userID, username)
	if err != nil {
		log.Printf("create jwt: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "jwt failed"})
		return
	}

	// set cookie for 30 days
	maxAge := 30 * 24 * 60 * 60
	c.SetCookie("qmanga_jwt", tokenStr, maxAge, "/", "", false, true)
	c.Redirect(http.StatusFound, "/")
}

func findOrCreateUser(db *sql.DB, googleID, name, email string) (string, string, error) {
	// try existing
	var id string
	var username string
	err := db.QueryRow("SELECT id, username FROM users WHERE google_id = $1", googleID).Scan(&id, &username)
	if err == nil {
		return id, username, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", "", fmt.Errorf("query user: %w", err)
	}

	// create new user
	uname := sanitizeUsername(name)
	// ensure unique
	base := uname
	for i := 0; i < 5; i++ {
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", uname).Scan(&exists)
		if err != nil {
			return "", "", fmt.Errorf("check username: %w", err)
		}
		if !exists {
			break
		}
		uname = fmt.Sprintf("%s_%04d", base, mrand.Intn(10000))
	}

	err = db.QueryRow("INSERT INTO users (username, email, google_id) VALUES ($1, $2, $3) RETURNING id", uname, email, googleID).Scan(&id)
	if err != nil {
		return "", "", fmt.Errorf("insert user: %w", err)
	}
	return id, uname, nil
}

var usernameSanitizeRE = regexp.MustCompile(`[^a-z0-9_]+`)

func sanitizeUsername(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "_")
	s = usernameSanitizeRE.ReplaceAllString(s, "")
	if s == "" {
		s = fmt.Sprintf("user%d", time.Now().Unix()%10000)
	}
	if len(s) > 30 {
		s = s[:30]
	}
	return s
}

func createJWT(userID, username string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret"
	}
	claims := jwt.MapClaims{
		"sub":      userID,
		"username": username,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(30 * 24 * time.Hour).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

// AuthMiddleware validates JWT from cookie and sets user_id in context.
func AuthMiddleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, err := c.Cookie("qmanga_jwt")
		if err != nil || tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
			return
		}
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "dev-secret"
		}
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			// ensure HMAC
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
			return
		}
		sub, ok := claims["sub"].(string)
		if !ok || sub == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid subject"})
			return
		}
		// optional: load user to ensure exists
		var exists bool
		if err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", sub).Scan(&exists); err != nil || !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}
		c.Set("user_id", sub)
		c.Next()
	}
}

func userMeHandler(c *gin.Context, db *sql.DB) {
	uid, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user id missing in context"})
		return
	}
	idStr, _ := uid.(string)
	var username, email, role string
	var createdAt time.Time
	err := db.QueryRow("SELECT username, email, role, created_at FROM users WHERE id = $1", idStr).Scan(&username, &email, &role, &createdAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": idStr, "username": username, "email": email, "role": role, "created_at": createdAt})
}
