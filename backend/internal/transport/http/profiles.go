package http

import (
	"errors"
	"github.com/gin-gonic/gin"
	"qmanga/internal/identity/application"
	"qmanga/internal/identity/domain"
)

type ProfileHandler struct {
	service  *application.ProfileService
	identity *IdentityHandler
}

func NewProfileHandler(service *application.ProfileService, identity *IdentityHandler) *ProfileHandler {
	return &ProfileHandler{service, identity}
}
func (h *ProfileHandler) Register(api *gin.RouterGroup) {
	api.GET("/account/preferences", h.identity.RequireUser(), expectedAccount(), h.preferences)
	api.PUT("/account/preferences", h.identity.RequireUser(), expectedAccount(), h.savePreferences)
	api.GET("/profiles/:id", h.optionalUser(), h.profile)
	api.PUT("/profiles/:id/follow", h.identity.RequireUser(), expectedAccount(), h.follow)
}
func (h *ProfileHandler) optionalUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "private, no-store")
		c.Header("Vary", "Cookie")
		raw, err := c.Cookie("qmanga_session")
		if err == nil && raw != "" {
			session, account, failure := h.identity.service.Authenticate(c.Request.Context(), raw)
			if failure == nil {
				c.Set("principal", principal{account: account, session: session})
			} else if !errors.Is(failure, domain.ErrCredentials) {
				h.identity.respondError(c, failure)
				return
			}
		}
		c.Next()
	}
}

type preferencesDTO struct {
	Version         *int64  `json:"version"`
	Bio             *string `json:"bio"`
	PublicProfile   *bool   `json:"public_profile"`
	ShowLibrary     *bool   `json:"show_library"`
	ShowActivity    *bool   `json:"show_activity"`
	AllowFollow     *bool   `json:"allow_follow"`
	NotifyReleases  *bool   `json:"notify_releases"`
	NotifyReplies   *bool   `json:"notify_replies"`
	NotifyFollowers *bool   `json:"notify_followers"`
	EmailDigest     *bool   `json:"email_digest"`
}

func preferenceView(p application.Preferences) gin.H {
	return gin.H{"version": p.Version, "bio": p.Bio, "public_profile": p.PublicProfile, "show_library": p.ShowLibrary, "show_activity": p.ShowActivity, "allow_follow": p.AllowFollow, "notify_releases": p.NotifyReleases, "notify_replies": p.NotifyReplies, "notify_followers": p.NotifyFollowers, "email_digest": p.EmailDigest}
}
func (h *ProfileHandler) preferences(c *gin.Context) {
	c.Header("Cache-Control", "private, no-store")
	a, _ := CurrentAccount(c)
	p, err := h.service.Preferences(c.Request.Context(), a.ID)
	if err != nil {
		h.failure(c, err)
		return
	}
	c.JSON(200, preferenceView(p))
}
func (h *ProfileHandler) savePreferences(c *gin.Context) {
	c.Header("Cache-Control", "private, no-store")
	var input preferencesDTO
	if !decode(c, &input) {
		return
	}
	if input.Version == nil || input.Bio == nil || input.PublicProfile == nil || input.ShowLibrary == nil || input.ShowActivity == nil || input.AllowFollow == nil || input.NotifyReleases == nil || input.NotifyReplies == nil || input.NotifyFollowers == nil || input.EmailDigest == nil {
		h.failure(c, domain.ErrInvalid)
		return
	}
	a, _ := CurrentAccount(c)
	p, err := h.service.SavePreferences(c.Request.Context(), a.ID, application.Preferences{Version: *input.Version, Bio: *input.Bio, PublicProfile: *input.PublicProfile, ShowLibrary: *input.ShowLibrary, ShowActivity: *input.ShowActivity, AllowFollow: *input.AllowFollow, NotifyReleases: *input.NotifyReleases, NotifyReplies: *input.NotifyReplies, NotifyFollowers: *input.NotifyFollowers, EmailDigest: *input.EmailDigest})
	if err != nil {
		h.failure(c, err)
		return
	}
	c.JSON(200, preferenceView(p))
}
func profileView(p application.PublicProfile) gin.H {
	library := []gin.H{}
	for _, i := range p.Library {
		library = append(library, gin.H{"title_id": i.TitleID, "slug": i.Slug, "name": i.Name, "cover_url": i.CoverURL, "status": i.Status})
	}
	activity := []gin.H{}
	for _, i := range p.Activity {
		activity = append(activity, gin.H{"label": i.Label, "href": i.Href, "created_at": i.CreatedAt})
	}
	return gin.H{"id": p.ID, "display_name": p.DisplayName, "bio": p.Bio, "avatar_url": p.AvatarURL, "created_at": p.CreatedAt, "public_profile": p.PublicProfile, "show_library": p.ShowLibrary, "show_activity": p.ShowActivity, "allow_follow": p.AllowFollow, "is_owner": p.IsOwner, "is_following": p.IsFollowing, "followers_count": p.FollowersCount, "following_count": p.FollowingCount, "library": library, "activity": activity}
}
func (h *ProfileHandler) profile(c *gin.Context) {
	c.Header("Cache-Control", "private, no-store")
	a, _ := CurrentAccount(c)
	p, err := h.service.Profile(c.Request.Context(), a.ID, c.Param("id"))
	if err != nil {
		h.failure(c, err)
		return
	}
	c.JSON(200, profileView(p))
}
func (h *ProfileHandler) follow(c *gin.Context) {
	c.Header("Cache-Control", "private, no-store")
	var input struct {
		Following *bool `json:"following"`
	}
	if !decode(c, &input) {
		return
	}
	if input.Following == nil {
		h.failure(c, domain.ErrInvalid)
		return
	}
	a, _ := CurrentAccount(c)
	if err := h.service.Follow(c.Request.Context(), a.ID, c.Param("id"), *input.Following); err != nil {
		h.failure(c, err)
		return
	}
	p, err := h.service.Profile(c.Request.Context(), a.ID, c.Param("id"))
	if errors.Is(err, domain.ErrNotFound) {
		c.JSON(200, gin.H{"following": false})
		return
	}
	if err != nil {
		h.failure(c, err)
		return
	}
	c.JSON(200, profileView(p))
}
func (h *ProfileHandler) failure(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		Error(c, 404, "not_found", "Профиль табылмады")
	case errors.Is(err, domain.ErrForbidden):
		Error(c, 403, "forbidden", "Рұқсат жеткіліксіз")
	case errors.Is(err, domain.ErrInvalid):
		Error(c, 422, "validation", "Профиль деректерін тексеріңіз")
	default:
		h.identity.respondError(c, err)
	}
}
