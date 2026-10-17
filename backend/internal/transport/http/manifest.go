package http

import (
	"github.com/gin-gonic/gin"
	"qmanga/internal/publishing/application"
)

func variantView(item application.Variant) gin.H {
	return gin.H{"id": item.ID, "url": coverURL(item.ID), "width": item.Width, "height": item.Height, "bytes": item.Bytes, "offset": item.Offset, "mime": item.Mime, "checksum": item.Checksum}
}
func manifestView(item application.Manifest) gin.H {
	pages := []gin.H{}
	for _, page := range item.Pages {
		variants := map[string][]gin.H{"eco": {}, "standard": {}, "high": {}}
		for tier, parts := range page.Variants {
			for _, part := range parts {
				variants[tier] = append(variants[tier], variantView(part))
			}
		}
		pages = append(pages, gin.H{"id": page.ID, "position": page.Position, "width": page.Width, "height": page.Height, "variants": variants})
	}
	return gin.H{"release_id": item.ReleaseID, "revision_id": item.RevisionID, "title_id": item.TitleID, "title_slug": item.TitleSlug, "title_name": item.TitleName, "chapter_label": item.ChapterLabel, "chapter_order": item.ChapterOrder, "project_id": item.ProjectID, "kind": item.Kind, "team": gin.H{"id": item.TeamID, "name": item.TeamName}, "direction": item.Direction, "previous_release_id": item.PreviousID, "next_release_id": item.NextID, "pages": pages}
}
func (handler *PublishingHandler) manifest(request *gin.Context) {
	item, failure := handler.service.Manifest(request.Request.Context(), handler.actor(request), request.Param("id"), "")
	if failure != nil {
		publicationError(request, failure)
		return
	}
	request.Header("Cache-Control", "public, max-age=0, must-revalidate")
	request.JSON(200, manifestView(item))
}
func (handler *PublishingHandler) preview(request *gin.Context) {
	draft, failure := handler.service.Revision(request.Request.Context(), handler.actor(request), request.Param("id"))
	if failure != nil {
		publicationError(request, failure)
		return
	}
	item, failure := handler.service.Manifest(request.Request.Context(), handler.actor(request), draft.ReleaseID, draft.ID)
	if failure != nil {
		publicationError(request, failure)
		return
	}
	request.Header("Cache-Control", "private, no-store")
	request.JSON(200, manifestView(item))
}
func (handler *PublishingHandler) original(request *gin.Context) {
	item, failure := handler.service.Original(request.Request.Context(), handler.actor(request), request.Param("id"), request.Param("page_id"))
	if failure != nil {
		publicationError(request, failure)
		return
	}
	request.Header("Cache-Control", "private, no-store")
	request.JSON(200, variantView(item))
}
