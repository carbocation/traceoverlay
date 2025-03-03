package main

import (
	"encoding/csv"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/carbocation/genomisc/ukbb/bulkprocess"
)

type Manifest struct {
	Zip                   string
	Dicom                 string
	Series                string
	InstanceNumber        int
	HasOverlayFromProject bool
	HasAutomatedOverlay   bool
	Metadata              map[string]string
}

func (m Manifest) OverlayFilename() string {
	return overlayFilename(m.Dicom)
}

func overlayFilename(dicom string) string {
	if strings.HasSuffix(dicom, ".png") {
		return dicom + ".mask.png"
	}

	return dicom + ".png.mask.png"
}

// TODO: Pick a smarter algorithm here
func UpdateManifest() error {
	global.m.Lock()
	defer global.m.Unlock()

	// First, look in the project directory to get updates to annotations.
	files, err := ioutil.ReadDir(filepath.Join(global.Config.LabelPath))
	if os.IsNotExist(err) {
		// Not a problem
	} else if err != nil {
		return err
	}
	overlaysExist := make(map[string]struct{})
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		overlaysExist[f.Name()] = struct{}{}
	}

	updatedManifest := global.manifest

	// Toggle to the latest knowledge about the manifest
	for i, v := range updatedManifest {
		_, hasOverlay := overlaysExist[overlayFilename(v.Dicom)]
		updatedManifest[i].HasOverlayFromProject = hasOverlay
	}

	return nil
}

// ReadManifest takes the path to a manifest file and extracts each line.
func ReadManifest(manifestPath, labelPath, automatedLabelPath, imagePath string) ([]Manifest, error) {
	// First, look in the labelPath to see if there are any annotations.
	files, err := ioutil.ReadDir(filepath.Join(labelPath))
	if os.IsNotExist(err) {
		// Not a problem
	} else if err != nil {
		return nil, err
	}
	overlaysExist := make(map[string]struct{})
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		overlaysExist[f.Name()] = struct{}{}
	}

	// Also look in the automatedLabelPath to see if there are any annotations
	// in the automated folder.
	files, err = ioutil.ReadDir(filepath.Join(automatedLabelPath))
	if os.IsNotExist(err) {
		// Not a problem
	} else if err != nil {
		return nil, err
	}
	automatedOverlaysExist := make(map[string]struct{})
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		automatedOverlaysExist[f.Name()] = struct{}{}
	}

	var recs [][]string

	if manifestPath == "" {
		// No manifest - just read the image directory contents

		recs = make([][]string, 0)
		recs = append(recs, []string{"dicom_file"})

		if strings.HasPrefix(imagePath, "gs://") {
			client, err := getGSClient()
			if err != nil {
				return nil, err
			}
			filenames, err := bulkprocess.ListFromGoogleStorage(imagePath, client)
			if err != nil {
				return nil, err
			}

			for _, filename := range filenames {
				recs = append(recs, []string{filename})
			}
		} else {
			files, err := ioutil.ReadDir(filepath.Join(imagePath))
			if os.IsNotExist(err) {
				// Not a problem
			} else if err != nil {
				return nil, err
			}
			for _, f := range files {
				if f.IsDir() {
					continue
				}
				recs = append(recs, []string{f.Name()})
			}
		}

	} else {
		// Read the manifest
		f, err := os.Open(manifestPath)
		if err != nil {
			return nil, err
		}

		cr := csv.NewReader(f)
		cr.Comma = '\t'
		recs, err = cr.ReadAll()
		if err != nil {
			return nil, err
		}
	}

	output := make([]Manifest, 0, len(recs))

	var headers []string
	headerIndices := struct {
		Zip            int
		Dicom          int
		Series         int
		InstanceNumber int
	}{}

	for i, cols := range recs {
		if i == 0 {
			// Capture header row.
			headers = cols
			for j, col := range cols {
				switch col {
				case "zip_file":
					headerIndices.Zip = j
				case "dicom_file":
					headerIndices.Dicom = j
				case "series":
					headerIndices.Series = j
				case "instance_number":
					headerIndices.InstanceNumber = j
				}
			}
			continue
		}

		intInstance, err := strconv.Atoi(cols[headerIndices.InstanceNumber])
		if err != nil {
			intInstance = 0
		}

		_, hasOverlay := overlaysExist[overlayFilename(cols[headerIndices.Dicom])]
		_, hasAutomatedOverlay := automatedOverlaysExist[overlayFilename(cols[headerIndices.Dicom])]

		// Build metadata from columns not explicitly handled.
		metadata := make(map[string]string)
		for j, colName := range headers {
			if j != headerIndices.Zip &&
				j != headerIndices.Dicom &&
				j != headerIndices.Series &&
				j != headerIndices.InstanceNumber {
				metadata[colName] = cols[j]
			}
		}

		output = append(output, Manifest{
			Zip:                   cols[headerIndices.Zip],
			Dicom:                 cols[headerIndices.Dicom],
			Series:                cols[headerIndices.Series],
			InstanceNumber:        intInstance,
			HasOverlayFromProject: hasOverlay,
			HasAutomatedOverlay:   hasAutomatedOverlay,
			Metadata:              metadata, // assign the extra metadata
		})
	}

	// Don't sort the manifest - use the manifest line order as the sort order

	return output, nil
}
