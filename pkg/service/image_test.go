package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/bsonger/devflow-release-service/pkg/model"
	"github.com/google/uuid"
)

type recordingExecResult struct{}

func (recordingExecResult) LastInsertId() (int64, error) { return 0, nil }
func (recordingExecResult) RowsAffected() (int64, error) { return 1, nil }

type recordingExecDB struct {
	query string
	args  []any
	err   error
}

func (db *recordingExecDB) ExecContext(_ context.Context, query string, args ...any) (sql.Result, error) {
	db.query = query
	db.args = args
	if db.err != nil {
		return nil, db.err
	}
	return recordingExecResult{}, nil
}

func TestArchiveActiveImageByNameSoftDeletesExistingRecord(t *testing.T) {
	now := time.Date(2026, 4, 9, 10, 0, 0, 0, time.UTC)
	appID := uuid.New()
	img := &model.Image{
		ApplicationID: appID,
		Name:          "demo-main",
	}
	db := &recordingExecDB{}

	if err := archiveActiveImageByName(context.Background(), db, img, now); err != nil {
		t.Fatalf("archiveActiveImageByName returned error: %v", err)
	}

	if db.query == "" {
		t.Fatal("expected ExecContext to be called")
	}
	if len(db.args) != 3 {
		t.Fatalf("got %d args want 3", len(db.args))
	}
	if got := db.args[0]; got != appID {
		t.Fatalf("arg 0 = %v want %v", got, appID)
	}
	if got := db.args[1]; got != "demo-main" {
		t.Fatalf("arg 1 = %v want demo-main", got)
	}
	if got := db.args[2]; got != now {
		t.Fatalf("arg 2 = %v want %v", got, now)
	}
}
