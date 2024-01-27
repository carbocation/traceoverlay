package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"sync"

	"github.com/carbocation/genomisc/cdf16match"
	"github.com/carbocation/genomisc/cdf16matchv2"
)

// For DICOM-specific needs

var cineMutex *sync.RWMutex = &sync.RWMutex{}
var cineManifestPath string
var cineBulkPath string
var cineLookup = make(map[cineZip]map[cineSeriesID][]cineValue)

type cineZip string
type cineSeriesID string

type cineValue struct {
	Dicom          string
	Series         cineSeriesID
	InstanceNumber int
}

const (
	CineColSeries = "series"
	CineColZip    = "zip_file"
	CineColDicom  = "dicom_file"
)

var (
	CineColInstancNumber = "instance_number"
	CinePixelScaling     = "official"
	CinePixelCDFMatch    = ""
	CinePixelCDFMatchV2  = ""
)

func init() {
	flag.StringVar(&cineManifestPath, "cinemanifest", "", "If set, should be a manifest containing images that can be stitched together into a CINE")
	flag.StringVar(&cineBulkPath, "cinebulkpath", "", "If set, should be a path (likely Google Storage) where UK Biobank-style Zip files reside")
	flag.StringVar(&CineColInstancNumber, "cine_sequence_column_name", "instance_number", "If cinemanifest is provided, this value represents the name of the column that indicates the order of the images with an increasing number.")
	flag.StringVar(&CinePixelScaling, "cine_pixel_scale", "official", "When displaying cines, apply this scaling. Options: raw (no scaling), pythonic (range from max to min observed pixels), official (use the recommended windowing from the DICOM). Default: official.")
	flag.StringVar(&CinePixelCDFMatch, "cine_pixel_cdf_match", "", fmt.Sprintf("After applying --cine_pixel_scale, apply a CDF Match algorithm? If not empty, options are: %s", cdf16match.PrecomputedList()))
	flag.StringVar(&CinePixelCDFMatchV2, "cine_pixel_cdf_match_v2", "", fmt.Sprintf("After applying --cine_pixel_scale, apply a CDF Match algorithm? If not empty, options are: %s", cdf16matchv2.PrecomputedList()))
	// flag.StringVar(&CineColSeries, "cine_series", "series", "If cinemanifest is provided, this value represents the name of the column that indicates the series grouping.")
}

func initializeCINEManifest(manMap map[string]struct{}) error {
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

			if len(head) < 4 {
				return fmt.Errorf("CINE manifest must be tab-delimited with at least 4 columns: series, zip_file, dicom_file, instance_number")
			}

			continue
		}

		zipKey := cineZip(line[head[CineColZip]])
		seriesKey := cineSeriesID(line[head[CineColSeries]])

		// To save memory, don't add a sample to the CINE manifest if they are
		// not in the manifest of samples to annotate.
		if _, exists := manMap[line[head[CineColZip]]]; !exists {
			continue
		}

		intInstanceNumber, err := strconv.Atoi(line[head[CineColInstancNumber]])
		if err != nil {
			// Ignore the error
			intInstanceNumber = 0
		}
		value := cineValue{Dicom: line[head[CineColDicom]], InstanceNumber: intInstanceNumber, Series: seriesKey}

		zipMap, exists := cineLookup[zipKey]
		seriesList := zipMap[seriesKey]
		seriesList = append(seriesList, value)

		if !exists {
			zipMap = make(map[cineSeriesID][]cineValue)
		}

		zipMap[seriesKey] = seriesList
		cineLookup[zipKey] = zipMap
	}

	return nil
}

func CINEFetchDicomNames(manMap map[string]struct{}, Zip, Series string) ([]string, error) {

	cineMutex.RLock()
	if len(cineLookup) == 0 {
		cineMutex.RUnlock()
		cineMutex.Lock()

		// Make sure that it wasn't changed while we were waiting for the lock
		if len(cineLookup) == 0 {
			if err := initializeCINEManifest(manMap); err != nil {
				cineMutex.Unlock()
				return nil, err
			}
		}
		cineMutex.Unlock()
		cineMutex.RLock()
	}

	defer cineMutex.RUnlock()

	zipMap := cineLookup[cineZip(Zip)]
	value := zipMap[cineSeriesID(Series)]

	sort.Slice(value, func(i, j int) bool { return value[i].InstanceNumber < value[j].InstanceNumber })

	out := make([]string, 0, len(value))

	for _, dicom := range value {
		out = append(out, dicom.Dicom)
	}

	return out, nil
}

// CINEFetchAllDicomNames yields the ordered list of DICOM names, and the
// ordered list of matching series names
func CINEFetchAllDicomNames(manMap map[string]struct{}, Zip string) ([]string, []string, error) {

	cineMutex.RLock()
	if len(cineLookup) == 0 {
		cineMutex.RUnlock()
		cineMutex.Lock()

		// Make sure that it wasn't changed while we were waiting for the lock
		if len(cineLookup) == 0 {
			if err := initializeCINEManifest(manMap); err != nil {
				cineMutex.Unlock()
				return nil, nil, err
			}
		}
		cineMutex.Unlock()
		cineMutex.RLock()
	}

	defer cineMutex.RUnlock()

	value := make([]cineValue, 0)

	zipMap := cineLookup[cineZip(Zip)]

	for _, seriesDicoms := range zipMap {
		value = append(value, seriesDicoms...)
	}

	sort.Slice(value, func(i, j int) bool { return value[i].InstanceNumber < value[j].InstanceNumber })

	out := make([]string, 0, len(value))
	outSeries := make([]string, 0, len(value))

	for _, dicom := range value {
		out = append(out, dicom.Dicom)
		outSeries = append(outSeries, string(dicom.Series))
	}

	return out, outSeries, nil
}
