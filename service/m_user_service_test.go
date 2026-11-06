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

func newMUser(id uint, label string) *model.MUser {
	f := false
	return &model.MUser{
		Id:        id,
		Email:     label,
		CreatedBy: 1,
		CreatedOn: response.JSONTime{Time: time.Now()},
		IsDelete:  &f,
	}
}

func seedMUserRows(t *testing.T, db *gorm.DB) {
	t.Helper()
	labels := []string{"alpha1", "alpha2", "beta3", "beta4", "beta5"}
	for i, l := range labels {
		mustNoError(t, db.Create(newMUser(uint(i+1), l)).Error)
	}
}

func TestMUserService_Constructor(t *testing.T) {
	db := setupTestDB(t)
	svc := NewMUserServiceImpl(db)
	if svc == nil {
		t.Fatal("service is nil")
	}
	impl, ok := svc.(*MUserServiceImpl)
	mustTrue(t, ok, "service is *MUserServiceImpl")
	mustTrue(t, impl.db == db, "db is stored")
}

func TestMUserService_Get(t *testing.T) {
	db := setupTestDB(t)
	svc := NewMUserServiceImpl(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mustNoError(t, db.Create(newMUser(10, "found")).Error)
		got, err := svc.GetMUser(ctx, 10)
		mustNoError(t, err)
		mustEqual(t, uint(10), got.Id, "id")
		mustEqual(t, "found", got.Email, "label")
	})

	t.Run("not found", func(t *testing.T) {
		got, err := svc.GetMUser(ctx, 999)
		mustTrue(t, got == nil, "result is nil")
		mustTrue(t, errors.Is(err, gorm.ErrRecordNotFound), "record not found")
	})
}

func TestMUserService_Create(t *testing.T) {
	db := setupTestDB(t)
	user := seedUser(t, db)
	svc := NewMUserServiceImpl(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		item := newMUser(20, "created")
		item.CreatedBy = 0
		item.CreatedOn = response.JSONTime{}
		mustNoError(t, svc.CreateMUser(ctx, item, user))

		var saved model.MUser
		mustNoError(t, db.First(&saved, 20).Error)
		mustEqual(t, "created", saved.Email, "label")
		mustEqual(t, user.Id, saved.CreatedBy, "created by")
		mustTrue(t, !saved.CreatedOn.IsZero(), "created on set")
	})

	t.Run("data exist", func(t *testing.T) {
		err := svc.CreateMUser(ctx, newMUser(20, "duplicate"), user)
		mustError(t, err, "data exist")
	})
}

func TestMUserService_Create_DBError(t *testing.T) {
	db := setupTestDB(t)
	user := seedUser(t, db)
	svc := NewMUserServiceImpl(db)
	closeDB(t, db)

	err := svc.CreateMUser(context.Background(), newMUser(30, "x"), user)
	mustAnyError(t, err)
	mustTrue(t, err.Error() != "data exist", "not a data exist error")
}

func TestMUserService_Update(t *testing.T) {
	db := setupTestDB(t)
	user := seedUser(t, db)
	svc := NewMUserServiceImpl(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mustNoError(t, db.Create(newMUser(40, "before")).Error)

		upd := newMUser(40, "after")
		mustNoError(t, svc.UpdateMUser(ctx, upd, user))
		mustEqual(t, "after", upd.Email, "response label")
		mustEqual(t, user.Id, upd.ModifiedBy, "response modified by")

		var saved model.MUser
		mustNoError(t, db.First(&saved, 40).Error)
		mustEqual(t, "after", saved.Email, "saved label")
		mustEqual(t, user.Id, saved.ModifiedBy, "saved modified by")
		mustTrue(t, !saved.ModifiedOn.IsZero(), "modified on set")
	})

	t.Run("not found", func(t *testing.T) {
		err := svc.UpdateMUser(ctx, newMUser(999, "none"), user)
		mustError(t, err, "data not found")
	})
}

func TestMUserService_Update_DBError(t *testing.T) {
	db := setupTestDB(t)
	user := seedUser(t, db)
	svc := NewMUserServiceImpl(db)
	closeDB(t, db)

	err := svc.UpdateMUser(context.Background(), newMUser(1, "x"), user)
	mustAnyError(t, err)
	mustTrue(t, err.Error() != "data not found", "not a not found error")
}

func TestMUserService_Delete(t *testing.T) {
	db := setupTestDB(t)
	user := seedUser(t, db)
	svc := NewMUserServiceImpl(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mustNoError(t, db.Create(newMUser(50, "del")).Error)
		mustNoError(t, svc.DeleteMUser(ctx, 50, user))

		var count int64
		mustNoError(t, db.Model(&model.MUser{}).Where("id = ?", 50).Count(&count).Error)
		mustEqual(t, int64(0), count, "row count")
	})

	t.Run("not found", func(t *testing.T) {
		err := svc.DeleteMUser(ctx, 999, user)
		mustError(t, err, "data not found")
	})
}

func TestMUserService_Delete_DBError(t *testing.T) {
	db := setupTestDB(t)
	user := seedUser(t, db)
	svc := NewMUserServiceImpl(db)
	closeDB(t, db)

	err := svc.DeleteMUser(context.Background(), 1, user)
	mustAnyError(t, err)
	mustTrue(t, err.Error() != "data not found", "not a not found error")
}

func TestMUserService_SoftDelete(t *testing.T) {
	db := setupTestDB(t)
	user := seedUser(t, db)
	svc := NewMUserServiceImpl(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mustNoError(t, db.Create(newMUser(60, "soft")).Error)
		mustNoError(t, svc.SoftDeleteMUser(ctx, 60, user))

		var saved model.MUser
		mustNoError(t, db.First(&saved, 60).Error)
		mustTrue(t, saved.IsDelete != nil && *saved.IsDelete, "is delete true")
		// NOTE: production code sets DeletedBy to the deleted user own Id (mUser.Id), not the actor.
		mustEqual(t, uint(60), saved.DeletedBy, "deleted by")
		mustTrue(t, !saved.DeletedOn.IsZero(), "deleted on set")
	})

	t.Run("not found", func(t *testing.T) {
		err := svc.SoftDeleteMUser(ctx, 999, user)
		mustError(t, err, "data not found")
	})
}

func TestMUserService_SoftDelete_DBError(t *testing.T) {
	db := setupTestDB(t)
	user := seedUser(t, db)
	svc := NewMUserServiceImpl(db)
	closeDB(t, db)

	err := svc.SoftDeleteMUser(context.Background(), 1, user)
	mustAnyError(t, err)
	mustTrue(t, err.Error() != "data not found", "not a not found error")
}

func TestMUserService_GetPage(t *testing.T) {
	db := setupTestDB(t)
	seedMUserRows(t, db)
	svc := NewMUserServiceImpl(db)
	ctx := context.Background()

	items := func(t *testing.T, p *response.Page) []model.MUser {
		t.Helper()
		list, ok := p.Content.([]model.MUser)
		mustTrue(t, ok, "content is []model.MUser")
		return list
	}

	t.Run("no sort no filter first page", func(t *testing.T) {
		page, err := svc.GetPageMUser(ctx, nil, nil, "", 0, 2, 2)
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
		page, err := svc.GetPageMUser(ctx, nil, nil, "", 1, 2, 2)
		mustNoError(t, err)
		mustEqual(t, 2, len(items(t, page)), "content length")
		mustTrue(t, !page.First, "not first")
		mustTrue(t, !page.Last, "not last")
	})

	t.Run("last page", func(t *testing.T) {
		page, err := svc.GetPageMUser(ctx, nil, nil, "", 2, 2, 2)
		mustNoError(t, err)
		mustEqual(t, 1, len(items(t, page)), "content length")
		mustTrue(t, page.Last, "last")
		mustTrue(t, !page.First, "not first")
		mustEqual(t, 4, page.Pageable.Offset, "offset")
	})

	t.Run("single page", func(t *testing.T) {
		page, err := svc.GetPageMUser(ctx, nil, nil, "", 0, 5, 5)
		mustNoError(t, err)
		mustEqual(t, int64(1), page.TotalPages, "total pages")
		mustEqual(t, 5, len(items(t, page)), "content length")
		mustTrue(t, page.Last, "last")
	})

	t.Run("sort asc", func(t *testing.T) {
		sorts := []request.Sort{{Id: "id", Desc: false}}
		page, err := svc.GetPageMUser(ctx, sorts, nil, "", 0, 5, 5)
		mustNoError(t, err)
		mustEqual(t, uint(1), items(t, page)[0].Id, "first id")
	})

	t.Run("sort desc", func(t *testing.T) {
		sorts := []request.Sort{{Id: "id", Desc: true}}
		page, err := svc.GetPageMUser(ctx, sorts, nil, "", 0, 5, 5)
		mustNoError(t, err)
		mustEqual(t, uint(5), items(t, page)[0].Id, "first id")
	})

	t.Run("filter equals", func(t *testing.T) {
		filters := []request.Filter{{Id: "email", Value: "alpha1", MatchMode: request.EQUALS}}
		page, err := svc.GetPageMUser(ctx, nil, filters, "", 0, 5, 5)
		mustNoError(t, err)
		mustEqual(t, int64(1), page.TotalElements, "total elements")
		mustEqual(t, uint(1), items(t, page)[0].Id, "id")
	})

	t.Run("filter contains", func(t *testing.T) {
		filters := []request.Filter{{Id: "email", Value: "alpha", MatchMode: request.CONTAINS}}
		page, err := svc.GetPageMUser(ctx, nil, filters, "", 0, 5, 5)
		mustNoError(t, err)
		mustEqual(t, int64(2), page.TotalElements, "total elements")
	})

	t.Run("filter no match is empty", func(t *testing.T) {
		filters := []request.Filter{{Id: "email", Value: "zzz", MatchMode: request.EQUALS}}
		page, err := svc.GetPageMUser(ctx, nil, filters, "", 0, 5, 5)
		mustNoError(t, err)
		mustEqual(t, int64(0), page.TotalElements, "total elements")
		mustTrue(t, page.Empty, "empty")
		mustEqual(t, 0, len(items(t, page)), "content length")
	})

	t.Run("search non-empty", func(t *testing.T) {
		page, err := svc.GetPageMUser(ctx, nil, nil, "beta", 0, 5, 5)
		mustNoError(t, err)
		mustEqual(t, int64(3), page.TotalElements, "total elements")
		mustEqual(t, 3, len(items(t, page)), "content length")
	})

	t.Run("search sanitized", func(t *testing.T) {
		page, err := svc.GetPageMUser(ctx, nil, nil, "al'pha;", 0, 5, 5)
		mustNoError(t, err)
		mustEqual(t, int64(2), page.TotalElements, "total elements")
	})

	t.Run("sort filter search combined", func(t *testing.T) {
		sorts := []request.Sort{{Id: "id", Desc: true}}
		filters := []request.Filter{{Id: "email", Value: "beta", MatchMode: request.CONTAINS}}
		page, err := svc.GetPageMUser(ctx, sorts, filters, "beta", 0, 2, 2)
		mustNoError(t, err)
		mustEqual(t, int64(3), page.TotalElements, "total elements")
		mustEqual(t, int64(2), page.TotalPages, "total pages")
		mustEqual(t, uint(5), items(t, page)[0].Id, "first id")
	})
}

func TestMUserService_GetPage_DBError(t *testing.T) {
	db := setupTestDB(t)
	svc := NewMUserServiceImpl(db)
	closeDB(t, db)

	page, err := svc.GetPageMUser(context.Background(), nil, nil, "", 0, 5, 5)
	mustTrue(t, page == nil, "page is nil")
	mustAnyError(t, err)
}

func TestMUserService_GetMUserByEmail(t *testing.T) {
	db := setupTestDB(t)
	seedUser(t, db)
	svc := NewMUserServiceImpl(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		got, err := svc.GetMUserByEmail(ctx, seedUserEmail)
		mustNoError(t, err)
		mustEqual(t, uint(1), got.Id, "id")
		mustEqual(t, seedUserEmail, got.Email, "email")
	})

	t.Run("not found", func(t *testing.T) {
		got, err := svc.GetMUserByEmail(ctx, "missing@example.com")
		mustTrue(t, got == nil, "result is nil")
		mustTrue(t, errors.Is(err, gorm.ErrRecordNotFound), "record not found")
	})
}
