package metrics

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	// Blank import required by vendor
	_ "github.com/influxdata/influxdb1-client"
	influx "github.com/influxdata/influxdb1-client/v2"
	log "github.com/sirupsen/logrus"
)

const (
	releaseKind  = "release"
	releasePatch = "patch"
	releaseMinor = "minor"
)

// Asset struct
type Asset struct {
	ContentType string `json:"content_type"`
	Downloads   int64  `json:"download_count"`
	Name        string `json:"name"`
}

// Release struct
type Release struct {
	Assets     *[]Asset `json:"assets"`
	Draft      bool     `json:"draft"`
	Name       string   `json:"tag_name"`
	Org        string   `json:"org"`
	Prerelease bool     `json:"prerelease"`
	Repo       string   `json:"repo,omitempty"`
}

func (r *Release) printJSON() {
	for _, asset := range *r.Assets {
		v := map[string]interface{}{
			"asset":     asset.Name,
			"downloads": asset.Downloads,
			"kind":      releaseKind,
			"name":      r.Name,
			"org":       r.Org,
			"repo":      r.Repo,
		}
		j, err := json.Marshal(v)
		if err != nil {
			log.Error("json", err)
		}

		fmt.Println(string(j))
	}
}

func (r *Release) printCSV() {
	for _, asset := range *r.Assets {
		fmt.Printf("%s%s%d%s%s%s%s%s%s%s%s\n",
			asset.Name, csvSeparator,
			asset.Downloads, csvSeparator,
			releaseKind, csvSeparator,
			r.Name, csvSeparator,
			r.Org, csvSeparator,
			r.Repo)
	}
}

func (r *Release) printInflux() {
	points := r.getPoint()
	for _, p := range points {
		fmt.Println(p.String())
	}
}

func (r *Release) getPoint() []influx.Point {
	n := releaseKind
	out := make([]influx.Point, len(*r.Assets))
	for index, asset := range *r.Assets {
		v := map[string]interface{}{
			"downloads": asset.Downloads,
		}
		t := map[string]string{
			"asset": asset.Name,
			"name":  r.Name,
			"org":   r.Org,
			"repo":  r.Repo,
		}

		m, err := influx.NewPoint(n, t, v, time.Now())
		if err != nil {
			log.Warn(err)
			continue
		}
		out[index] = *m
	}

	return out
}

func (r *Release) getName(option string) string {
	nameFormat, err := regexp.Compile("^(v)([0-9]+)\\.([0-9]+)\\.([0-9]+).*$")
	if err != nil {
		log.Error("Error checking minor format ", err)
	}

	name := r.Name
	if option == releasePatch && nameFormat.MatchString(name) {
		name = nameFormat.ReplaceAllString(name, "$1$2.$3.$4")
	}
	if option == releaseMinor && nameFormat.MatchString(name) {
		name = nameFormat.ReplaceAllString(name, "$1$2.$3.x")
	}

	return name
}

func (r *Release) filterAssets(match string) {
	if match == "all" {
		return
	}

	shaFormat, err := regexp.Compile("^sha([0-9]+)sum.*")
	if err != nil {
		log.Error("Error checking sha format ", err)
	}

	newAsset := []Asset{}
	for _, asset := range *r.Assets {
		if match == "binary" && asset.ContentType != "application/octet-stream" {
			continue
		}
		if match == "sha" && !shaFormat.MatchString(asset.Name) {
			continue
		}
		newAsset = append(newAsset, asset)
	}
	r.Assets = &newAsset
}

func (r *Release) aggregateAssets(rel *Release) {
	newAsset := *r.Assets
	for _, relAsset := range *rel.Assets {
		found := false
		for i := range newAsset {
			if newAsset[i].Name == relAsset.Name {
				newAsset[i].Downloads += relAsset.Downloads
				found = true
				break
			}
		}
		if !found {
			newAsset = append(newAsset, relAsset)
		}
	}
	r.Assets = &newAsset
}
