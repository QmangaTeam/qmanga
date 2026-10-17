package http

import (
	"errors"
	"log"

	"github.com/gin-gonic/gin"
	"qmanga/internal/teams/application"
	"qmanga/internal/teams/domain"
)

type TeamHandler struct{ service *application.Service }
type teamDTO struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Approved    bool   `json:"approved"`
	Role        string `json:"my_role,omitempty"`
}

func viewTeam(team domain.Team, role domain.Role) teamDTO {
	return teamDTO{team.ID, team.Slug, team.Name, team.Description, team.Approved, string(role)}
}
func NewTeamHandler(service *application.Service) *TeamHandler { return &TeamHandler{service: service} }
func (handler *TeamHandler) Register(api *gin.RouterGroup, identity *IdentityHandler) {
	studio := api.Group("/studio")
	studio.Use(identity.RequireUser())
	studio.GET("/teams", handler.list)
	studio.POST("/teams", handler.create)
	studio.GET("/teams/:id/members", handler.members)
	studio.POST("/teams/:id/invitations", handler.invite)
	studio.DELETE("/teams/:id/members/:user_id", handler.remove)
	api.POST("/teams/invitations/accept", identity.RequireUser(), handler.accept)
	admin := api.Group("/admin")
	admin.Use(identity.RequireUser(), RequireRole("admin"))
	admin.GET("/teams/pending", handler.pending)
	admin.POST("/teams/:id/approve", handler.approve)
}
func (handler *TeamHandler) list(request *gin.Context) {
	account, _ := CurrentAccount(request)
	items, failure := handler.service.List(request.Request.Context(), account.ID)
	if failure != nil {
		teamError(request, failure)
		return
	}
	views := make([]teamDTO, 0, len(items))
	for _, item := range items {
		views = append(views, viewTeam(item.Team, item.Role))
	}
	request.JSON(200, gin.H{"items": views, "next_cursor": nil})
}
func (handler *TeamHandler) create(request *gin.Context) {
	var input struct {
		Slug        string `json:"slug"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if !decode(request, &input) {
		return
	}
	account, _ := CurrentAccount(request)
	team, failure := handler.service.Create(request.Request.Context(), account.ID, input.Slug, input.Name, input.Description)
	if failure != nil {
		teamError(request, failure)
		return
	}
	request.JSON(201, viewTeam(team, domain.Owner))
}
func (handler *TeamHandler) pending(request *gin.Context) {
	account, _ := CurrentAccount(request)
	items, failure := handler.service.Pending(request.Request.Context(), account.ID, requestOffset(request))
	if failure != nil {
		teamError(request, failure)
		return
	}
	views := make([]teamDTO, 0, len(items))
	for _, team := range items {
		views = append(views, viewTeam(team, ""))
	}
	request.JSON(200, gin.H{"items": views, "next_cursor": nil})
}
func (handler *TeamHandler) approve(request *gin.Context) {
	account, _ := CurrentAccount(request)
	if failure := handler.service.Approve(request.Request.Context(), account.ID, request.Param("id")); failure != nil {
		teamError(request, failure)
		return
	}
	request.Status(204)
}
func (handler *TeamHandler) invite(request *gin.Context) {
	var input struct {
		Email string      `json:"email"`
		Role  domain.Role `json:"role"`
	}
	if !decode(request, &input) {
		return
	}
	account, _ := CurrentAccount(request)
	if failure := handler.service.Invite(request.Request.Context(), account.ID, request.Param("id"), input.Email, input.Role); failure != nil {
		teamError(request, failure)
		return
	}
	request.JSON(202, gin.H{"message": "Шақыру поштаға жіберілді"})
}
func (handler *TeamHandler) accept(request *gin.Context) {
	var input struct {
		Token string `json:"token"`
	}
	if !decode(request, &input) {
		return
	}
	account, _ := CurrentAccount(request)
	teamID, failure := handler.service.Accept(request.Request.Context(), account.ID, input.Token)
	if failure != nil {
		teamError(request, failure)
		return
	}
	request.JSON(200, gin.H{"team_id": teamID, "message": "Командаға қосылдыңыз"})
}
func (handler *TeamHandler) members(request *gin.Context) {
	account, _ := CurrentAccount(request)
	items, failure := handler.service.Members(request.Request.Context(), account.ID, request.Param("id"))
	if failure != nil {
		teamError(request, failure)
		return
	}
	views := make([]gin.H, 0, len(items))
	for _, member := range items {
		views = append(views, gin.H{"user_id": member.UserID, "email": member.Email, "role": member.Role})
	}
	request.JSON(200, gin.H{"items": views, "next_cursor": nil})
}
func (handler *TeamHandler) remove(request *gin.Context) {
	account, _ := CurrentAccount(request)
	if failure := handler.service.Remove(request.Request.Context(), account.ID, request.Param("id"), request.Param("user_id")); failure != nil {
		teamError(request, failure)
		return
	}
	request.Status(204)
}
func teamError(request *gin.Context, failure error) {
	switch {
	case errors.Is(failure, domain.ErrForbidden):
		Error(request, 403, "forbidden", "Бұл әрекетке рұқсат жоқ")
	case errors.Is(failure, domain.ErrNotFound):
		Error(request, 404, "not_found", "Команда табылмады")
	case errors.Is(failure, domain.ErrExpired):
		Error(request, 410, "expired", "Шақыру мерзімі аяқталған немесе бұрын қабылданған")
	case errors.Is(failure, domain.ErrConflict):
		Error(request, 409, "conflict", "Команда атауы бос емес немесе иесін жоюға болмайды")
	case errors.Is(failure, domain.ErrInvalid):
		Error(request, 422, "validation", "Команда деректерін тексеріңіз")
	default:
		log.Printf("team failure: %T", failure)
		Error(request, 503, "unavailable", "Қызмет уақытша қолжетімсіз")
	}
}
