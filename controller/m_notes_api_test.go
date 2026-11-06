package controller

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

var notesSpec = crudSpec{
	base: "/m_notes",
	createBody: func(id int) string {
		return fmt.Sprintf(`{"id":%[1]d,"title":"title %[1]d","content":"content"}`, id)
	},
	updateBody: func(id int) string {
		return fmt.Sprintf(`{"id":%[1]d,"title":"updated","content":"changed"}`, id)
	},
}

func TestMNotes_CRUDSuite(t *testing.T) {
	runCRUDSuite(t, notesSpec)
}

func TestMNotes_PageSpecialCharQuery(t *testing.T) {
	setupDB(t)
	r := setupRouter()
	w := doRequest(r, http.MethodGet, "/m_notes?_q="+url.QueryEscape("a%b'"), "")
	expectStatus(t, w, http.StatusBadRequest)
}

func TestMNotes_CreateAndFetchBack(t *testing.T) {
	setupDB(t)
	r := setupRouter()

	expectStatus(t, doRequest(r, http.MethodPost, "/m_notes", `{"id":5,"title":"hello","content":"world"}`), http.StatusOK)
	w := doRequest(r, http.MethodGet, "/m_notes/5", "")
	expectStatus(t, w, http.StatusOK)
	data, ok := decodeBody(t, w)["data"].(map[string]interface{})
	if !ok || data["title"] != "hello" {
		t.Fatalf("unexpected data: %v", w.Body.String())
	}

	expectStatus(t, doRequest(r, http.MethodPut, "/m_notes/5", `{"id":5,"title":"new","content":"x"}`), http.StatusOK)
	w = doRequest(r, http.MethodGet, "/m_notes/5", "")
	data = decodeBody(t, w)["data"].(map[string]interface{})
	if data["title"] != "new" {
		t.Fatalf("title not updated: %v", w.Body.String())
	}
}
