package main

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/rancher/github-metrics/metrics"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	Version  = "dev"
	released = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)
)

func main() {
	log.SetOutput(colorable.NewColorableStdout())

	if err := mainErr(); err != nil {
		log.Fatal(err)
	}
}

func mainErr() error {
	app := cli.NewApp()
	app.Name = "github-metrics"
	app.Version = Version
	app.Usage = "github-metrics [OPTIONS]"
	app.Before = func(ctx *cli.Context) error {
		if ctx.GlobalBool("debug") {
			log.SetLevel(log.DebugLevel)
			if !released.MatchString(app.Version) {
				log.Warnf("This is not an officially supported version (%s) of github-metrics. Please download the latest official release at https://github.com/rancher/github-metrics/releases/latest", app.Version)
			}
		}
		if len(ctx.GlobalString("org")) == 0 {
			return fmt.Errorf("Github organization is required")
		}
		if len(ctx.GlobalString("repo")) == 0 {
			return fmt.Errorf("Github repository is required")
		}
		if len(ctx.GlobalString("username")) == 0 || len(ctx.GlobalString("token")) == 0 {
			return fmt.Errorf("Github user and token are required")
		}
		match := ctx.GlobalString("match")
		if match != "sha" && match != "binary" && match != "all" {
			return fmt.Errorf("%s match not supported - Supported match [sha, binary, all]", match)
		}

		output := ctx.GlobalString("output")
		if output != "csv" && output != "json" && output != "influx" {
			return fmt.Errorf("%s output not supported - Supported output [csv, json, influx]", output)
		}

		interval := ctx.GlobalString("interval")
		_, err := time.ParseDuration(interval)
		if len(interval) == 0 || err != nil {
			return fmt.Errorf("Interval must be a valid GoLang duration string")
		}

		influxdb := ctx.GlobalString("influxdb")
		influxurl := ctx.GlobalString("influxurl")
		if output == "influx" {
			if len(influxdb) == 0 || len(influxurl) == 0 {
				return fmt.Errorf("Check your influxdb and/or influxurl params.")
			}
		}
		return nil
	}
	app.Author = "Rancher Labs, Inc."
	app.Email = ""
	app.Action = getGithubMetrics
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug,d",
			Usage: "Debug logging",
		},
		cli.BoolFlag{
			Name:  "once",
			Usage: "print stats to stdout once and exit",
		},
		cli.StringFlag{
			Name:  "interval",
			Usage: "reporting interval",
			Value: "90s",
		},
		cli.StringFlag{
			Name:   "org, o",
			EnvVar: "GITHUB_ORG",
			Usage:  "Github organization to get metrics",
			Value:  "rancher",
		},
		cli.StringFlag{
			Name:   "repo, r",
			EnvVar: "GITHUB_REPO",
			Usage:  "Github repo to get metrics",
			Value:  "",
		},
		cli.StringFlag{
			Name:   "username, u",
			EnvVar: "GITHUB_USERNAME",
			Usage:  "Github username to authenticate as",
			Value:  "",
		},
		cli.StringFlag{
			Name:   "token, t",
			EnvVar: "GITHUB_TOKEN",
			Usage:  "Github personal access token to authenticate with",
			Value:  "",
		},
		cli.StringFlag{
			Name:  "csv_sep",
			Usage: "CSV output separator",
			Value: ",",
		},
		cli.StringFlag{
			Name:  "match",
			Usage: "Which kinds of files to match [sha, binary, all]",
			Value: "binary",
		},
		cli.BoolFlag{
			Name:  "minor, m",
			Usage: "Combine equivalent minor versions, e.g. 1.0.0 and 1.0.1 into 1.0.x",
		},
		cli.StringFlag{
			Name:  "output",
			Usage: "Which output format [csv, json, influx]",
			Value: "json",
		},
		cli.BoolFlag{
			Name:  "patch, p",
			Usage: "Combine equivalent patch versions, e.g. 1.0.0-rc1 and 1.0.0 into 1.0.0",
		},
		cli.BoolFlag{
			Name:  "prerelease, prereleases",
			Usage: "Include prereleases",
		},
		cli.BoolFlag{
			Name:  "insecure",
			Usage: "Insecure connection",
		},
		cli.BoolFlag{
			Name:  "preview",
			Usage: "Just print output to stdout",
		},
		cli.StringFlag{
			Name:  "influxurl",
			Usage: "Influx url connection",
			Value: "http://localhost:8086",
		},
		cli.StringFlag{
			Name:  "influxdb",
			Usage: "Influx db name",
			Value: "telemetry",
		},
		cli.StringFlag{
			Name:  "influxuser",
			Usage: "Influx username",
			Value: "",
		},
		cli.StringFlag{
			Name:  "influxpass",
			Usage: "Influx password",
			Value: "",
		},
		cli.IntFlag{
			Name:  "batch",
			Usage: "Influx batch size",
			Value: 2000,
		},
		cli.IntFlag{
			Name:  "flush",
			Usage: "Influx flush every seconds",
			Value: 60,
		},
	}
	return app.Run(os.Args)
}

func getGithubMetrics(ctx *cli.Context) error {
	metrics := metrics.NewMetrics(metrics.NewConfig(ctx))
	metrics.GetData()

	return nil
}
