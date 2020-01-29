# github-metrics

Utility to get github repo, releases and assets metrics.

## Building

`make`

## Running

`./bin/github-metrics`

## Usage

github-metrics get repo, releases and assets from github api and send them to a influx in order to be explored by grafana. 

It get data every interval and send metrics every flush seconds or batch size. 

```
NAME:
   github-metrics - github-metrics [OPTIONS]

USAGE:
   github-metrics [global options] command [command options] [arguments...]

VERSION:
   dev

AUTHOR(S):
   Rancher Labs, Inc.

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --debug, -d                  Debug logging
   --once                       print stats to stdout once and exit
   --interval value             reporting interval (default: "90s")
   --org value, -o value        Github organization to get metrics (default: "rancher") [$GITHUB_ORG]
   --repo value, -r value       Github repo to get metrics [$GITHUB_REPO]
   --username value, -u value   Github username to authenticate as [$GITHUB_USERNAME]
   --token value, -t value      Github personal access token to authenticate with [$GITHUB_TOKEN]
   --csv_sep value              CSV output separator (default: ",")
   --match value                Which kinds of files to match [sha, binary, all] (default: "binary")
   --minor, -m                  Combine equivalent minor versions, e.g. 1.0.0 and 1.0.1 into 1.0.x
   --output value               Which output format [csv, json, influx] (default: "json")
   --patch, -p                  Combine equivalent patch versions, e.g. 1.0.0-rc1 and 1.0.0 into 1.0.0
   --prerelease, --prereleases  Include prereleases
   --insecure                   Insecure connection
   --preview                    Just print output to stdout
   --influxurl value            Influx url connection (default: "http://localhost:8086")
   --influxdb value             Influx db name (default: "telemetry")
   --influxuser value           Influx username
   --influxpass value           Influx password
   --batch value                Influx batch size (default: 2000)
   --flush value                Influx flush every seconds (default: 60)
   --help, -h                   show help
   --version, -v                print the version
```

NOTE: You need influx already installed and running. The influx db would be created if doesn't exist.

## Metrics

Metrics can be get on distinct formats

* csv

```
forks,issues,kind,name,org,stars,watchers
797,386,repo,k3s,rancher,11323,239
asset,downloads,kind,name,org,repo
k3s,2244,release,v1.17.2+k3s1,rancher,k3s
```

* json

```
{"forks":797,"issues":386,"kind":"repo","name":"k3s","org":"rancher","stars":11323,"watchers":239}
{"asset":"k3s","downloads":2253,"kind":"release","name":"v1.17.2+k3s1","org":"rancher","repo":"k3s"}
```

* influx

```
repo,name=k3s,org=rancher forks=797i,issues=386i,stars=11323i,watchers=239i 1580329221960838000
release,asset=k3s,name=v1.17.2+k3s1,org=rancher,repo=k3s downloads=2252i 1580329224452236000
```

## License
Copyright (c) 2019 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
