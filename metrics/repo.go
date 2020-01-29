package metrics

import (
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/influxdata/influxdb1-client"
	influx "github.com/influxdata/influxdb1-client/v2"
	log "github.com/sirupsen/logrus"
)

const repoKind = "repo"

type Repo struct {
	Forks    int64  `json:"forks_count"`
	Issues   int64  `json:"open_issues_count"`
	Name     string `json:"name"`
	Org      string `json:"org,omitempty"`
	Stars    int64  `json:"stargazers_count"`
	Watchers int64  `json:"subscribers_count"`
}

func (r *Repo) printJson() {
	v := map[string]interface{}{
		"forks":    r.Forks,
		"issues":   r.Issues,
		"kind":     repoKind,
		"name":     r.Name,
		"org":      r.Org,
		"stars":    r.Stars,
		"watchers": r.Watchers,
	}
	j, err := json.Marshal(v)
	if err != nil {
		log.Error("json", err)
	}

	fmt.Println(string(j))

}

func (r *Repo) printCSV() {
	fmt.Printf("%d%s%d%s%s%s%s%s%s%s%d%s%d\n",
		r.Forks, csvSeparator,
		r.Issues, csvSeparator,
		repoKind, csvSeparator,
		r.Name, csvSeparator,
		r.Org, csvSeparator,
		r.Stars, csvSeparator,
		r.Watchers)
}

func (r *Repo) printInflux() {
	p := r.getPoint()
	fmt.Println(p[0].String())
}

func (r *Repo) getPoint() []influx.Point {
	n := repoKind

	out := make([]influx.Point, 1)
	v := map[string]interface{}{
		"forks":    r.Forks,
		"issues":   r.Issues,
		"stars":    r.Stars,
		"watchers": r.Watchers,
	}
	t := map[string]string{
		"name": r.Name,
		"org":  r.Org,
	}

	m, err := influx.NewPoint(n, t, v, time.Now())
	if err != nil {
		log.Warn(err)
	}

	out[0] = *m

	return out
}
