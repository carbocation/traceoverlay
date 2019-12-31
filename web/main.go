package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/carbocation/genomisc/overlay"
	"github.com/davecgh/go-spew/spew"
)

var global *Global

func init() {
	// Prevent seed re-use
	rand.Seed(int64(time.Now().Nanosecond()))

	flag.Usage = func() {
		flag.PrintDefaults()

		log.Println("Optional config file layout:")
		// json.NewEncoder(os.Stderr).Encode()
		bts, err := json.MarshalIndent(overlay.JSONConfig{Labels: overlay.LabelMap{"Background": overlay.Label{Color: "", ID: 0}}}, "", "  ")
		if err == nil {
			log.Println(string(bts))
		}
	}
}

func main() {
	errors := make(chan error, 1)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig,
		os.Interrupt,
		os.Kill,
		syscall.SIGTERM,
		syscall.SIGUSR1,
		//syscall.SIGINFO,
	)

	log.Println("Launched with arguments:")
	log.Println(os.Args)

	jsonConfig := flag.String("config", "", "Path to JSON file with configuration. If this includes 'manifest' and 'project' keys, then those do not need to be set on the command line. If set on the command line, they will override the config file.")
	manifest := flag.String("manifest", "", "Tab-delimited manifest file which contains a zip_file and a dicom_file column (at least).")
	project := flag.String("project", "", "Project name. Defines a folder into which all overlays will be written.")
	port := flag.Int("port", 9019, "Port for HTTP server")
	//dbName := flag.String("db_name", "pubrank", "Name of the database schema to connect to")
	flag.Parse()

	var config overlay.JSONConfig
	var err error
	if *jsonConfig != "" {
		config, err = overlay.ParseJSONConfigFromPath(*jsonConfig)
		if err == nil {
			if *manifest == "" {
				*manifest = config.ManifestPath
			}
			if *project == "" {
				*project = config.Project
			}
			if *port == 0 {
				*port = config.Port
			}
		} else {
			log.Fatalln(err)
		}

		if !config.Labels.Valid() {
			config.Labels = make(overlay.LabelMap)
		}

		log.Printf("Using configuration:\n%s\n", spew.Sdump(config))

		log.Printf("Labels in effect:\n%v\n", config.Labels.Sorted())
	}

	if *manifest == "" || *project == "" {
		flag.Usage()
		os.Exit(1)
	}

	// If the project isn't a fully path that will store the annotations
	// (inferred by the presence of a path delimiter), assume it's a
	// subdirectory and create it if it does not exist:
	newpath := *project
	if !strings.Contains(*project, "/") {
		log.Printf("Creating directory ./%s/ if it does not yet exist\n", *project)
		newpath = filepath.Join(".", *project)
		if err := os.MkdirAll(newpath, os.ModePerm); err != nil {
			log.Fatalln(err)
		}
	}

	manifestLines, err := ReadManifest(*manifest, *project)
	if err != nil {
		log.Fatalln(err)
	}

	global = &Global{
		Site:      "TraceOverlay",
		Company:   "Carbocation Corporation",
		Email:     "james@carbocation.com",
		SnailMail: "4 Longfellow Pl Apt 2901, Boston MA 02114",
		log:       log.New(os.Stderr, log.Prefix(), log.Ldate|log.Ltime),
		db:        nil,

		Project:      newpath,
		ManifestPath: *manifest,
		manifest:     manifestLines,

		Config: config,
	}

	global.log.Println("Launching", global.Site)

	go func() {
		global.log.Println("Starting HTTP server on port", *port)
		if err := http.ListenAndServe(fmt.Sprintf(`:%d`, *port), router(global)); err != nil {
			errors <- err
			global.log.Println(err)
			sig <- syscall.SIGTERM
			return
		}
	}()

Outer:
	for {
		select {
		case sigl := <-sig:

			//if sigl == syscall.SIGINFO || sigl == syscall.SIGUSR1 {
			if sigl == syscall.SIGUSR1 {
				SigStatus()
				continue
			}

			// By default, exit
			global.log.Printf("\nExit: %s\n", sigl.String())

			break Outer

		case err := <-errors:
			if err == nil {
				global.log.Println("Finished")
				break Outer
			}

			// Return a status code indicating failure
			global.log.Println("Exiting due to error", err)
			os.Exit(1)
		}
	}
}

func SigStatus() {
	global.log.Println("There are", runtime.NumGoroutine(), "goroutines running")
}
