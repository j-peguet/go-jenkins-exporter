# Go-jenkins-exporter

go-jenkins-exporter is an exporter for Prometheus, which allows you to monitor/alert on multiple jenkins job statuses and properties.

## Installation

### Using source code

```shell
go get -u -v github.com/goodbins/go-jenkins-exporter
cd $GOPATH/src/github.com/goodbins/go-jenkins-exporter
make deps
make install
```

You can create your local Docker image using:

```shell
make image
```

### Using Docker

```shell
docker pull goodbins/go-jenkins-exporter:latest
docker run -it -p 5000:5000 goodbins/go-jenkins-exporter:latest
```

## Usage

Say you have a jenkins instance at http://jenkins-ci:8080

First thing to do is to export some env vars:

```shell
export JENKINS_USERNAME=yourusername
export JENKINS_PASSWORD=yourpassword
```

Note: You can also use a token instead of a password.

Then you can launch the exporter using the following command:

```shell
./go-jenkins-exporter -j jenkins-ci:8080 -r 2s
```

By default, go-jenkins-exporter listens at [localhost:5000](localhost:5000)

Using the public registry Docker image:

```shell
docker run -it \
    -p 5000:5000 \
    -e JENKINS_USERNAME=yourusername \
    -e JENKINS_PASSWORD=yourpassword \
    --restart=unless-stopped \
    goodbins/go-jenkins-exporter:latest -j jenkins-ci:8080 -r 2s
```

For more configuration options you can use:

```shell
./go-jenkins-exporter --help
```

This gives something like:

```console
Usage:
  go-jenkins-exporter [flags]

Flags:
  -h, --help               help for go-jenkins-exporter
  -j, --jenkins string     Jenkins API host:port pair
  -l, --listen string      Exporter host:port pair (default "localhost:5000")
  -m, --metrics string     Path under which to expose metrics (default "/metrics")
  -a, --path string        Jenkins API path (default "/api/json")
  -r, --rate duration      Set metrics update rate in seconds (default 1s)
  -s, --ssl                Enable TLS (default false)
  -t, --timeout duration   Jenkins API timeout in seconds (default 10s)
  -v, --verbose            Enable verbosity
      --version            version for go-jenkins-exporter
```

## Prometheus configuration

You can add the endpoint to your prometheus.yml file:

```yaml
scrape_configs:
  - job_name: 'go-jenkins-exporter'
    scrape_interval: 5s
    static_configs:
      - targets: ['localhost:5000']
```

## Metrics
The exporter returns the following data:
* the last build (jenkins_job_last_build_xxx)
* the last successful build (jenkins_job_last_successful_build_xxx)
* the last failed build (jenkins_job_last_failed_build_xxx)
* the last stable build (jenkins_job_last_stable_build_xxx)
* the last unstable build (jenkins_job_last_unstable_build_xxx)
* the last completed build (jenkins_job_last_completed_build_xxx)
* the last not completed build (jenkins_job_last_unsuccessful_build_xxx)

For each type of build, the following metrics are displayed:
* The build number
* The result
* The color (color) *Only for the last build (jenkins_job_last_build_xxx)* 
* The cause of the trigger (cause)
* Build duration (duration_seconds)
* Build timestamp (timestamp_seconds)
* Waiting time (queuing_duration_seconds) *Only for Jenkins v2 API*
* Total duration (total_duration_seconds) *Only for Jenkins v2 API*

### Corresponding values
Prometheus imposes a digital data format. A code has therefore been put in place to determine these states.

#### Result
The build result is available for all builds.
| Prometheus Value | Corresponding |
| - | - |
| 0   | Failure |
| 0.5 | Unstable |
| 1   | Success |
| 3   | Not Build |
| 4   | Running |
| 100 | Unknown (default) |

#### Color
The colour is only available for the last build of each job. This provides additional information to the "Result", such as whether the job has been deactivated.
| Prometheus Value | Corresponding |
| - | - |
| 0   | Blue |
| 1   | Red |
| 2   | Yellow |
| 3   | Not Build |
| 4   | Disabled |
| 5   | Aborted |
| 6   | Gray |
| 100 | Unknown (default) |

#### Cause
The cause determines which action triggered the build. There are many cases, so don't hesitate to make a pull request to add to the list.
| Prometheus Value | Started by |
| - | - |
| 0   | Timer |
| 1   | User |
| 2   | Upstream |
| 3   | SCM Change |
| 4   | Branch Indexing |
| 5   | Gitlab Webhook |
| 6   | Command line |
| 7   | Remote Host |
| 8   | Replayed |
| 9   | Restarted |
| 10  | Git Action (Push, Merge Request) |
| 100 | Unknown (default) |
| -1  | Value if the API not provide this info |

Note: Due to certain plugins or api versions, all the above data may not be available.
## Tested version

List of Jenkins API versions tested:
* 1.625.3
* 1.642.1
* 2.7.3
* 2.19.1
* 2.11
* 2.23
* 2.32.3
* 2.73.1
* 2.107.3
* 2.138.3
* 2.164.3
* 2.179
* 2.190.2
* 2.204.1
* 2.222.1
* 2.222.3
* 2.233
* 2.249.1
* 2.263.2
* 2.263.4
* 2.303.2
* 2.319.1

## Licence
Unless otherwise noted, the go-jenkins-exporter source files are distributed under the MIT license found in the LICENSE file.

## Next steps...

 - Using bndr/gojenkins to interact with Jenkins API
 - Expose the metrics of the slave nodes
 - Create a helm chart to deploy the exporter on k8s
 - write unit tests
 
## Contribute
Go to [contributing.md](CONTRIBUTING.md)