package http

import (
	"github.com/gin-gonic/gin"
	"qmanga/internal/identity/application"
	"time"
)

const registrationImportCookie = "qmanga_guest_registration"

func (handler *IdentityHandler) registrationImportToken(request *gin.Context) (string, error) {
	existing, _ := request.Cookie(registrationImportCookie)
	return handler.service.RegistrationImportToken(existing)
}
func (handler *IdentityHandler) rememberRegistrationImport(request *gin.Context, token string) {
	handler.cookie(request, registrationImportCookie, token, time.Now().Add(application.RegistrationImportLifetime))
}
func (handler *IdentityHandler) guestImport(request *gin.Context) {
	request.Header("Cache-Control", "private, no-store")
	account, _ := CurrentAccount(request)
	token, _ := request.Cookie(registrationImportCookie)
	allowed, err := handler.service.RegistrationImportAllowed(request.Request.Context(), account.ID, token)
	if err != nil {
		handler.respondError(request, err)
		return
	}
	request.JSON(200, gin.H{"allowed": allowed})
}
func (handler *IdentityHandler) finishGuestImport(request *gin.Context) {
	request.Header("Cache-Control", "private, no-store")
	account, _ := CurrentAccount(request)
	token, _ := request.Cookie(registrationImportCookie)
	if err := handler.service.CompleteRegistrationImport(request.Request.Context(), account.ID, token); err != nil {
		handler.respondError(request, err)
		return
	}
	request.Status(204)
}
