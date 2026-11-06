package service

import (
	"os"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/amsatrio/gin_notes/model"
	"github.com/amsatrio/gin_notes/model/response"
)

func TestMain(m *testing.M) {
	os.Setenv("LOG_MODE", "ERROR")
	os.Exit(m.Run())
}

// setupTestDB creates a unique in-memory SQLite database for the test and migrates all models.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(
		sqlite.Open("file:"+name+"?mode=memory&cache=shared"),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}

	err = db.AutoMigrate(
		&model.MRole{},
		&model.MBiodata{},
		&model.MUser{},
		&model.MNotes{},
		&model.TToken{},
		&model.TResetPassword{},
	)
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	})

	return db
}

// closeDB closes the underlying connection so that subsequent queries fail.
func closeDB(t *testing.T, db *gorm.DB) {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql db: %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("failed to close sql db: %v", err)
	}
}

const (
	seedUserEmail    = "user@example.com"
	seedUserPassword = "secret"
)

// seedUser inserts a role (id 1), biodata (id 1) and user (id 1) and returns the user.
func seedUser(t *testing.T, db *gorm.DB) *model.MUser {
	t.Helper()

	f := false
	now := response.JSONTime{Time: time.Now()}

	role := model.MRole{Id: 1, Name: "Admin", Code: "ADMIN", Level: 1, CreatedBy: 1, CreatedOn: now, IsDelete: &f}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("failed to seed role: %v", err)
	}

	biodata := model.MBiodata{Id: 1, Fullname: "Seed User", MobilePhone: "0812345", CreatedBy: 1, CreatedOn: now, IsDelete: &f}
	if err := db.Create(&biodata).Error; err != nil {
		t.Fatalf("failed to seed biodata: %v", err)
	}

	user := model.MUser{
		Id:        1,
		BiodataId: biodata.Id,
		RoleId:    role.Id,
		Email:     seedUserEmail,
		Password:  seedUserPassword,
		CreatedBy: 1,
		CreatedOn: now,
		IsDelete:  &f,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	return &user
}

func mustNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func mustError(t *testing.T, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error %q, got nil", msg)
	}
	if msg != "" && err.Error() != msg {
		t.Fatalf("expected error %q, got %q", msg, err.Error())
	}
}

func mustAnyError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}

func mustEqual[T comparable](t *testing.T, want, got T, what string) {
	t.Helper()
	if want != got {
		t.Fatalf("%s: want %v, got %v", what, want, got)
	}
}

func mustTrue(t *testing.T, cond bool, what string) {
	t.Helper()
	if !cond {
		t.Fatalf("expected true: %s", what)
	}
}
