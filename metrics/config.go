package metrics

import (
	"time"

	"github.com/urfave/cli"
)

const githubAPIURL = "https://api.github.com/repos"

// Config struct
type Config struct {
	batch      int
	csvSep     string
	flush      int
	insecure   bool
	influxurl  string
	influxdb   string
	influxuser string
	influxpass string
	match      string
	minor      bool
	once       bool
	org        string
	output     string
	patch      bool
	prerelease bool
	preview    bool
	repo       string
	token      string
	url        string
	user       string
	interval   time.Duration
}

// NewConfig function
func NewConfig(ctx *cli.Context) *Config {
	interval, _ := time.ParseDuration(ctx.GlobalString("interval"))

	return &Config{
		batch:      ctx.GlobalInt("batch"),
		csvSep:     ctx.GlobalString("csv_sep"),
		flush:      ctx.GlobalInt("flush"),
		insecure:   ctx.GlobalBool("insecure"),
		interval:   interval,
		influxurl:  ctx.GlobalString("influxurl"),
		influxdb:   ctx.GlobalString("influxdb"),
		influxuser: ctx.GlobalString("influxuser"),
		influxpass: ctx.GlobalString("influxpass"),
		match:      ctx.GlobalString("match"),
		minor:      ctx.GlobalBool("minor"),
		once:       ctx.GlobalBool("once"),
		patch:      ctx.GlobalBool("patch"),
		prerelease: ctx.GlobalBool("prereleases"),
		preview:    ctx.GlobalBool("preview"),
		output:     ctx.GlobalString("output"),
		org:        ctx.GlobalString("org"),
		repo:       ctx.GlobalString("repo"),
		token:      ctx.GlobalString("token"),
		url:        githubAPIURL,
		user:       ctx.GlobalString("username"),
	}
}
