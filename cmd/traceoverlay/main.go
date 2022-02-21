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

	_ "github.com/carbocation/genomisc/compileinfoprint"
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

	jsonConfig := flag.String("config", "", "Path to JSON file with configuration.")
	port := flag.Int("port", 9019, "Port for HTTP server")
	flag.Parse()

	if *jsonConfig == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	var config overlay.JSONConfig
	var err error

	config, err = overlay.ParseJSONConfigFromPath(*jsonConfig)
	if err != nil {
		log.Fatalln(err)
	}

	// If the command line port is the default port number AND the config file
	// port is set, use the config file. Otherwise, use the command line port.
	if config.Port != 0 && *port == 9019 {
		port = &config.Port
	}

	if !config.Labels.Valid() {
		config.Labels = make(overlay.LabelMap)
	}

	for _, label := range config.Labels {
		if label.Color == "#000000" && label.ID != 0 {
			log.Fatalf("Configuration problem: label %+v has color #000000, which is reserved for the background (label ID 0).\n", label)
		}
	}

	log.Printf("Using configuration:\n%s\n", spew.Sdump(config))

	log.Printf("Labels in effect:\n%v\n", config.Labels.Sorted())

	if config.LabelPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	// If the project isn't a fully path that will store the annotations
	// (inferred by the presence of a path delimiter), assume it's a
	// subdirectory and create it if it does not exist:
	if !strings.Contains(config.LabelPath, "/") {
		newpath := config.LabelPath
		log.Printf("Creating directory ./%s/ if it does not yet exist\n", config.LabelPath)
		newpath = filepath.Join(".", config.LabelPath)
		if err := os.MkdirAll(newpath, os.ModePerm); err != nil {
			log.Fatalln(err)
		}

		config.LabelPath = newpath
	}

	manifestLines, err := ReadManifest(config.ManifestPath, config.LabelPath, config.ImagePath)
	if err != nil {
		log.Fatalln(err)
	}

	global = &Global{
		Site:      "TraceOverlay",
		Company:   "James Pirruccello and The General Hospital Corporation",
		Email:     "jpirruccello@mgh.harvard.edu",
		SnailMail: "55 Fruit Street, Boston MA 02114",
		log:       log.New(os.Stderr, log.Prefix(), log.Ldate|log.Ltime),
		db:        nil,

		Project:      filepath.Base(config.LabelPath),
		ManifestPath: config.ManifestPath,
		manifest:     manifestLines,

		Config: config,
	}

	global.log.Println("Launching", global.Site)

	go func() {
		global.log.Println("Starting HTTP server on port", *port)

		routing, err := router(global)
		if err != nil {
			errors <- err
			global.log.Println(err)
			sig <- syscall.SIGTERM
			return
		}

		if err := http.ListenAndServe(fmt.Sprintf(`:%d`, *port), routing); err != nil {
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
