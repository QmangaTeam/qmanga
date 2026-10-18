package http

import "github.com/gin-gonic/gin"

func (handler *TeamHandler) invitations(request *gin.Context) {
	account, _ := CurrentAccount(request)
	items, failure := handler.service.Invitations(request.Request.Context(), account.ID, request.Param("id"), requestOffset(request))
	if failure != nil {
		teamError(request, failure)
		return
	}
	views := make([]gin.H, 0, len(items))
	for _, item := range items {
		views = append(views, gin.H{"id": item.ID, "email": item.Email, "role": item.Role, "status": item.Status, "expires_at": item.ExpiresAt})
	}
	request.JSON(200, gin.H{"items": views})
}
func (handler *TeamHandler) revokeInvitation(request *gin.Context) {
	account, _ := CurrentAccount(request)
	if failure := handler.service.RevokeInvitation(request.Request.Context(), account.ID, request.Param("id"), request.Param("invitation_id")); failure != nil {
		teamError(request, failure)
		return
	}
	request.Status(204)
}
