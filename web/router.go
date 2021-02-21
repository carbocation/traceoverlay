package main

import (
	"io/fs"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/interpose/middleware"
	"github.com/justinas/alice"
)

func router(config *Global) (http.Handler, error) {
	router := mux.NewRouter()
	POST := router.Methods("POST").Subrouter()
	GET := router.Methods("GET", "HEAD").Subrouter()

	h := handler{Global: config, router: router}

	GET.HandleFunc("/", h.Index).Name("index")
	GET.HandleFunc("/goroutines", h.Goroutines)
	GET.HandleFunc("/{template:(?:about|privacy|TOS|DMCA)}", h.TemplateOnly)
	GET.HandleFunc("/traceoverlay/{manifest_index}", h.TraceOverlay).Name("traceoverlay")
	GET.HandleFunc("/traceoverlay/cine/{zip}/{series}", h.TraceOverlayCINE).Name("traceoverlay_cine")
	GET.HandleFunc("/listproject", h.ListProject).Name("listproject")

	//
	// POST
	//
	POST.Handle("/", http.NotFoundHandler())
	POST.HandleFunc("/traceoverlay/{manifest_index}", h.TraceOverlayPost)

	// Static assets
	assetFilesystem, err := fs.Sub(embeddedTemplates, "templates/static")
	if err != nil {
		return nil, err
	}
	staticHandler := http.StripPrefix(h.Assets(), http.FileServer(http.FS(assetFilesystem)))
	GET.PathPrefix(h.Assets()).Handler(middleware.MaxAgeHandler(60*60*24*364, staticHandler))

	standard := alice.New(
		// Log all requests to STDOUT
		middleware.GorillaLog(),
	)

	return standard.Then(router), nil
}
