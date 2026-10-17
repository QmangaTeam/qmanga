package http

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"qmanga/internal/identity/application"
	"qmanga/internal/identity/domain"
)

type IdentityHandler struct {
	service *application.Service
	origin  string
	secure  bool
	limiter *rateLimiter
}
type userDTO struct {
	ID            string `json:"id"`
	DisplayName   string `json:"display_name"`
	Email         string `json:"email"`
	Role          string `json:"role"`
	EmailVerified bool   `json:"email_verified"`
}
type principal struct {
	account domain.Account
	session application.Session
}

func userView(account domain.Account) userDTO {
	return userDTO{ID: account.ID, Email: account.Email, Role: account.Role, EmailVerified: account.Verified, DisplayName: account.DisplayName}
}
func NewIdentityHandler(service *application.Service, origin string, secure bool) *IdentityHandler {
	return &IdentityHandler{service: service, origin: origin, secure: secure, limiter: &rateLimiter{entries: make(map[string]rateEntry)}}
}
func (handler *IdentityHandler) Register(api *gin.RouterGroup) {
	staff := api.Group("/moderation/accounts", handler.RequireUser(), expectedAccount(), RequireRole("admin", "moderator"))
	staff.GET("", handler.accounts)
	staff.PUT("/:id", handler.manageAccount)
	api.GET("/admin/audit", handler.RequireUser(), expectedAccount(), RequireRole("admin"), handler.audit)
	auth := api.Group("/auth")
	auth.Use(handler.CheckOrigin())
	auth.POST("/register", handler.limit(), handler.register)
	auth.POST("/login", handler.limit(), handler.login)
	auth.POST("/forgot-password", handler.limit(), handler.forgotPassword)
	auth.POST("/resend-verification", handler.limit(), handler.resendVerification)
	auth.POST("/reset-password", handler.limit(), handler.resetPassword)
	auth.POST("/verify-email", handler.limit(), handler.verifyEmail)
	auth.GET("/me", handler.RequireUser(), handler.me)
	auth.POST("/profile", handler.RequireUser(), handler.updateProfile)
	auth.GET("/csrf", handler.RequireUser(), handler.csrf)
	auth.POST("/logout", handler.RequireUser(), handler.logout)
}
func decode(request *gin.Context, destination any) bool {
	request.Request.Body = http.MaxBytesReader(request.Writer, request.Request.Body, 64<<10)
	decoder := json.NewDecoder(request.Request.Body)
	decoder.DisallowUnknownFields()
	if failure := decoder.Decode(destination); failure != nil {
		Error(request, 400, "invalid_json", "Сұрау пішімі дұрыс емес")
		return false
	}
	if failure := decoder.Decode(new(any)); failure != io.EOF {
		Error(request, 400, "invalid_json", "Сұрау пішімі дұрыс емес")
		return false
	}
	return true
}
func (handler *IdentityHandler) CheckOrigin() gin.HandlerFunc {
	return func(request *gin.Context) {
		if request.Request.Method != "GET" && request.Request.Method != "HEAD" && request.GetHeader("Origin") != handler.origin {
			Error(request, 403, "origin", "Сұрау көзіне рұқсат жоқ")
			return
		}
		request.Next()
	}
}
func (handler *IdentityHandler) RequireUser() gin.HandlerFunc {
	return func(request *gin.Context) {
		raw, failure := request.Cookie("qmanga_session")
		if failure != nil {
			Error(request, 401, "unauthorized", "Алдымен кіріңіз")
			return
		}
		session, account, failure := handler.service.Authenticate(request.Request.Context(), raw)
		if failure != nil {
			if !errors.Is(failure, domain.ErrCredentials) {
				handler.respondError(request, failure)
				return
			}
			Error(request, 401, "unauthorized", "Қайта кіріңіз")
			return
		}
		if request.Request.Method != "GET" && request.Request.Method != "HEAD" {
			if request.GetHeader("Origin") != handler.origin || !handler.service.ValidCSRF(session, request.GetHeader("X-CSRF-Token")) {
				Error(request, 403, "csrf", "Сұрауды растау мүмкін болмады")
				return
			}
		}
		request.Set("principal", principal{account: account, session: session})
		request.Next()
	}
}
func CurrentAccount(request *gin.Context) (domain.Account, bool) {
	value, exists := request.Get("principal")
	if !exists {
		return domain.Account{}, false
	}
	identity, valid := value.(principal)
	return identity.account, valid
}
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(request *gin.Context) {
		account, exists := CurrentAccount(request)
		if exists {
			for _, role := range roles {
				if account.Role == role && account.Verified {
					request.Next()
					return
				}
			}
		}
		Error(request, 403, "forbidden", "Рұқсат жеткіліксіз")
	}
}
func (handler *IdentityHandler) register(request *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decode(request, &input) {
		return
	}
	if failure := handler.service.Register(request.Request.Context(), input.Email, input.Password); failure != nil {
		handler.respondError(request, failure)
		return
	}
	request.JSON(202, gin.H{"message": "Поштаңызды тексеріңіз. Растау сілтемесі жіберілді."})
}
func (handler *IdentityHandler) login(request *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decode(request, &input) {
		return
	}
	result, failure := handler.service.Login(request.Request.Context(), input.Email, input.Password)
	if failure != nil {
		handler.respondError(request, failure)
		return
	}
	handler.cookie(request, "qmanga_session", result.SessionToken, result.ExpiresAt)
	handler.cookie(request, "qmanga_csrf", result.CSRFToken, result.ExpiresAt)
	request.JSON(200, gin.H{"user": userView(result.Account)})
}
func (handler *IdentityHandler) cookie(request *gin.Context, name, value string, expiration time.Time) {
	maxAge := int(time.Until(expiration).Seconds())
	if value == "" {
		maxAge = -1
	}
	http.SetCookie(request.Writer, &http.Cookie{Name: name, Value: value, Path: "/", Expires: expiration, MaxAge: maxAge, HttpOnly: true, Secure: handler.secure, SameSite: http.SameSiteLaxMode})
}
func (handler *IdentityHandler) me(request *gin.Context) {
	account, _ := CurrentAccount(request)
	request.JSON(200, gin.H{"user": userView(account)})
}
func (handler *IdentityHandler) csrf(request *gin.Context) {
	value, _ := request.Get("principal")
	identity := value.(principal)
	token, failure := request.Cookie("qmanga_csrf")
	if failure != nil || !handler.service.ValidCSRF(identity.session, token) {
		Error(request, 403, "csrf", "Сессияны жаңарту үшін қайта кіріңіз")
		return
	}
	request.JSON(200, gin.H{"csrf_token": token})
}
func (handler *IdentityHandler) logout(request *gin.Context) {
	raw, _ := request.Cookie("qmanga_session")
	if failure := handler.service.Logout(request.Request.Context(), raw); failure != nil {
		handler.respondError(request, failure)
		return
	}
	handler.cookie(request, "qmanga_session", "", time.Unix(1, 0))
	handler.cookie(request, "qmanga_csrf", "", time.Unix(1, 0))
	request.Status(204)
}
func (handler *IdentityHandler) requestToken(request *gin.Context, kind string) {
	var input struct {
		Email string `json:"email"`
	}
	if !decode(request, &input) {
		return
	}
	if failure := handler.service.RequestToken(request.Request.Context(), input.Email, kind); failure != nil {
		handler.respondError(request, failure)
		return
	}
	request.JSON(202, gin.H{"message": "Егер мекенжай тіркелген болса, нұсқаулық поштаңызға жіберіледі."})
}
func (handler *IdentityHandler) forgotPassword(request *gin.Context) {
	handler.requestToken(request, "reset")
}
func (handler *IdentityHandler) resendVerification(request *gin.Context) {
	handler.requestToken(request, "verify")
}
func (handler *IdentityHandler) verifyEmail(request *gin.Context) {
	var input struct {
		Token string `json:"token"`
	}
	if !decode(request, &input) {
		return
	}
	if failure := handler.service.VerifyEmail(request.Request.Context(), input.Token); failure != nil {
		handler.respondError(request, failure)
		return
	}
	request.JSON(200, gin.H{"message": "Поштаңыз расталды. Енді кіре аласыз."})
}
func (handler *IdentityHandler) resetPassword(request *gin.Context) {
	var input struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if !decode(request, &input) {
		return
	}
	if failure := handler.service.ResetPassword(request.Request.Context(), input.Token, input.Password); failure != nil {
		handler.respondError(request, failure)
		return
	}
	handler.cookie(request, "qmanga_session", "", time.Unix(1, 0))
	handler.cookie(request, "qmanga_csrf", "", time.Unix(1, 0))
	request.JSON(200, gin.H{"message": "Құпиясөз жаңартылды. Қайта кіріңіз."})
}
func (handler *IdentityHandler) respondError(request *gin.Context, failure error) {
	switch {
	case errors.Is(failure, domain.ErrInvalid):
		Error(request, 422, "validation", "Пошта мен құпиясөзді тексеріңіз. Құпиясөз 10–128 таңба болуы керек.")
	case errors.Is(failure, domain.ErrCredentials):
		Error(request, 401, "credentials", "Пошта немесе құпиясөз дұрыс емес")
	case errors.Is(failure, domain.ErrExpired):
		Error(request, 410, "expired", "Сілтеме мерзімі аяқталған немесе бұрын қолданылған")
	case errors.Is(failure, domain.ErrConflict):
		Error(request, 409, "conflict", "Бұл әрекетті орындау мүмкін емес")
	default:
		log.Printf("identity failure: %T", failure)
		Error(request, 503, "unavailable", "Қызмет уақытша қолжетімсіз. Кейінірек қайталап көріңіз.")
	}
}

type rateEntry struct {
	count int
	until time.Time
}
type rateLimiter struct {
	mutex   sync.Mutex
	entries map[string]rateEntry
}

func (handler *IdentityHandler) limit() gin.HandlerFunc {
	return func(request *gin.Context) {
		handler.limiter.mutex.Lock()
		now := time.Now()
		key := request.ClientIP()
		entry := handler.limiter.entries[key]
		if now.After(entry.until) {
			entry = rateEntry{until: now.Add(time.Minute)}
		}
		entry.count++
		if len(handler.limiter.entries) > 4096 {
			for storedKey, stored := range handler.limiter.entries {
				if now.After(stored.until) {
					delete(handler.limiter.entries, storedKey)
				}
			}
		}
		allowed := entry.count <= 30 && len(handler.limiter.entries) <= 8192
		if allowed {
			handler.limiter.entries[key] = entry
		}
		handler.limiter.mutex.Unlock()
		if !allowed {
			request.Header("Retry-After", "60")
			Error(request, 429, "rate_limit", "Сәл күтіп, қайталап көріңіз")
			return
		}
		request.Next()
	}
}
func (handler *IdentityHandler) updateProfile(request *gin.Context) {
	var input struct {
		AccountID   string `json:"account_id"`
		DisplayName string `json:"display_name"`
	}
	if !decode(request, &input) {
		return
	}
	account, _ := CurrentAccount(request)
	if account.ID != input.AccountID {
		Error(request, 409, "account_changed", "Аккаунт өзгерді. Бетті жаңартыңыз")
		return
	}
	if failure := handler.service.UpdateProfile(request.Request.Context(), account.ID, input.DisplayName); failure != nil {
		handler.respondError(request, failure)
		return
	}
	request.JSON(200, gin.H{"message": "Профиль сақталды"})
}
