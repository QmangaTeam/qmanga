package http

import (
	"github.com/gin-gonic/gin"
	"qmanga/internal/identity/application"
)

func (handler *IdentityHandler) accounts(request *gin.Context) {
	account, _ := CurrentAccount(request)
	items, failure := handler.service.Accounts(request.Request.Context(), account.ID, request.Query("q"), request.Query("after"))
	if failure != nil {
		handler.respondError(request, failure)
		return
	}
	result := []gin.H{}
	for _, item := range items {
		result = append(result, gin.H{"id": item.ID, "name": item.Name, "role": item.Role, "blocked": item.Blocked, "version": item.Version})
	}
	request.JSON(200, gin.H{"items": result, "has_more": len(items) == 50})
}
func (handler *IdentityHandler) manageAccount(request *gin.Context) {
	var input struct {
		Role    string `json:"role"`
		Reason  string `json:"reason"`
		Blocked bool   `json:"blocked"`
		Version int64  `json:"version"`
	}
	if !decode(request, &input) {
		return
	}
	account, _ := CurrentAccount(request)
	if failure := handler.service.ManageAccount(request.Request.Context(), account.ID, request.Param("id"), application.AccountChange{Role: input.Role, Reason: input.Reason, Blocked: input.Blocked, Version: input.Version}); failure != nil {
		handler.respondError(request, failure)
		return
	}
	request.Status(204)
}
func (handler *IdentityHandler) audit(request *gin.Context) {
	request.Header("Cache-Control", "no-store")
	account, _ := CurrentAccount(request)
	items, failure := handler.service.Audit(request.Request.Context(), account.ID, request.Query("after"), application.AuditFilter{Actor: request.Query("actor"), Object: request.Query("object"), Action: request.Query("action"), From: request.Query("from"), To: request.Query("to")})
	if failure != nil {
		handler.respondError(request, failure)
		return
	}
	result := []gin.H{}
	for _, item := range items {
		result = append(result, gin.H{"id": item.ID, "actor": item.Actor, "action": item.Action, "entity_type": item.EntityType, "entity_id": item.EntityID, "reason": item.Reason, "created_at": item.CreatedAt, "actor_name": item.ActorName, "target_name": item.TargetName, "target_path": item.TargetPath, "before": item.Before, "after": item.After})
	}
	request.JSON(200, gin.H{"items": result, "has_more": len(items) == 50})
}
