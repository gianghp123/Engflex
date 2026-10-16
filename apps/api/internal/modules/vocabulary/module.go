// Package vocabulary wires the vocabulary module: the shared dictionary,
// user flashcard decks and their rows, the deck taxonomy, and per-user word
// state. All five tables are served by this one module because they are one
// feature.
package vocabulary

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"engflex-api/internal/database/repositories"
	"engflex-api/internal/modules/vocabulary/controllers"
	"engflex-api/internal/modules/vocabulary/services"
	"engflex-api/internal/server/middleware"
)

// RegisterRoutes mounts the vocabulary endpoints under /api/v1.
func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	categorySvc := services.NewVocabularyCategoryService(repositories.NewVocabularyCategoryRepository(db))
	categoryCtl := controllers.NewVocabularyCategoryController(categorySvc)

	deckSvc := services.NewVocabularyDeckService(repositories.NewVocabularyDeckRepository(db))
	deckCtl := controllers.NewVocabularyDeckController(deckSvc)

	itemSvc := services.NewVocabularyDeckItemService(repositories.NewVocabularyDeckItemRepository(db), repositories.NewVocabularyDeckRepository(db))
	itemCtl := controllers.NewVocabularyDeckItemController(itemSvc)

	itemRepo := repositories.NewVocabularyItemRepository(db)
	dictSvc := services.NewVocabularyItemService(itemRepo)
	dictCtl := controllers.NewVocabularyItemController(dictSvc)

	stateSvc := services.NewUserVocabularyService(repositories.NewUserVocabularyRepository(db), itemRepo)
	stateCtl := controllers.NewUserVocabularyController(stateSvc)

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
