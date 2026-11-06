package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/amsatrio/gin_notes/initializer"
	"github.com/amsatrio/gin_notes/middleware"
	"github.com/amsatrio/gin_notes/model"
	"github.com/amsatrio/gin_notes/model/response"
)

const (
	seedEmail    = "test@example.com"
	seedPassword = "secret"
	seedRoleCode = "ADMIN"
)

func TestMain(m *testing.M) {
	os.Setenv("AUTH_JWT_TOKEN_SECRET", "test-secret")
	os.Setenv("AUTH_JWT_TOKEN_EXPIRED_MS", "3600000")
	os.Setenv("AUTH_JWT_ENABLE", "false")
	os.Setenv("LOG_MODE", "ERROR")
	os.Setenv("REDIS_ENABLE", "false")
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

// setupDB opens a unique in-memory sqlite db, migrates all models, seeds
// role + biodata + user and assigns it to initializer.DB.
func setupDB(t *testing.T) *gorm.DB {
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

	seedData(t, db)

	prev := initializer.DB
	initializer.DB = db
	t.Cleanup(func() {
		initializer.DB = prev
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	})
	return db
}

func seedData(t *testing.T, db *gorm.DB) {
	t.Helper()
	f := false
	now := response.JSONTime{Time: time.Now()}

	role := model.MRole{Id: 1, Name: "Admin", Code: seedRoleCode, Level: 1, CreatedBy: 1, CreatedOn: now, IsDelete: &f}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("seed role: %v", err)
	}
	bio := model.MBiodata{Id: 1, Fullname: "Seed User", MobilePhone: "0812345", CreatedBy: 1, CreatedOn: now, IsDelete: &f}
	if err := db.Create(&bio).Error; err != nil {
		t.Fatalf("seed biodata: %v", err)
	}
	user := model.MUser{
		Id: 1, BiodataId: 1, RoleId: 1, Email: seedEmail, Password: seedPassword,
		CreatedBy: 1, CreatedOn: now, IsDelete: &f,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
}

func closeDB(t *testing.T, db *gorm.DB) {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("close sql db: %v", err)
	}
}

// crudHandlers groups the 7 standard handlers of a resource.
type crudHandlers struct {
	Page, Create, Update, Index, Delete, SoftDelete, Header gin.HandlerFunc
}

func registerCRUD(r *gin.Engine, base string, h crudHandlers) {
	g := r.Group(base)
	g.GET("", h.Page)
	g.POST("", h.Create)
	g.GET("/header", h.Header)
	g.PUT("/delete/:id", h.SoftDelete)
	g.PUT("/:id", h.Update)
	g.GET("/:id", h.Index)
	g.DELETE("/:id", h.Delete)
}

// newRouter builds an engine with the error middleware, cookie sessions and
// a middleware that sets the "username" context value (if non-empty).
func newRouter(username string) *gin.Engine {
	r := gin.New()
	r.Use(middleware.CustomErrorApiMiddleware())
	store := cookie.NewStore([]byte("test-session-secret"))
	r.Use(sessions.Sessions("session", store))
	r.Use(func(c *gin.Context) {
		if username != "" {
			c.Set("username", username)
		}
		c.Next()
	})
	return r
}

// setupRouter returns a router with every controller registered, with the
// authenticated username set to the seeded user.
func setupRouter() *gin.Engine {
	return setupRouterAs(seedEmail)
}

func setupRouterAs(username string) *gin.Engine {
	r := newRouter(username)

	registerCRUD(r, "/m_notes", crudHandlers{MNotesPage, MNotesCreate, MNotesUpdate, MNotesIndex, MNotesDelete, MNotesSoftDelete, MNotesHeader})
	registerCRUD(r, "/m_biodata", crudHandlers{MBiodataPage, MBiodataCreate, MBiodataUpdate, MBiodataIndex, MBiodataDelete, MBiodataSoftDelete, MBiodataHeader})
	registerCRUD(r, "/m_role", crudHandlers{MRolePage, MRoleCreate, MRoleUpdate, MRoleIndex, MRoleDelete, MRoleSoftDelete, MRoleHeader})
	registerCRUD(r, "/m_user", crudHandlers{MUserPage, MUserCreate, MUserUpdate, MUserIndex, MUserDelete, MUserSoftDelete, MUserHeader})
	registerCRUD(r, "/t_token", crudHandlers{TTokenPage, TTokenCreate, TTokenUpdate, TTokenIndex, TTokenDelete, TTokenSoftDelete, TTokenHeader})
	registerCRUD(r, "/t_reset_password", crudHandlers{TResetPasswordPage, TResetPasswordCreate, TResetPasswordUpdate, TResetPasswordIndex, TResetPasswordDelete, TResetPasswordSoftDelete, TResetPasswordHeader})

	r.POST("/login", JwtLogin)
	r.POST("/logout", JwtLogout)
	r.POST("/refresh", JwtRefreshToken)
	return r
}

// doRequest performs a request against the router and returns the recorder.
func doRequest(r http.Handler, method, target, body string) *httptest.ResponseRecorder {
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, target, nil)
	} else {
		req = httptest.NewRequest(method, target, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func expectStatus(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Fatalf("status: want %d, got %d, body: %s", want, w.Code, w.Body.String())
	}
}

func decodeBody(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	out := map[string]interface{}{}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode body: %v (%s)", err, w.Body.String())
	}
	return out
}

// crudSpec describes a resource for the generic CRUD test suite.
type crudSpec struct {
	base string
	// createBody returns a valid JSON create body for the given id.
	createBody func(id int) string
	// updateBody returns a valid JSON update body (must contain an id).
	updateBody func(id int) string
}

// runCRUDSuite exercises every handler of a resource with its main branches.
func runCRUDSuite(t *testing.T, spec crudSpec) {
	t.Helper()

	t.Run("page", func(t *testing.T) {
		setupDB(t)
		r := setupRouter()

		// create data to be listed
		expectStatus(t, doRequest(r, http.MethodPost, spec.base, spec.createBody(50)), http.StatusOK)

		w := doRequest(r, http.MethodGet, spec.base+"?_page=0&_size=5", "")
		expectStatus(t, w, http.StatusOK)
		if decodeBody(t, w)["data"] == nil {
			t.Fatalf("expected page data")
		}

		// defaults
		expectStatus(t, doRequest(r, http.MethodGet, spec.base, ""), http.StatusOK)

		// valid sort + filter + global search
		q := url.Values{}
		q.Set("_sort", `[{"id":"id","desc":true}]`)
		q.Set("_filter", `[{"id":"id","value":50,"matchMode":"EQUALS","dataType":"NUMBER","mode":"AND"}]`)
		q.Set("_q", "abc 123")
		expectStatus(t, doRequest(r, http.MethodGet, spec.base+"?"+q.Encode(), ""), http.StatusOK)

		// ascending sort + non EQUALS filter
		q = url.Values{}
		q.Set("_sort", `[{"id":"id","desc":false}]`)
		q.Set("_filter", `[{"id":"id","value":"5","matchMode":"CONTAINS","dataType":"TEXT","mode":"AND"}]`)
		expectStatus(t, doRequest(r, http.MethodGet, spec.base+"?"+q.Encode(), ""), http.StatusOK)

		// total elements not a multiple of size
		expectStatus(t, doRequest(r, http.MethodGet, spec.base+"?_size=1", ""), http.StatusOK)
	})

	t.Run("page_bad_params", func(t *testing.T) {
		setupDB(t)
		r := setupRouter()

		expectStatus(t, doRequest(r, http.MethodGet, spec.base+"?_page=abc", ""), http.StatusBadRequest)
		expectStatus(t, doRequest(r, http.MethodGet, spec.base+"?_size=abc", ""), http.StatusBadRequest)
		expectStatus(t, doRequest(r, http.MethodGet, spec.base+"?_q="+url.QueryEscape("a;b"), ""), http.StatusBadRequest)
		expectStatus(t, doRequest(r, http.MethodGet, spec.base+"?_sort="+url.QueryEscape("{bad"), ""), http.StatusBadRequest)
		expectStatus(t, doRequest(r, http.MethodGet, spec.base+"?_filter="+url.QueryEscape("{bad"), ""), http.StatusBadRequest)
	})

	t.Run("page_db_error", func(t *testing.T) {
		setupDB(t)
		r := setupRouter()
		q := url.Values{}
		q.Set("_sort", `[{"id":"no_such_column","desc":false}]`)
		expectStatus(t, doRequest(r, http.MethodGet, spec.base+"?"+q.Encode(), ""), http.StatusBadRequest)
	})

	t.Run("create", func(t *testing.T) {
		setupDB(t)
		r := setupRouter()

		w := doRequest(r, http.MethodPost, spec.base, spec.createBody(60))
		expectStatus(t, w, http.StatusOK)
		if got := decodeBody(t, w)["message"]; got != "success" {
			t.Fatalf("message: %v", got)
		}

		// data exist
		expectStatus(t, doRequest(r, http.MethodPost, spec.base, spec.createBody(60)), http.StatusBadRequest)

		// validation failure (missing id)
		expectStatus(t, doRequest(r, http.MethodPost, spec.base, `{}`), http.StatusBadRequest)

		// malformed json (not a validation error)
		expectStatus(t, doRequest(r, http.MethodPost, spec.base, `{bad`), http.StatusBadRequest)
	})

	t.Run("create_user_not_found", func(t *testing.T) {
		setupDB(t)
		r := setupRouterAs("nobody@example.com")
		expectStatus(t, doRequest(r, http.MethodPost, spec.base, spec.createBody(61)), http.StatusBadRequest)
		expectStatus(t, doRequest(r, http.MethodPut, spec.base+"/1", spec.updateBody(1)), http.StatusBadRequest)
		expectStatus(t, doRequest(r, http.MethodDelete, spec.base+"/1", ""), http.StatusBadRequest)
		expectStatus(t, doRequest(r, http.MethodPut, spec.base+"/delete/1", ""), http.StatusBadRequest)
	})

	t.Run("update", func(t *testing.T) {
		setupDB(t)
		r := setupRouter()

		expectStatus(t, doRequest(r, http.MethodPost, spec.base, spec.createBody(70)), http.StatusOK)

		w := doRequest(r, http.MethodPut, spec.base+"/70", spec.updateBody(70))
		expectStatus(t, w, http.StatusOK)

		// bad id
		expectStatus(t, doRequest(r, http.MethodPut, spec.base+"/abc", spec.updateBody(70)), http.StatusBadRequest)

		// bind failures
		expectStatus(t, doRequest(r, http.MethodPut, spec.base+"/70", `{}`), http.StatusBadRequest)
		expectStatus(t, doRequest(r, http.MethodPut, spec.base+"/70", `{bad`), http.StatusBadRequest)

		// data not found
		expectStatus(t, doRequest(r, http.MethodPut, spec.base+"/9999", spec.updateBody(9999)), http.StatusBadRequest)
	})

	t.Run("index", func(t *testing.T) {
		setupDB(t)
		r := setupRouter()

		expectStatus(t, doRequest(r, http.MethodPost, spec.base, spec.createBody(80)), http.StatusOK)

		w := doRequest(r, http.MethodGet, spec.base+"/80", "")
		expectStatus(t, w, http.StatusOK)
		if decodeBody(t, w)["data"] == nil {
			t.Fatalf("expected data")
		}

		// not found
		expectStatus(t, doRequest(r, http.MethodGet, spec.base+"/9999", ""), http.StatusBadRequest)
		// bad id
		expectStatus(t, doRequest(r, http.MethodGet, spec.base+"/abc", ""), http.StatusBadRequest)
	})

	t.Run("index_db_error", func(t *testing.T) {
		db := setupDB(t)
		r := setupRouter()
		closeDB(t, db)
		expectStatus(t, doRequest(r, http.MethodGet, spec.base+"/1", ""), http.StatusBadRequest)
	})

	t.Run("delete", func(t *testing.T) {
		setupDB(t)
		r := setupRouter()

		expectStatus(t, doRequest(r, http.MethodPost, spec.base, spec.createBody(90)), http.StatusOK)

		expectStatus(t, doRequest(r, http.MethodDelete, spec.base+"/90", ""), http.StatusOK)
		// already deleted -> not found
		expectStatus(t, doRequest(r, http.MethodDelete, spec.base+"/90", ""), http.StatusBadRequest)
		// bad id
		expectStatus(t, doRequest(r, http.MethodDelete, spec.base+"/abc", ""), http.StatusBadRequest)
	})

	t.Run("soft_delete", func(t *testing.T) {
		setupDB(t)
		r := setupRouter()

		expectStatus(t, doRequest(r, http.MethodPost, spec.base, spec.createBody(91)), http.StatusOK)

		expectStatus(t, doRequest(r, http.MethodPut, spec.base+"/delete/91", ""), http.StatusOK)
		// not found
		expectStatus(t, doRequest(r, http.MethodPut, spec.base+"/delete/9999", ""), http.StatusBadRequest)
		// bad id
		expectStatus(t, doRequest(r, http.MethodPut, spec.base+"/delete/abc", ""), http.StatusBadRequest)
	})

	t.Run("header", func(t *testing.T) {
		setupDB(t)
		r := setupRouter()
		w := doRequest(r, http.MethodGet, spec.base+"/header", "")
		expectStatus(t, w, http.StatusOK)
		if decodeBody(t, w)["data"] == nil {
			t.Fatalf("expected header data")
		}
	})
}

// Nested association JSON snippets. Gin validates nested structs, so their
// "required" Id must be present; created_on is NOT NULL for biodata.
const (
	nestedBiodataJSON = `"MBiodata":{"id":1,"createdOn":"2024-02-16 10:33:10","isDelete":false}`
	nestedRoleJSON    = `"MRole":{"id":1,"createdOn":"2024-02-16 10:33:10"}`
	nestedUserJSON    = `"MUser":{"id":1,"createdOn":"2024-02-16 10:33:10",` + nestedBiodataJSON + `,` + nestedRoleJSON + `}`
)
