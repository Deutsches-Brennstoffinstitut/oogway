package oogway

import (
	"context"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/NYTimes/gziphandler"
	"github.com/gorilla/mux"
	"github.com/rjeczalik/notify"
)

const (
	contentDir      = "content"
	contentPageFile = "index.html"
)

var (
	content = newTplCache()
	routes  = newRouter()
)

func loadContent(dir string, funcMap template.FuncMap) error {
	content.clear()
	routes.clear()
	contentDirPath := filepath.Join(dir, contentDir)

	if _, err := os.Stat(contentDirPath); os.IsNotExist(err) || isEmptyDir(contentDirPath) {
		return nil
	}

	return filepath.WalkDir(contentDirPath, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && d.Name() == contentPageFile {
			tpl, err := content.load(path, funcMap)

			if err != nil {
				log.Printf("Error loading template %s: %s", path, err)
				return nil
			}

			route := filepath.Dir(path)[len(contentDirPath):] + "/"
			routes.addRoute(route, tpl)
		}

		return nil
	})
}

func watchContent(ctx context.Context, dir string, funcMap template.FuncMap) error {
	if err := loadContent(dir, funcMap); err != nil {
		return err
	}

	change := make(chan notify.EventInfo, 1)

	go func() {
		for {
			select {
			case <-change:
				if err := loadContent(dir, funcMap); err != nil {
					log.Printf("Error updating content: %s", err)
				}
			case <-ctx.Done():
				notify.Stop(change)
				return
			}
		}
	}()

	if err := notify.Watch(filepath.Join(dir, contentDir, "..."), change, notify.All); err != nil {
		return err
	}

	return nil
}

func servePage(router *mux.Router) {
	router.PathPrefix("/").Handler(gziphandler.GzipHandler(http.HandlerFunc(renderPage)))
}

func renderPage(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	tpl := routes.findTemplate(path)

	if tpl == nil {
		w.WriteHeader(http.StatusNotFound)

		if cfg.Content.NotFound != "" {
			path = cfg.Content.NotFound

			if !strings.HasSuffix(path, "/") {
				path += "/"
			}

			tpl = routes.findTemplate(path)

			if tpl != nil {
				if err := tpl.Execute(w, nil); err != nil {
					log.Printf("Error rendering page %s: %s", r.URL.Path, err)
					w.WriteHeader(http.StatusInternalServerError)
				}
			}
		}

		return
	}

	if err := tpl.Execute(w, nil); err != nil {
		log.Printf("Error rendering page %s: %s", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
