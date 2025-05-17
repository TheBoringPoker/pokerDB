package storage

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"pokerDB/pkg/models"
)

// TestSQLiteIntegration exercises creating a game and actions using SQLite.
func TestSQLiteIntegration(t *testing.T) {
	cfg := Config{Dialect: DialectSQLite, DSN: "file::memory:?cache=shared"}
	db, err := NewDB(cfg)
	if err != nil {
		if strings.Contains(err.Error(), "CGO_ENABLED=0") {
			t.Skip("sqlite driver requires CGO, skipping")
		}
		t.Fatalf("failed to open sqlite: %v", err)
	}

	if err := db.AutoMigrate(&models.Game{}, &models.Ledger{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	game := models.NewGame(uuid.New(), 2)
	if err := db.Create(game).Error; err != nil {
		t.Fatalf("create game: %v", err)
	}
	game.AddAction(uuid.New(), models.ActionCheck, 0)
	if err := db.Model(&game).Update("action_log", game.ActionLog).Error; err != nil {
		t.Fatalf("update log: %v", err)
	}

	var loaded models.Game
	if err := db.First(&loaded, "id = ?", game.ID).Error; err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(loaded.ActionLog) != 1 {
		t.Fatalf("expected 1 action got %d", len(loaded.ActionLog))
	}
}
