[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/arturmon/s3MediaStreamer/main.yml?branch=main)](https://github.com/arturmon/s3MediaStreamer/actions/workflows/main.yml?query=branch%3Amain)
![Supported Go Versions](https://img.shields.io/badge/Go-%201.19%2C%201.20%2C%201.21-lightgrey.svg)
[![Coverage Status](https://coveralls.io/repos/github/arturmon/s3MediaStreamer/badge.svg?branch=main)](https://coveralls.io/github/arturmon/s3MediaStreamer?branch=main)
[![Docker](https://img.shields.io/docker/pulls/arturmon/s3stream)](https://hub.docker.com/r/arturmon/s3stream)
[![Docker Stars](https://badgen.net/docker/stars/arturmon/s3stream?icon=docker&label=stars)](https://hub.docker.com/r/arturmon/s3stream)
[![Docker Image Size](https://badgen.net/docker/size/arturmon/s3stream?icon=docker&label=image%20size)](https://hub.docker.com/r/arturmon/s3stream)
![Github issues](https://img.shields.io/github/issues/arturmon/s3MediaStreamer)


## Infrastructures Diagrams
* [Core](infrastructure.md 'Infrastructure')
* [Sequence diagrams](sequence_diagrams.md 'Infrastructure Sequence diagrams')

## Core Functions

* [Core Functions](core_function.md 'Core Functions')
## Environment

* [Environment](environments_var.md 'Environment')

## Services
use docker-compose.yaml to run all the necessary components

| Service           | required | 
|-------------------|----------|
| postgresql        | [*]      |
| postgres-exporter | [-]      |
| redis             | [-]      |
| memcache          | [-]      |
| mongodb           | [-]      |
| rabbitmq          | [*]      |
| minio             | [*]      |
| minio-mc          | [*]      |
| prometheus        | [-]      |
| grafana           | [-]      |
| consul            | [*]      |

** rabbitmq exchanging messages used only for with S3


## Api Functions

* [Api Functions v1](api_functions_v1.md 'Api Functions')\
\
  Example endpoints in file  [test-api.http](https://raw.githubusercontent.com/arturmon/s3MediaStreamer/main/test-api.http 'Tests api')


## Generate SWAGGER
```shell
cd app && swag init --parseDependency --parseDepth=1
```


### DEBUG
```
http://localhost:6060/debug/pprof/
```