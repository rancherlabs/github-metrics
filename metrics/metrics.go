package metrics

import (
	"os"
	"os/signal"
	"sync"
	"time"

	_ "github.com/influxdata/influxdb1-client"
	influx "github.com/influxdata/influxdb1-client/v2"
	log "github.com/sirupsen/logrus"
)

const (
	influxFlush = 60
	influxCheck = 3600
)

var csvSeparator string

type Metric interface {
	printJson()
	printCSV()
	printInflux()
	getPoint() []influx.Point
}

type Metrics struct {
	Input   chan Metric
	Exit    chan os.Signal
	Readers []chan struct{}
	Config  *Config
}

func NewMetrics(conf *Config) *Metrics {
	r := &Metrics{
		Readers: []chan struct{}{},
		Config:  conf,
	}

	r.Input = make(chan Metric, 1)
	r.Exit = make(chan os.Signal, 1)
	signal.Notify(r.Exit, os.Interrupt, os.Kill)

	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true

	csvSeparator = r.Config.csvSep

	return r
}

func (r *Metrics) Close() {
	close(r.Input)
	close(r.Exit)

}

func (r *Metrics) addReader() chan struct{} {
	chan_new := make(chan struct{}, 1)
	r.Readers = append(r.Readers, chan_new)

	return chan_new
}

func (r *Metrics) closeReaders() {
	for _, r_chan := range r.Readers {
		if r_chan != nil {
			r_chan <- struct{}{}
		}
	}
	r.Readers = nil
}

func (r *Metrics) GetData() {
	var in, out sync.WaitGroup
	indone := make(chan struct{}, 1)
	outdone := make(chan struct{}, 1)

	in.Add(1)
	go func() {
		defer in.Done()
		r.getRepoData(r.addReader())
	}()

	in.Add(1)
	go func() {
		defer in.Done()
		r.getReleaseData(r.addReader())
	}()

	out.Add(1)
	go func() {
		defer out.Done()
		r.getOutput()
	}()

	go func() {
		in.Wait()
		close(r.Input)
		close(indone)
	}()

	go func() {
		out.Wait()
		close(outdone)
	}()

	for {
		select {
		case <-indone:
			go r.closeReaders()
			<-outdone
			return
		case <-outdone:
			log.Error("Aborting...")
			go r.closeReaders()
			return
		case <-r.Exit:
			//close(r.Exit)
			log.Info("Exit signal detected....Closing...")
			go r.closeReaders()
			select {
			case <-outdone:
				return
			}
		}
	}
}

func (r *Metrics) getRepoData(stop chan struct{}) {
	r.getRepo()

	if r.Config.once {
		return
	}

	ticker := time.NewTicker(r.Config.interval)

	for {
		select {
		case <-ticker.C:
			log.Info("Tick on getting repo data")
			go r.getRepo()
		case <-stop:
			return
		}
	}
}

func (r *Metrics) getRepo() {
	uri := "/" + r.Config.org + "/" + r.Config.repo

	log.Infof("Getting repo data from %s...", r.Config.url+uri)

	repo := &Repo{
		Org: r.Config.org,
	}

	_, err := getJSON(r.Config.url+uri, r.Config.user, r.Config.token, r.Config.insecure, repo)
	if err != nil {
		log.Error("Error getting repo JSON from ", r.Config.url+uri, err)
	}

	r.Input <- repo
}

func (r *Metrics) getReleaseData(stop chan struct{}) {
	r.getRelease(r.addReader())

	if r.Config.once {
		return
	}

	ticker := time.NewTicker(r.Config.interval)

	for {
		select {
		case <-ticker.C:
			log.Debug("Tick on Getting release data")
			go r.getRelease(r.addReader())
		case <-stop:
			return
		}
	}
}

func (r *Metrics) getRelease(stop chan struct{}) {
	var err error
	urlChan := make(chan string, 1)

	uri := "/" + r.Config.org + "/" + r.Config.repo + "/releases"

	log.Infof("Getting release data from %s...", r.Config.url+uri)

	next := r.Config.url + uri
	urlChan <- next
	releases := &[]Release{}

	for {
		select {
		case url := <-urlChan:
			if url == "" {
				close(urlChan)
				r.filterReleases(releases)
				return
			}
			nextRel := &[]Release{}
			next, err = getJSON(url, r.Config.user, r.Config.token, r.Config.insecure, nextRel)
			if err != nil {
				log.Error("Getting release JSON from ", next, err)
			}
			*releases = append(*releases, *nextRel...)
			urlChan <- next
		case <-stop:
			return
		}
	}
}

func (r *Metrics) filterReleases(releases *[]Release) {
	for _, release := range *releases {
		release.Org = r.Config.org
		release.Repo = r.Config.repo
		if !r.Config.prerelease && release.Prerelease {
			continue
		}
		release.filterAssets(r.Config.match)
		if len(*release.Assets) > 0 {
			//log.Info(release.getName("patch"))
			input := release
			r.Input <- &input
		}
	}
}

func (r *Metrics) getOutput() {
	switch r.Config.output {
	case "json", "csv":
		r.print()
	case "influx":
		if r.Config.preview {
			r.print()
		} else {
			r.sendToInflux()
		}
	}
}

func (r *Metrics) print() {
	for {
		select {
		case metric := <-r.Input:
			if metric == nil {
				return
			}
			if r.Config.output == "json" {
				metric.printJson()
			}
			if r.Config.output == "csv" {
				metric.printCSV()
			}
			if r.Config.output == "influx" {
				metric.printInflux()
			}
		}
	}
}

func (r *Metrics) sendToInflux() {
	var points []influx.Point
	var index, p_len int

	i := newInflux(r.Config.influxurl, r.Config.influxdb, r.Config.influxuser, r.Config.influxpass)

	if i.Connect() {
		connected := i.CheckConnect(influxCheck)
		defer i.Close()

		ticker := time.NewTicker(time.Second * time.Duration(influxFlush))

		index = 0
		for {
			select {
			case <-connected:
				return
			case <-ticker.C:
				if len(points) > 0 {
					log.Debug("Tick on sending to influx")
					if i.sendToInflux(points, 1) {
						points = []influx.Point{}
					} else {
						return
					}
				}
			case p := <-r.Input:
				if p != nil {
					m := p.getPoint()
					points = append(points, m...)
					p_len = len(points)
					if p_len == r.Config.batch {
						if i.sendToInflux(points, 1) {
							points = []influx.Point{}
						} else {
							return
						}
					}
					index++
				} else {
					p_len = len(points)
					if p_len > 0 {
						if i.sendToInflux(points, 1) {
							points = []influx.Point{}
						}
					}
					return
				}
			}
		}
	}
}
