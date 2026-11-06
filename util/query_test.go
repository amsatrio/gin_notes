package util

import (
	"reflect"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/amsatrio/gin_notes/model/request"
)

type queryNote struct {
	ID       int    `gorm:"primaryKey" json:"id"`
	FullName string `json:"fullName"`
	Title    string `json:"title"`
	Age      int    `json:"age"`
}

func (queryNote) TableName() string { return "notes" }

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })

	if err := db.AutoMigrate(&queryNote{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	seed := []queryNote{
		{ID: 1, FullName: "Alice Smith", Title: "alpha", Age: 30},
		{ID: 2, FullName: "Bob Jones", Title: "beta", Age: 25},
		{ID: 3, FullName: "Carol White", Title: "gamma", Age: 35},
		{ID: 4, FullName: "Dave Black", Title: "alpha two", Age: 40},
	}
	if err := db.Create(&seed).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}
	return db
}

func ids(notes []queryNote) []int {
	out := make([]int, len(notes))
	for i, n := range notes {
		out[i] = n.ID
	}
	return out
}

func TestApplyPaginate(t *testing.T) {
	db := newTestDB(t)

	tests := []struct {
		name  string
		page  int
		limit int
		want  []int
	}{
		{"first page", 0, 2, []int{1, 2}},
		{"second page", 1, 2, []int{3, 4}},
		{"beyond range", 5, 2, []int{}},
		{"negative page defaults to offset 0", -1, 3, []int{1, 2, 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var notes []queryNote
			if err := db.Order("id").Scopes(ApplyPaginate(tt.page, tt.limit)).Find(&notes).Error; err != nil {
				t.Fatal(err)
			}
			if got := ids(notes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplySorting(t *testing.T) {
	db := newTestDB(t)

	t.Run("no sorts returns same db", func(t *testing.T) {
		if got := ApplySorting(db, nil); got != db {
			t.Error("expected same db instance")
		}
	})

	t.Run("ascending", func(t *testing.T) {
		var notes []queryNote
		q := ApplySorting(db, []request.Sort{{Id: "age", Desc: false}})
		if err := q.Find(&notes).Error; err != nil {
			t.Fatal(err)
		}
		if got, want := ids(notes), []int{2, 1, 3, 4}; !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("descending", func(t *testing.T) {
		var notes []queryNote
		q := ApplySorting(db, []request.Sort{{Id: "age", Desc: true}})
		if err := q.Find(&notes).Error; err != nil {
			t.Fatal(err)
		}
		if got, want := ids(notes), []int{4, 3, 1, 2}; !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

func TestApplyFiltering(t *testing.T) {
	db := newTestDB(t)

	t.Run("no filters", func(t *testing.T) {
		if got := ApplyFiltering(db, nil); got != db {
			t.Error("expected same db instance")
		}
	})

	t.Run("equals string", func(t *testing.T) {
		var notes []queryNote
		q := ApplyFiltering(db, []request.Filter{{Id: "title", Value: "beta", MatchMode: request.EQUALS}})
		if err := q.Find(&notes).Error; err != nil {
			t.Fatal(err)
		}
		if got, want := ids(notes), []int{2}; !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("equals non-string", func(t *testing.T) {
		var notes []queryNote
		q := ApplyFiltering(db, []request.Filter{{Id: "age", Value: 35, MatchMode: request.EQUALS}})
		if err := q.Find(&notes).Error; err != nil {
			t.Fatal(err)
		}
		if got, want := ids(notes), []int{3}; !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("contains uses LIKE", func(t *testing.T) {
		var notes []queryNote
		q := ApplyFiltering(db, []request.Filter{{Id: "title", Value: "alpha", MatchMode: request.CONTAINS}})
		if err := q.Order("id").Find(&notes).Error; err != nil {
			t.Fatal(err)
		}
		if got, want := ids(notes), []int{1, 4}; !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("multiple filters are ANDed", func(t *testing.T) {
		var notes []queryNote
		q := ApplyFiltering(db, []request.Filter{
			{Id: "title", Value: "alpha", MatchMode: request.CONTAINS},
			{Id: "age", Value: 40, MatchMode: request.EQUALS},
		})
		if err := q.Find(&notes).Error; err != nil {
			t.Fatal(err)
		}
		if got, want := ids(notes), []int{4}; !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("non-string value with default mode matches everything", func(t *testing.T) {
		var notes []queryNote
		q := ApplyFiltering(db, []request.Filter{{Id: "title", Value: 123, MatchMode: request.CONTAINS}})
		if err := q.Find(&notes).Error; err != nil {
			t.Fatal(err)
		}
		if len(notes) != 4 {
			t.Errorf("got %d rows, want 4", len(notes))
		}
	})
}

func TestApplyGlobalSearch(t *testing.T) {
	t.Setenv("LOG_MODE", "ERROR")
	db := newTestDB(t)

	modelMap := map[string]string{
		"title": "string",
		"age":   "int",
	}

	t.Run("empty search returns same db", func(t *testing.T) {
		if got := ApplyGlobalSearch(db, "", modelMap); got != db {
			t.Error("expected same db instance")
		}
	})

	t.Run("matches string columns only", func(t *testing.T) {
		var notes []queryNote
		q := ApplyGlobalSearch(db, "alpha", modelMap)
		if err := q.Order("id").Find(&notes).Error; err != nil {
			t.Fatal(err)
		}
		if got, want := ids(notes), []int{1, 4}; !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("multiple string columns are ORed with snake_case names", func(t *testing.T) {
		var notes []queryNote
		m := map[string]string{"fullName": "string", "title": "string"}
		q := ApplyGlobalSearch(db, "bob", m)
		if err := q.Find(&notes).Error; err != nil {
			t.Fatal(err)
		}
		if got, want := ids(notes), []int{2}; !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("special characters are stripped", func(t *testing.T) {
		var notes []queryNote
		q := ApplyGlobalSearch(db, "beta'; DROP TABLE notes;--", modelMap)
		// after cleansing: "beta DROP TABLE notes" -> no title contains that
		if err := q.Find(&notes).Error; err != nil {
			t.Fatalf("query should not fail: %v", err)
		}
		if len(notes) != 0 {
			t.Errorf("expected no rows, got %v", ids(notes))
		}
		var count int64
		if err := db.Model(&queryNote{}).Count(&count).Error; err != nil || count != 4 {
			t.Errorf("table should be intact, count=%d err=%v", count, err)
		}
	})

	t.Run("search of only special characters yields empty cleansed query", func(t *testing.T) {
		// cleansed search is empty so no where clause is added
		q := ApplyGlobalSearch(db, "!!!", modelMap)
		var notes []queryNote
		if err := q.Find(&notes).Error; err != nil {
			t.Fatal(err)
		}
		if len(notes) == 0 {
			// searchQuery is non-empty ("title like '%%'") so all rows match
			t.Error("expected all rows to match empty pattern")
		}
	})

	t.Run("non-string columns only builds empty where", func(t *testing.T) {
		q := ApplyGlobalSearch(db, "abc", map[string]string{"age": "int"})
		sql := q.ToSQL(func(tx *gorm.DB) *gorm.DB {
			var notes []queryNote
			return q.Find(&notes)
		})
		if strings.Contains(strings.ToLower(sql), "like") {
			t.Errorf("unexpected LIKE clause: %s", sql)
		}
	})
}

type fieldTypesSample struct {
	Name   string  `json:"name"`
	Age    int     `json:"age"`
	Score  float64 `json:"score"`
	Active bool    `json:"active"`
}

func TestGetJSONFieldTypes(t *testing.T) {
	got := GetJSONFieldTypes(fieldTypesSample{})
	want := map[string]string{
		"name":   "string",
		"age":    "int",
		"score":  "float64",
		"active": "bool",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
