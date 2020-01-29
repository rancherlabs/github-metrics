package metrics

import (
	"time"

	// Blank import required by vendor
	_ "github.com/influxdata/influxdb1-client"
	influx "github.com/influxdata/influxdb1-client/v2"
	log "github.com/sirupsen/logrus"
)

func check(e error, m string) {
	if e != nil {
		log.Error("[Error]: ", m, e)
	}
}

// Influx struct
type Influx struct {
	url     string
	db      string
	user    string
	pass    string
	cli     influx.Client
	batch   influx.BatchPoints
	timeout time.Duration
}

func newInflux(u, d, us, pa string) *Influx {
	var a = &Influx{
		url:  u,
		db:   d,
		user: us,
		pass: pa,
	}

	a.timeout = time.Duration(10)
	return a
}

func (i *Influx) check(retry int) bool {
	respTime, _, err := i.cli.Ping(i.timeout)
	if err != nil {
		log.Error("[Error]: ", err)
		log.Error("Influx disconnected...")
		connected := false
		for index := 0; index < retry && !connected; index++ {
			log.Error("Reconnecting ", index+1, " of ", retry, "...")
			connected = i.Connect()
			if !connected {
				time.Sleep(time.Duration(1) * time.Second)
			}
		}
		if err != nil {
			log.Error("Failed to connect to influx ", i.url)
			return false
		}
		log.Info("Influx response time: ", respTime)
		return true
	}
	log.Info("Influx response time: ", respTime)
	return true
}

// CheckConnect function
func (i *Influx) CheckConnect(interval int) chan bool {
	ticker := time.NewTicker(time.Second * time.Duration(interval))

	connected := make(chan bool)

	go func() {
		for {
			select {
			case <-ticker.C:
				if !i.check(2) {
					close(connected)
					return
				}
			}
		}
	}()

	return connected
}

// Connect function
func (i *Influx) Connect() bool {
	var err error
	log.Info("Connecting to Influx...")

	i.cli, err = influx.NewHTTPClient(influx.HTTPConfig{
		Addr:     i.url,
		Username: i.user,
		Password: i.pass,
	})

	if err != nil {
		log.Error("[Error]: ", err)
		return false
	}

	if i.check(0) {
		i.createDb()
		return true
	}
	return false
}

func (i *Influx) init() {
	i.newBatch()
}

// Close function
func (i *Influx) Close() {
	message := "Closing Influx connection..."
	err := i.cli.Close()
	check(err, message)
	log.Info(message)
}

func (i *Influx) createDb() {
	var err error
	log.Info("Creating Influx database if not exists...")

	comm := "CREATE DATABASE " + i.db

	q := influx.NewQuery(comm, "", "")
	_, err = i.cli.Query(q)
	if err != nil {
		log.Error("[Error] ", err)
	} else {
		log.Info("Influx database ", i.db, " created.")
	}
}

func (i *Influx) newBatch() {
	var err error
	message := "Creating Influx batch..."
	i.batch, err = influx.NewBatchPoints(influx.BatchPointsConfig{
		Database:  i.db,
		Precision: "s",
	})
	check(err, message)
	log.Info(message)

}

func (i *Influx) newPoint(m influx.Point) {
	message := "Adding point to batch..."
	fields, _ := m.Fields()
	pt, err := influx.NewPoint(m.Name(), m.Tags(), fields, m.Time())
	check(err, message)
	i.batch.AddPoint(pt)
}

func (i *Influx) newPoints(m []influx.Point) {
	log.Info("Adding ", len(m), " points to batch...")
	for index := range m {
		i.newPoint(m[index])
	}
}

func (i *Influx) Write() {
	start := time.Now()
	log.Info("Writing batch points...")

	// Write the batch
	err := i.cli.Write(i.batch)
	if err != nil {
		log.Error("[Error]: ", err)

	}

	log.Info("Time to write ", len(i.batch.Points()), " points: ", float64((time.Since(start))/time.Millisecond), "ms")
}

func (i *Influx) sendToInflux(m []influx.Point, retry int) bool {
	if i.check(retry) {
		i.init()
		i.newPoints(m)
		i.Write()
		return true
	}
	return false
}
