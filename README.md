[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/arturmon/s3MediaStreamer/main.yml?branch=main)](https://github.com/arturmon/s3MediaStreamer/actions/workflows/main.yml?query=branch%3Amain)
![Supported Go Versions](https://img.shields.io/badge/Go-%201.21%2C%201.22%2C%201.23-lightgrey.svg)
[![Coverage Status](https://coveralls.io/repos/github/arturmon/s3MediaStreamer/badge.svg?branch=main)](https://coveralls.io/github/arturmon/s3MediaStreamer?branch=main)
[![Docker](https://img.shields.io/docker/pulls/arturmon/s3stream)](https://hub.docker.com/r/arturmon/s3stream)
[![Docker Stars](https://badgen.net/docker/stars/arturmon/s3stream?icon=docker&label=stars)](https://hub.docker.com/r/arturmon/s3stream)
[![Docker Image Size](https://badgen.net/docker/size/arturmon/s3stream?icon=docker&label=image%20size)](https://hub.docker.com/r/arturmon/s3stream)
![Github issues](https://img.shields.io/github/issues/arturmon/s3MediaStreamer)

## Intro

[Docs Pages](https://arturmon.github.io/s3MediaStreamer/)

## Docs

[README.md](docs/README.md 'README.md')

## Infra Repo

[s3MediaStreamer-Infra](https://github.com/arturmon/s3MediaStreamer-Infra)

## Env Repo
[s3MediaStreamer-env](https://github.com/arturmon/s3MediaStreamer-env)

## Important !!!
1. When deploying to Rancher desktop, it is mandatory to add events to the bucket
2. if you use graylog, you need to manually add the GELF udp input

## Local Kuberntes
use Rancher desktop and devspave usage

## Run Devspace
```shell
export DEVSPACE_CONFIG=/home/amudrykh/GolandProjects/s3MediaStreamer/devspace.yaml
devspace use namespace media & devspace deploy
```

```shell
export DEVSPACE_CONFIG=/home/amudrykh/GolandProjects/s3MediaStreamer/devspace.yaml
kubectl port-forward service/redis-master 6379:6379 -n database > /dev/null 2>&1 & \
kubectl port-forward service/postgresql 5432:5432 -n database > /dev/null 2>&1 & \
kubectl port-forward service/postgresql 2345:2345 -n database > /dev/null 2>&1 & \
devspace use namespace media & devspace dev
```

### Devspace container command

- Run DLV `app_debug`
- Run `app_run`
- Run build `app_build`
- Synchronize `app_sync`
- View all aliases use: `alias`
