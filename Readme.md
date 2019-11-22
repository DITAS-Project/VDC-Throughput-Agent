# DITAS - VDC Throughput Agent

The VDC Throughput Agent is one of the monitoring sidecars used to observe the behavior of VDCs within the DITAS project. The agent observes all incoming and outgoing requests from a VDC by observing the underlying socket layer. The data is aggregated over time and send to the monitoring database.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

### Prerequisites

To use this component, you need at least go 1.10 and dep 0.5 and [pktstat](https://github.com/dleonard0/pktstat).

To install the go lang tools go to: [Go Getting Started](https://golang.org/doc/install)



### Installing

For installation you have two options, building and running it on your local machine or using the docker approach.

For local testing and building for that you can follow the following steps:

install dependencies (only needs to be done once):

compile
```
CGO_ENABLED=0 GOOS=linux go build -a --installsuffix cgo --ldflags="-w -s -X main.Build=$(git rev-parse --short HEAD)" -o thr-agnt
```

to run locally:
```
./thr-agnt
```

For the docker approach, you can use the provided dockerfile to build a running artifact as a Docker container.

build the docker container:
```
docker build -t ditas/throughput-agent -f Dockerfile.artifact . 
```

Attach the docker container to a VDC or other microservice like component:
```
docker run -v ./traffic.json:/opt/blueprint/traffic.json --pid=container:<APPID> ditas/throughput-agent
```
Here `<APPID>` must be the container ID of the application you want to observe. Also, refer to the **Configuration** section for information about the `traffic.json`-config file.

## Running the tests

For testing you can use:
```
 go test ./...
```

For that make sure you have [pktstat](https://github.com/dleonard0/pktstat) installed and an elastic search running locally at the default port.


## Configuration
To configure the agent, you can specify the following values in a JSON file:
 * ElasticSearchURL => The URL that all aggregated data is sent to
 * ElasticBasicAuth => boolean to indicate if authentication for the elastic is required
 * ElasticUser => username for the elasticsearch
 * ElasticPassword => password for the elasticsearch
 * trace => boolean if every message should be logged to the console
 * VDCName => the Name used to store the information under
 * windowTime => the time window that is used to aggregate connections in seconds
 * ignore => List of `ip:port`-data that should not be aggregated or reported
 * components => map to name connections, e.g., `*.:3306:"database server`.
 * verbose => boolean to indicate if the agent should use verbose logging (recommended for debugging)

An example file could look like this:
```
{
    "ElasticSearchURL":"http://127.0.0.1:9200",
    "ElasticBasicAuth": true,
    "ElasticUser": "user",
    "ElasticPassword": "123456",
    "trace": false,
    "VDCName":"tubvdc",
    "windowTime":10,
    "ignore":["(.*):9200""],
    "components":{".*:3306":"mysql",".*:9042":"cassandra"},
    "verbose":true
}   
```

Alternatively, use can use flags with the same name to configure the agent.

## Built With

* [viper](https://github.com/spf13/viper)
* [pktstat](https://github.com/dleonard0/pktstat)
* [ElasticSearch](https://www.elastic.co/)

## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/your/project/tags). 

## License

This project is licensed under the Apache 2.0 - see the [LICENSE.md](LICENSE.md) file for details.

## Acknowledgments

This is being developed for the [DITAS Project](https://www.ditas-project.eu/)
