package shortpaste

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (app *App) handleLink(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		app.handleGetLink(w, r)
	case "POST", "PUT":
		app.handleCreateLink(w, r)
	}
}

func (app *App) handleGetLink(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/l/"), "/")
	if id == "" {
		var links []Link
		app.db.Find(&links)
		json.NewEncoder(w).Encode(map[string][]Link{"links": links})
	} else {
		var link Link
		if err := app.db.First(&link, "id = ?", id).Error; err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Link for `%s` not found!\n", id)
			return
		}

		http.Redirect(w, r, link.Link, http.StatusTemporaryRedirect)
	}
}

func (app *App) handleCreateLink(w http.ResponseWriter, r *http.Request) {
	link := Link{}
	if err := json.NewDecoder(r.Body).Decode(&link); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("%v", err), "message": "failed"})
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/l/")
	if id != "" && link.ID == "" {
		link.ID = id
	}
	if err := link.validate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("%v", err), "message": "failed"})
		return
	}
	if err := app.db.Create(&link).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("%v", err), "message": "failed to create DB entry"})
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "created"})
}
