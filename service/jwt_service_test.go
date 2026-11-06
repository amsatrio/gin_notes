package service

import (
	"context"
	"errors"
	"testing"

	"github.com/amsatrio/gin_notes/model/request"
)

func TestJwtService_Constructor(t *testing.T) {
	db := setupTestDB(t)
	svc := NewJwtServiceImpl(db)
	if svc == nil {
		t.Fatal("service is nil")
	}
	impl, ok := svc.(*JwtServiceImpl)
	mustTrue(t, ok, "service is *JwtServiceImpl")
	mustTrue(t, impl.db == db, "db is stored")
}

func TestJwtService_JwtAuthenticate_Success(t *testing.T) {
	db := setupTestDB(t)
	seedUser(t, db)
	svc := NewJwtServiceImpl(db)

	authorities, err := svc.JwtAuthenticate(context.Background(), request.RequestAuth{
		Username: seedUserEmail,
		Password: seedUserPassword,
	})
	mustNoError(t, err)
	mustEqual(t, 1, len(authorities), "authorities length")
	mustEqual(t, "ADMIN", authorities[0], "authority")
}

func TestJwtService_JwtAuthenticate_InvalidCredentials(t *testing.T) {
	db := setupTestDB(t)
	seedUser(t, db)
	svc := NewJwtServiceImpl(db)
	ctx := context.Background()

	t.Run("wrong password", func(t *testing.T) {
		authorities, err := svc.JwtAuthenticate(ctx, request.RequestAuth{Username: seedUserEmail, Password: "wrong"})
		mustTrue(t, authorities == nil, "authorities is nil")
		mustError(t, err, "username / password invalid")
	})

	t.Run("unknown user", func(t *testing.T) {
		authorities, err := svc.JwtAuthenticate(ctx, request.RequestAuth{Username: "nobody@example.com", Password: seedUserPassword})
		mustTrue(t, authorities == nil, "authorities is nil")
		mustError(t, err, "username / password invalid")
	})
}

func TestJwtService_JwtAuthenticate_DBError(t *testing.T) {
	db := setupTestDB(t)
	seedUser(t, db)
	svc := NewJwtServiceImpl(db)
	closeDB(t, db)

	authorities, err := svc.JwtAuthenticate(context.Background(), request.RequestAuth{Username: seedUserEmail, Password: seedUserPassword})
	mustTrue(t, authorities == nil, "authorities is nil")
	mustAnyError(t, err)
	mustTrue(t, !errors.Is(err, errors.New("username / password invalid")) && err.Error() != "username / password invalid", "not invalid-credentials error")
}
