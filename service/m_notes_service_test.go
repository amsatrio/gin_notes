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

func newMNotes(id uint, label string) *model.MNotes {
	f := false
	return &model.MNotes{
		Id:        id,
		Title:     label,
		CreatedBy: 1,
		CreatedOn: response.JSONTime{Time: time.Now()},
		IsDelete:  &f,
	}
}

func seedMNotesRows(t *testing.T, db *gorm.DB) {
	t.Helper()
	labels := []string{"alpha1", "alpha2", "beta3", "beta4", "beta5"}
	for i, l := range labels {
		mustNoError(t, db.Create(newMNotes(uint(i+1), l)).Error)
	}
}

func TestMNotesService_Constructor(t *testing.T) {
	db := setupTestDB(t)
	svc := NewMNotesServiceImpl(db)
	if svc == nil {
		t.Fatal("service is nil")
	}
	impl, ok := svc.(*MNotesServiceImpl)
	mustTrue(t, ok, "service is *MNotesServiceImpl")
	mustTrue(t, impl.db == db, "db is stored")
}

func TestMNotesService_Get(t *testing.T) {
	db := setupTestDB(t)
	svc := NewMNotesServiceImpl(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mustNoError(t, db.Create(newMNotes(10, "found")).Error)
		got, err := svc.GetMNotes(ctx, 10)
		mustNoError(t, err)
		mustEqual(t, uint(10), got.Id, "id")
		mustEqual(t, "found", got.Title, "label")
	})

	t.Run("not found", func(t *testing.T) {
		got, err := svc.GetMNotes(ctx, 999)
		mustTrue(t, got == nil, "result is nil")
		mustTrue(t, errors.Is(err, gorm.ErrRecordNotFound), "record not found")
	})
}

func TestMNotesService_Create(t *testing.T) {
	db := setupTestDB(t)
	user := seedUser(t, db)
	svc := NewMNotesServiceImpl(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		item := newMNotes(20, "created")
		item.CreatedBy = 0
		item.CreatedOn = response.JSONTime{}
		mustNoError(t, svc.CreateMNotes(ctx, item, user))

		var saved model.MNotes
		mustNoError(t, db.First(&saved, 20).Error)
		mustEqual(t, "created", saved.Title, "label")
		mustEqual(t, user.Id, saved.CreatedBy, "created by")
		mustTrue(t, !saved.CreatedOn.IsZero(), "created on set")
	})

	t.Run("data exist", func(t *testing.T) {
		err := svc.CreateMNotes(ctx, newMNotes(20, "duplicate"), user)
		mustError(t, err, "data exist")
	})
}

func TestMNotesService_Create_DBError(t *testing.T) {
	db := setupTestDB(t)
	user := seedUser(t, db)
	svc := NewMNotesServiceImpl(db)
	closeDB(t, db)

	err := svc.CreateMNotes(context.Background(), newMNotes(30, "x"), user)
	mustAnyError(t, err)
	mustTrue(t, err.Error() != "data exist", "not a data exist error")
}

func TestMNotesService_Update(t *testing.T) {
	db := setupTestDB(t)
	user := seedUser(t, db)
	svc := NewMNotesServiceImpl(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mustNoError(t, db.Create(newMNotes(40, "before")).Error)

		upd := newMNotes(40, "after")
		mustNoError(t, svc.UpdateMNotes(ctx, upd, user))
		mustEqual(t, "after", upd.Title, "response label")
		mustEqual(t, user.Id, upd.ModifiedBy, "response modified by")

		var saved model.MNotes
		mustNoError(t, db.First(&saved, 40).Error)
		mustEqual(t, "after", saved.Title, "saved label")
		mustEqual(t, user.Id, saved.ModifiedBy, "saved modified by")
		mustTrue(t, !saved.ModifiedOn.IsZero(), "modified on set")
	})

	t.Run("not found", func(t *testing.T) {
		err := svc.UpdateMNotes(ctx, newMNotes(999, "none"), user)
		mustError(t, err, "data not found")
	})
}

func TestMNotesService_Update_DBError(t *testing.T) {
	db := setupTestDB(t)
	user := seedUser(t, db)
	svc := NewMNotesServiceImpl(db)
	closeDB(t, db)

	err := svc.UpdateMNotes(context.Background(), newMNotes(1, "x"), user)
	mustAnyError(t, err)
	mustTrue(t, err.Error() != "data not found", "not a not found error")
}

func TestMNotesService_Delete(t *testing.T) {
	db := setupTestDB(t)
	user := seedUser(t, db)
	svc := NewMNotesServiceImpl(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mustNoError(t, db.Create(newMNotes(50, "del")).Error)
		mustNoError(t, svc.DeleteMNotes(ctx, 50, user))

		var count int64
		mustNoError(t, db.Model(&model.MNotes{}).Where("id = ?", 50).Count(&count).Error)
		mustEqual(t, int64(0), count, "row count")
	})

	t.Run("not found", func(t *testing.T) {
		err := svc.DeleteMNotes(ctx, 999, user)
		mustError(t, err, "data not found")
	})
}

func TestMNotesService_Delete_DBError(t *testing.T) {
	db := setupTestDB(t)
	user := seedUser(t, db)
	svc := NewMNotesServiceImpl(db)
	closeDB(t, db)

	err := svc.DeleteMNotes(context.Background(), 1, user)
	mustAnyError(t, err)
	mustTrue(t, err.Error() != "data not found", "not a not found error")
}

func TestMNotesService_SoftDelete(t *testing.T) {
	db := setupTestDB(t)
	user := seedUser(t, db)
	svc := NewMNotesServiceImpl(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mustNoError(t, db.Create(newMNotes(60, "soft")).Error)
		mustNoError(t, svc.SoftDeleteMNotes(ctx, 60, user))

		var saved model.MNotes
		mustNoError(t, db.First(&saved, 60).Error)
		mustTrue(t, saved.IsDelete != nil && *saved.IsDelete, "is delete true")
		mustEqual(t, user.Id, saved.DeletedBy, "deleted by")
		mustTrue(t, !saved.DeletedOn.IsZero(), "deleted on set")
	})

	t.Run("not found", func(t *testing.T) {
		err := svc.SoftDeleteMNotes(ctx, 999, user)
		mustError(t, err, "data not found")
	})
}

func TestMNotesService_SoftDelete_DBError(t *testing.T) {
	db := setupTestDB(t)
	user := seedUser(t, db)
	svc := NewMNotesServiceImpl(db)
	closeDB(t, db)

	err := svc.SoftDeleteMNotes(context.Background(), 1, user)
	mustAnyError(t, err)
	mustTrue(t, err.Error() != "data not found", "not a not found error")
}

func TestMNotesService_GetPage(t *testing.T) {
	db := setupTestDB(t)
	seedMNotesRows(t, db)
	svc := NewMNotesServiceImpl(db)
	ctx := context.Background()

	items := func(t *testing.T, p *response.Page) []model.MNotes {
		t.Helper()
		list, ok := p.Content.([]model.MNotes)
		mustTrue(t, ok, "content is []model.MNotes")
		return list
	}

	t.Run("no sort no filter first page", func(t *testing.T) {
		page, err := svc.GetPageMNotes(ctx, nil, nil, "", 0, 2, 2)
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
		page, err := svc.GetPageMNotes(ctx, nil, nil, "", 1, 2, 2)
		mustNoError(t, err)
		mustEqual(t, 2, len(items(t, page)), "content length")
		mustTrue(t, !page.First, "not first")
		mustTrue(t, !page.Last, "not last")
	})

	t.Run("last page", func(t *testing.T) {
		page, err := svc.GetPageMNotes(ctx, nil, nil, "", 2, 2, 2)
		mustNoError(t, err)
		mustEqual(t, 1, len(items(t, page)), "content length")
		mustTrue(t, page.Last, "last")
		mustTrue(t, !page.First, "not first")
		mustEqual(t, 4, page.Pageable.Offset, "offset")
	})

	t.Run("single page", func(t *testing.T) {
		page, err := svc.GetPageMNotes(ctx, nil, nil, "", 0, 5, 5)
		mustNoError(t, err)
		mustEqual(t, int64(1), page.TotalPages, "total pages")
		mustEqual(t, 5, len(items(t, page)), "content length")
		mustTrue(t, page.Last, "last")
	})

	t.Run("sort asc", func(t *testing.T) {
		sorts := []request.Sort{{Id: "id", Desc: false}}
		page, err := svc.GetPageMNotes(ctx, sorts, nil, "", 0, 5, 5)
		mustNoError(t, err)
		mustEqual(t, uint(1), items(t, page)[0].Id, "first id")
	})

	t.Run("sort desc", func(t *testing.T) {
		sorts := []request.Sort{{Id: "id", Desc: true}}
		page, err := svc.GetPageMNotes(ctx, sorts, nil, "", 0, 5, 5)
		mustNoError(t, err)
		mustEqual(t, uint(5), items(t, page)[0].Id, "first id")
	})

	t.Run("filter equals", func(t *testing.T) {
		filters := []request.Filter{{Id: "title", Value: "alpha1", MatchMode: request.EQUALS}}
		page, err := svc.GetPageMNotes(ctx, nil, filters, "", 0, 5, 5)
		mustNoError(t, err)
		mustEqual(t, int64(1), page.TotalElements, "total elements")
		mustEqual(t, uint(1), items(t, page)[0].Id, "id")
	})

	t.Run("filter contains", func(t *testing.T) {
		filters := []request.Filter{{Id: "title", Value: "alpha", MatchMode: request.CONTAINS}}
		page, err := svc.GetPageMNotes(ctx, nil, filters, "", 0, 5, 5)
		mustNoError(t, err)
		mustEqual(t, int64(2), page.TotalElements, "total elements")
	})

	t.Run("filter no match is empty", func(t *testing.T) {
		filters := []request.Filter{{Id: "title", Value: "zzz", MatchMode: request.EQUALS}}
		page, err := svc.GetPageMNotes(ctx, nil, filters, "", 0, 5, 5)
		mustNoError(t, err)
		mustEqual(t, int64(0), page.TotalElements, "total elements")
		mustTrue(t, page.Empty, "empty")
		mustEqual(t, 0, len(items(t, page)), "content length")
	})

	t.Run("search non-empty", func(t *testing.T) {
		page, err := svc.GetPageMNotes(ctx, nil, nil, "beta", 0, 5, 5)
		mustNoError(t, err)
		mustEqual(t, int64(3), page.TotalElements, "total elements")
		mustEqual(t, 3, len(items(t, page)), "content length")
	})

	t.Run("search sanitized", func(t *testing.T) {
		page, err := svc.GetPageMNotes(ctx, nil, nil, "al'pha;", 0, 5, 5)
		mustNoError(t, err)
		mustEqual(t, int64(2), page.TotalElements, "total elements")
	})

	t.Run("sort filter search combined", func(t *testing.T) {
		sorts := []request.Sort{{Id: "id", Desc: true}}
		filters := []request.Filter{{Id: "title", Value: "beta", MatchMode: request.CONTAINS}}
		page, err := svc.GetPageMNotes(ctx, sorts, filters, "beta", 0, 2, 2)
		mustNoError(t, err)
		mustEqual(t, int64(3), page.TotalElements, "total elements")
		mustEqual(t, int64(2), page.TotalPages, "total pages")
		mustEqual(t, uint(5), items(t, page)[0].Id, "first id")
	})
}

func TestMNotesService_GetPage_DBError(t *testing.T) {
	db := setupTestDB(t)
	svc := NewMNotesServiceImpl(db)
	closeDB(t, db)

	page, err := svc.GetPageMNotes(context.Background(), nil, nil, "", 0, 5, 5)
	mustTrue(t, page == nil, "page is nil")
	mustAnyError(t, err)
}
