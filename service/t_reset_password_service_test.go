package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/amsatrio/gin_notes/model"
	"github.com/amsatrio/gin_notes/model/request"
	"github.com/amsatrio/gin_notes/model/response"
)

func newTResetPassword(id uint, label string) *model.TResetPassword {
	f := false
	return &model.TResetPassword{
		Id:        id,
		ResetFor:  label,
		CreatedBy: 1,
		CreatedOn: response.JSONTime{Time: time.Now()},
		IsDelete:  &f,
	}
}

func seedTResetPasswordRows(t *testing.T, db *gorm.DB) {
	t.Helper()
	labels := []string{"alpha1", "alpha2", "beta3", "beta4", "beta5"}
	for i, l := range labels {
		mustNoError(t, db.Create(newTResetPassword(uint(i+1), l)).Error)
	}
}

func TestTResetPasswordService_Constructor(t *testing.T) {
	db := setupTestDB(t)
	svc := NewTResetPasswordServiceImpl(db)
	if svc == nil {
		t.Fatal("service is nil")
	}
	impl, ok := svc.(*TResetPasswordServiceImpl)
	mustTrue(t, ok, "service is *TResetPasswordServiceImpl")
	mustTrue(t, impl.db == db, "db is stored")
}

func TestTResetPasswordService_Get(t *testing.T) {
	db := setupTestDB(t)
	svc := NewTResetPasswordServiceImpl(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mustNoError(t, db.Create(newTResetPassword(10, "found")).Error)
		got, err := svc.GetTResetPassword(ctx, 10)
		mustNoError(t, err)
		mustEqual(t, uint(10), got.Id, "id")
		mustEqual(t, "found", got.ResetFor, "label")
	})

	t.Run("not found", func(t *testing.T) {
		got, err := svc.GetTResetPassword(ctx, 999)
		mustTrue(t, got == nil, "result is nil")
		mustTrue(t, errors.Is(err, gorm.ErrRecordNotFound), "record not found")
	})
}

func TestTResetPasswordService_Create(t *testing.T) {
	db := setupTestDB(t)
	user := seedUser(t, db)
	svc := NewTResetPasswordServiceImpl(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		item := newTResetPassword(20, "created")
		item.CreatedBy = 0
		item.CreatedOn = response.JSONTime{}
		mustNoError(t, svc.CreateTResetPassword(ctx, item, user))

		var saved model.TResetPassword
		mustNoError(t, db.First(&saved, 20).Error)
		mustEqual(t, "created", saved.ResetFor, "label")
		mustEqual(t, user.Id, saved.CreatedBy, "created by")
		mustTrue(t, !saved.CreatedOn.IsZero(), "created on set")
	})

	t.Run("data exist", func(t *testing.T) {
		err := svc.CreateTResetPassword(ctx, newTResetPassword(20, "duplicate"), user)
		mustError(t, err, "data exist")
	})
}

func TestTResetPasswordService_Create_DBError(t *testing.T) {
	db := setupTestDB(t)
	user := seedUser(t, db)
	svc := NewTResetPasswordServiceImpl(db)
	closeDB(t, db)

	err := svc.CreateTResetPassword(context.Background(), newTResetPassword(30, "x"), user)
	mustAnyError(t, err)
	mustTrue(t, err.Error() != "data exist", "not a data exist error")
}

func TestTResetPasswordService_Update(t *testing.T) {
	db := setupTestDB(t)
	user := seedUser(t, db)
	svc := NewTResetPasswordServiceImpl(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mustNoError(t, db.Create(newTResetPassword(40, "before")).Error)

		upd := newTResetPassword(40, "after")
		mustNoError(t, svc.UpdateTResetPassword(ctx, upd, user))
		mustEqual(t, "after", upd.ResetFor, "response label")
		mustEqual(t, user.Id, upd.ModifiedBy, "response modified by")

		var saved model.TResetPassword
		mustNoError(t, db.First(&saved, 40).Error)
		mustEqual(t, "after", saved.ResetFor, "saved label")
		mustEqual(t, user.Id, saved.ModifiedBy, "saved modified by")
		mustTrue(t, !saved.ModifiedOn.IsZero(), "modified on set")
	})

	t.Run("not found", func(t *testing.T) {
		err := svc.UpdateTResetPassword(ctx, newTResetPassword(999, "none"), user)
		mustError(t, err, "data not found")
	})
}

func TestTResetPasswordService_Update_DBError(t *testing.T) {
	db := setupTestDB(t)
	user := seedUser(t, db)
	svc := NewTResetPasswordServiceImpl(db)
	closeDB(t, db)

	err := svc.UpdateTResetPassword(context.Background(), newTResetPassword(1, "x"), user)
	mustAnyError(t, err)
	mustTrue(t, err.Error() != "data not found", "not a not found error")
}

func TestTResetPasswordService_Delete(t *testing.T) {
	db := setupTestDB(t)
	user := seedUser(t, db)
	svc := NewTResetPasswordServiceImpl(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mustNoError(t, db.Create(newTResetPassword(50, "del")).Error)
		mustNoError(t, svc.DeleteTResetPassword(ctx, 50, user))

		var count int64
		mustNoError(t, db.Model(&model.TResetPassword{}).Where("id = ?", 50).Count(&count).Error)
		mustEqual(t, int64(0), count, "row count")
	})

	t.Run("not found", func(t *testing.T) {
		err := svc.DeleteTResetPassword(ctx, 999, user)
		mustError(t, err, "data not found")
	})
}

func TestTResetPasswordService_Delete_DBError(t *testing.T) {
	db := setupTestDB(t)
	user := seedUser(t, db)
	svc := NewTResetPasswordServiceImpl(db)
	closeDB(t, db)

	err := svc.DeleteTResetPassword(context.Background(), 1, user)
	mustAnyError(t, err)
	mustTrue(t, err.Error() != "data not found", "not a not found error")
}

func TestTResetPasswordService_SoftDelete(t *testing.T) {
	db := setupTestDB(t)
	user := seedUser(t, db)
	svc := NewTResetPasswordServiceImpl(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mustNoError(t, db.Create(newTResetPassword(60, "soft")).Error)
		mustNoError(t, svc.SoftDeleteTResetPassword(ctx, 60, user))

		var saved model.TResetPassword
		mustNoError(t, db.First(&saved, 60).Error)
		mustTrue(t, saved.IsDelete != nil && *saved.IsDelete, "is delete true")
		mustEqual(t, user.Id, saved.DeletedBy, "deleted by")
		mustTrue(t, !saved.DeletedOn.IsZero(), "deleted on set")
	})

	t.Run("not found", func(t *testing.T) {
		err := svc.SoftDeleteTResetPassword(ctx, 999, user)
		mustError(t, err, "data not found")
	})
}

func TestTResetPasswordService_SoftDelete_DBError(t *testing.T) {
	db := setupTestDB(t)
	user := seedUser(t, db)
	svc := NewTResetPasswordServiceImpl(db)
	closeDB(t, db)

	err := svc.SoftDeleteTResetPassword(context.Background(), 1, user)
	mustAnyError(t, err)
	mustTrue(t, err.Error() != "data not found", "not a not found error")
}

func TestTResetPasswordService_GetPage(t *testing.T) {
	db := setupTestDB(t)
	seedTResetPasswordRows(t, db)
	svc := NewTResetPasswordServiceImpl(db)
	ctx := context.Background()

	items := func(t *testing.T, p *response.Page) []model.TResetPassword {
		t.Helper()
		list, ok := p.Content.([]model.TResetPassword)
		mustTrue(t, ok, "content is []model.TResetPassword")
		return list
	}

	t.Run("no sort no filter first page", func(t *testing.T) {
		page, err := svc.GetPageTResetPassword(ctx, nil, nil, "", 0, 2, 2)
		mustNoError(t, err)
		mustEqual(t, int64(5), page.TotalElements, "total elements")
		mustEqual(t, int64(3), page.TotalPages, "total pages")
		mustEqual(t, 2, len(items(t, page)), "content length")
		mustTrue(t, page.First, "first")
		mustTrue(t, !page.Last, "not last")
		mustTrue(t, !page.Empty, "not empty")
		mustEqual(t, 0, page.Number, "number")
		mustEqual(t, 2, page.Size, "size")
	})

	t.Run("middle page", func(t *testing.T) {
		page, err := svc.GetPageTResetPassword(ctx, nil, nil, "", 1, 2, 2)
		mustNoError(t, err)
		mustEqual(t, 2, len(items(t, page)), "content length")
		mustTrue(t, !page.First, "not first")
		mustTrue(t, !page.Last, "not last")
	})

	t.Run("last page", func(t *testing.T) {
		page, err := svc.GetPageTResetPassword(ctx, nil, nil, "", 2, 2, 2)
		mustNoError(t, err)
		mustEqual(t, 1, len(items(t, page)), "content length")
		mustTrue(t, page.Last, "last")
		mustTrue(t, !page.First, "not first")
		mustEqual(t, 4, page.Pageable.Offset, "offset")
	})

	t.Run("single page", func(t *testing.T) {
		page, err := svc.GetPageTResetPassword(ctx, nil, nil, "", 0, 5, 5)
		mustNoError(t, err)
		mustEqual(t, int64(1), page.TotalPages, "total pages")
		mustEqual(t, 5, len(items(t, page)), "content length")
		mustTrue(t, page.Last, "last")
	})

	t.Run("sort asc", func(t *testing.T) {
		sorts := []request.Sort{{Id: "id", Desc: false}}
		page, err := svc.GetPageTResetPassword(ctx, sorts, nil, "", 0, 5, 5)
		mustNoError(t, err)
		mustEqual(t, uint(1), items(t, page)[0].Id, "first id")
	})

	t.Run("sort desc", func(t *testing.T) {
		sorts := []request.Sort{{Id: "id", Desc: true}}
		page, err := svc.GetPageTResetPassword(ctx, sorts, nil, "", 0, 5, 5)
		mustNoError(t, err)
		mustEqual(t, uint(5), items(t, page)[0].Id, "first id")
	})

	t.Run("filter equals", func(t *testing.T) {
		filters := []request.Filter{{Id: "reset_for", Value: "alpha1", MatchMode: request.EQUALS}}
		page, err := svc.GetPageTResetPassword(ctx, nil, filters, "", 0, 5, 5)
		mustNoError(t, err)
		mustEqual(t, int64(1), page.TotalElements, "total elements")
		mustEqual(t, uint(1), items(t, page)[0].Id, "id")
	})

	t.Run("filter contains", func(t *testing.T) {
		filters := []request.Filter{{Id: "reset_for", Value: "alpha", MatchMode: request.CONTAINS}}
		page, err := svc.GetPageTResetPassword(ctx, nil, filters, "", 0, 5, 5)
		mustNoError(t, err)
		mustEqual(t, int64(2), page.TotalElements, "total elements")
	})

	t.Run("filter no match is empty", func(t *testing.T) {
		filters := []request.Filter{{Id: "reset_for", Value: "zzz", MatchMode: request.EQUALS}}
		page, err := svc.GetPageTResetPassword(ctx, nil, filters, "", 0, 5, 5)
		mustNoError(t, err)
		mustEqual(t, int64(0), page.TotalElements, "total elements")
		mustTrue(t, page.Empty, "empty")
		mustEqual(t, 0, len(items(t, page)), "content length")
	})

	t.Run("search non-empty", func(t *testing.T) {
		page, err := svc.GetPageTResetPassword(ctx, nil, nil, "beta", 0, 5, 5)
		mustNoError(t, err)
		mustEqual(t, int64(3), page.TotalElements, "total elements")
		mustEqual(t, 3, len(items(t, page)), "content length")
	})

	t.Run("search sanitized", func(t *testing.T) {
		page, err := svc.GetPageTResetPassword(ctx, nil, nil, "al'pha;", 0, 5, 5)
		mustNoError(t, err)
		mustEqual(t, int64(2), page.TotalElements, "total elements")
	})

	t.Run("sort filter search combined", func(t *testing.T) {
		sorts := []request.Sort{{Id: "id", Desc: true}}
		filters := []request.Filter{{Id: "reset_for", Value: "beta", MatchMode: request.CONTAINS}}
		page, err := svc.GetPageTResetPassword(ctx, sorts, filters, "beta", 0, 2, 2)
		mustNoError(t, err)
		mustEqual(t, int64(3), page.TotalElements, "total elements")
		mustEqual(t, int64(2), page.TotalPages, "total pages")
		mustEqual(t, uint(5), items(t, page)[0].Id, "first id")
	})
}

func TestTResetPasswordService_GetPage_DBError(t *testing.T) {
	db := setupTestDB(t)
	svc := NewTResetPasswordServiceImpl(db)
	closeDB(t, db)

	page, err := svc.GetPageTResetPassword(context.Background(), nil, nil, "", 0, 5, 5)
	mustTrue(t, page == nil, "page is nil")
	mustAnyError(t, err)
}
