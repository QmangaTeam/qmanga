package http

import (
	"github.com/gin-gonic/gin"
	"qmanga/internal/identity/application"
	"qmanga/internal/identity/domain"
	"strings"
)

func apiTokenRouteAllowed(method, path string) bool {
	switch method + " " + path {
	case "POST /api/v1/studio/revisions/:id/submit", "PUT /api/v1/studio/projects/:id/cover", "PUT /api/v1/studio/projects/:id/cover-file", "GET /api/v1/auth/me", "GET /api/v1/studio/projects", "GET /api/v1/studio/projects/:id", "GET /api/v1/studio/projects/:id/chapters", "GET /api/v1/studio/projects/:id/revisions", "POST /api/v1/studio/projects/:id/chapters", "POST /api/v1/studio/projects/:id/revisions", "GET /api/v1/studio/revisions/:id", "GET /api/v1/studio/revisions/:id/manifest", "PUT /api/v1/studio/revisions/:id/pages", "PUT /api/v1/studio/revisions/:id/order", "PUT /api/v1/studio/revisions/:id/attribution", "DELETE /api/v1/studio/revisions/:id/pages/:page_id", "DELETE /api/v1/studio/revisions/:id", "POST /api/v1/studio/revisions/:id/pages/:page_id/retry", "POST /api/v1/studio/revisions/:id/publish",
		"GET /api/v1/admin/titles", "GET /api/v1/admin/titles/:id", "GET /api/v1/admin/projects/:id", "GET /api/v1/admin/catalog-teams", "GET /api/v1/admin/taxonomy", "POST /api/v1/admin/titles", "PUT /api/v1/admin/titles/:id", "POST /api/v1/admin/titles/:id/chapters", "POST /api/v1/admin/titles/:id/projects", "PUT /api/v1/admin/titles/:id/categories":
		return true
	}
	return false
}
func DirectPublishToken(c *gin.Context) (string, bool) {
	v, exists := c.Get("api_token")
	t, ok := v.(application.APIToken)
	return t.ID, exists && ok && t.DirectPublish && t.ID != ""
}
func (h *IdentityHandler) authenticateBearer(c *gin.Context) bool {
	values := c.Request.Header.Values("Authorization")
	if len(values) != 1 {
		Error(c, 401, "unauthorized", "API token required")
		return false
	}
	parts := strings.Split(values[0], " ")
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		Error(c, 401, "unauthorized", "API token required")
		return false
	}
	t, a, e := h.service.AuthenticateAPIToken(c.Request.Context(), parts[1])
	if e != nil {
		Error(c, 401, "unauthorized", "Invalid API token")
		return false
	}
	if !apiTokenRouteAllowed(c.Request.Method, c.FullPath()) {
		Error(c, 403, "token_scope", "API token scope denied")
		return false
	}
	if expected := c.GetHeader("X-Qmanga-Account"); expected != "" && expected != a.ID {
		Error(c, 403, "account_changed", "API token account mismatch")
		return false
	}
	c.Set("principal", principal{account: a})
	c.Set("api_token", t)
	c.Request.Header.Set("X-Qmanga-Account", a.ID)
	c.Header("Cache-Control", "private, no-store")
	c.Next()
	return true
}
func (h *IdentityHandler) RegisterAPITokens(api *gin.RouterGroup) {
	group := api.Group("/admin/api-tokens", h.RequireUser(), expectedAccount(), RequireRole("admin"))
	group.GET("", h.listAPITokens)
	group.POST("", h.createAPIToken)
	group.DELETE("/:id", h.revokeAPIToken)
}
func (h *IdentityHandler) listAPITokens(c *gin.Context) {
	a, _ := CurrentAccount(c)
	items, e := h.service.ListAPITokens(c.Request.Context(), a.ID)
	if e != nil {
		h.respondError(c, e)
		return
	}
	c.JSON(200, gin.H{"items": items})
}
func (h *IdentityHandler) createAPIToken(c *gin.Context) {
	var in struct {
		Name   string `json:"name"`
		Days   int    `json:"expires_in_days"`
		Direct bool   `json:"direct_publish"`
	}
	if !decode(c, &in) {
		return
	}
	a, _ := CurrentAccount(c)
	raw, item, e := h.service.CreateAPIToken(c.Request.Context(), a.ID, in.Name, in.Days, in.Direct)
	if e != nil {
		if e == domain.ErrForbidden {
			Error(c, 403, "forbidden", "Access denied")
		} else {
			h.respondError(c, e)
		}
		return
	}
	c.JSON(201, gin.H{"token": raw, "item": item})
}
func (h *IdentityHandler) revokeAPIToken(c *gin.Context) {
	a, _ := CurrentAccount(c)
	e := h.service.RevokeAPIToken(c.Request.Context(), a.ID, c.Param("id"))
	if e != nil {
		if e == domain.ErrNotFound {
			Error(c, 404, "not_found", "Token not found")
		} else {
			h.respondError(c, e)
		}
		return
	}
	c.Status(204)
}
