package main

import (
	"encoding/csv"
	"flag"
	"io"
	"os"
	"sort"
	"strconv"
	"sync"
)

// For DICOM-specific needs

var cineMutex *sync.RWMutex = &sync.RWMutex{}
var cineManifestPath string
var cineBulkPath string
var cineLookup = make(map[cineKey][]cineValue)

type cineKey struct {
	Zip    string
	Series string
}

type cineValue struct {
	Dicom          string
	InstanceNumber int
}

const (
	CineColZip           = "zip_file"
	CineColSeries        = "series"
	CineColDicom         = "dicom_file"
	CineColInstancNumber = "instance_number"
)

func init() {
	flag.StringVar(&cineManifestPath, "cinemanifest", "", "If set, should be a manifest containing images that can be stitched together into a CINE")
	flag.StringVar(&cineBulkPath, "cinebulkpath", "", "If set, should be a path (likely Google Storage) where UK Biobank-style Zip files reside")
}

func initializeCINEManifest() error {
	cineManifestFile, err := os.Open(cineManifestPath)
	if err != nil {
		return err
	}
	defer cineManifestFile.Close()

	r := csv.NewReader(cineManifestFile)
	r.Comma = '\t'

	head := make(map[string]int)

	for i := 0; ; i++ {
		line, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if i == 0 {
			for col, label := range line {
				head[label] = col
			}
			continue
		}

		key := cineKey{Zip: line[head[CineColZip]], Series: line[head[CineColSeries]]}

		intInstanceNumber, err := strconv.Atoi(line[head[CineColInstancNumber]])
		if err != nil {
			// Ignore the error
			intInstanceNumber = 0
		}
		value := cineValue{Dicom: line[head[CineColDicom]], InstanceNumber: intInstanceNumber}

		cineLookup[key] = append(cineLookup[key], value)
	}

	return nil
}

func CINEFetchDicomNames(Zip, Series string) ([]string, error) {
	if len(cineLookup) == 0 {
		cineMutex.Lock()

		// Make sure that it wasn't changed while we were waiting for the lock
		if len(cineLookup) == 0 {
			if err := initializeCINEManifest(); err != nil {
				cineMutex.Unlock()
				return nil, err
			}
		}
		cineMutex.Unlock()
	}

	cineMutex.RLock()
	defer cineMutex.RUnlock()

	key := cineKey{Zip: Zip, Series: Series}
	value := cineLookup[key]

	sort.Slice(value, func(i, j int) bool { return value[i].InstanceNumber < value[j].InstanceNumber })

	out := make([]string, 0, len(value))

	for _, dicom := range value {
		out = append(out, dicom.Dicom)
	}

	return out, nil
}
