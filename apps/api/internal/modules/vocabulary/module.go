// Package vocabulary registers routes over dependencies supplied by composition.
package vocabulary

import (
	"engflex-api/internal/modules/vocabulary/controllers"
	"engflex-api/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, categoryCtl *controllers.VocabularyCategoryController, deckCtl *controllers.VocabularyDeckController, itemCtl *controllers.VocabularyDeckItemController, dictCtl *controllers.VocabularyItemController, stateCtl *controllers.UserVocabularyController) {
	c := rg.Group("/vocabulary-categories")
	c.GET("", middleware.RequireAuth(), categoryCtl.List)
	c.GET("/:id", middleware.RequireAuth(), categoryCtl.GetByID)
	// Categories are shared taxonomy, not owned content, so admin-only
	// (unlike decks/items, there's no "owner" to fall back to).
	c.POST("", middleware.RequireAuth(), middleware.RequireAdmin(), categoryCtl.Create)
	c.PUT("/:id", middleware.RequireAuth(), middleware.RequireAdmin(), categoryCtl.Update)
	c.DELETE("/:id", middleware.RequireAuth(), middleware.RequireAdmin(), categoryCtl.Delete)

	d := rg.Group("/vocabulary-decks")
	d.GET("", middleware.RequireAuth(), deckCtl.List)
	d.GET("/:id", middleware.RequireAuth(), deckCtl.GetByID)
	d.POST("", middleware.RequireAuth(), deckCtl.Create)
	// Update/Delete serve both owners (their own deck) and admins (system
	// decks) — the service enforces which, so no RequireAdmin here.
	d.PUT("/:id", middleware.RequireAuth(), deckCtl.Update)
	d.DELETE("/:id", middleware.RequireAuth(), deckCtl.Delete)

	i := rg.Group("/vocabulary-deck-items")
	i.GET("", middleware.RequireAuth(), itemCtl.List)
	i.GET("/:id", middleware.RequireAuth(), itemCtl.GetByID)
	i.POST("", middleware.RequireAuth(), itemCtl.Create)
	i.PUT("/:id", middleware.RequireAuth(), itemCtl.Update)
	i.DELETE("/:id", middleware.RequireAuth(), itemCtl.Delete)

	v := rg.Group("/vocabulary-items")
	v.GET("", middleware.RequireAuth(), dictCtl.List)
	v.GET("/:id", middleware.RequireAuth(), dictCtl.GetByID)
	v.POST("", middleware.RequireAuth(), dictCtl.Create)
	// Update/Delete only ever touch admin-curated entries (service-enforced:
	// CreatedByUserID must be nil), so these are admin-gated outright.
	v.PUT("/:id", middleware.RequireAuth(), middleware.RequireAdmin(), dictCtl.Update)
	v.DELETE("/:id", middleware.RequireAuth(), middleware.RequireAdmin(), dictCtl.Delete)

	u := rg.Group("/user-vocabulary")
	u.GET("", middleware.RequireAuth(), stateCtl.List)
	u.POST("", middleware.RequireAuth(), stateCtl.Save)
	u.DELETE("/:itemId", middleware.RequireAuth(), stateCtl.Unsave)

	admin := rg.Group("/admin")
	admin.POST("/vocabulary-decks", middleware.RequireAuth(), middleware.RequireAdmin(), deckCtl.CreateSystemDeck)
	admin.POST("/vocabulary-items", middleware.RequireAuth(), middleware.RequireAdmin(), dictCtl.CreateSystemEntry)
}
