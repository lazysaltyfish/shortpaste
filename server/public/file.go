package public

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"shortpaste/core/config"
	"shortpaste/core/database"
	"shortpaste/core/tools"
	"strings"
	"text/template"

	"github.com/go-chi/chi/v5"
)

const PdfExt = ".pdf"

func FileGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	db := database.Get()

	var file database.File
	if err := db.First(&file, "id = ?", id).Error; err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	filePath := path.Join(config.GetDataDirPath(), "files", file.ID, file.Name)
	// Currently not be used by this page.
	_, isDownload := r.URL.Query()["download"]

	_, isView := r.URL.Query()["view"]
	_, isInline := r.URL.Query()["inline"]
	if isDownload || isView {
		if isDownload {
			file.DownloadCount += 1
			db.Save(&file)
		}

		w.Header().Set("Content-Disposition", "attachment; filename="+file.Name)
		http.ServeFile(w, r, filePath)
		return
	} else if isInline {
		w.Header().Set("Content-Disposition", "inline")
		http.ServeFile(w, r, filePath)
		return
	}

	file.HitCount += 1
	db.Save(&file)

	t, err := template.ParseFS(templateFS, "templates/file.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fi, err := os.Stat(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := struct {
		Name  string
		Src   string
		InlineSrc string
		Image bool
		Pdf   bool
		Size  string
	}{
		Name:  file.Name,
		Src:   "/f/" + id + "?view",
		InlineSrc: "/f/" + id + "?inline",
		Image: strings.HasPrefix(file.MIME, "image/"),
		Pdf:   filepath.Ext(file.Name) == PdfExt,
		Size:  tools.IECFormat(fi.Size()),
	}
	t.Execute(w, data)
}
