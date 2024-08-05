[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/arturmon/s3MediaStreamer/main.yml?branch=main)](https://github.com/arturmon/s3MediaStreamer/actions/workflows/main.yml?query=branch%3Amain)
![Supported Go Versions](https://img.shields.io/badge/Go-%201.20%2C%201.21-lightgrey.svg)
[![Coverage Status](https://coveralls.io/repos/github/arturmon/s3MediaStreamer/badge.svg?branch=main)](https://coveralls.io/github/arturmon/s3MediaStreamer?branch=main)
[![Docker](https://img.shields.io/docker/pulls/arturmon/s3stream)](https://hub.docker.com/r/arturmon/s3stream)
[![Docker Stars](https://badgen.net/docker/stars/arturmon/s3stream?icon=docker&label=stars)](https://hub.docker.com/r/arturmon/s3stream)
[![Docker Image Size](https://badgen.net/docker/size/arturmon/s3stream?icon=docker&label=image%20size)](https://hub.docker.com/r/arturmon/s3stream)
![Github issues](https://img.shields.io/github/issues/arturmon/s3MediaStreamer)

## Intro

[Docs Pages](https://arturmon.github.io/s3MediaStreamer/)

## Docs

[README.md](docs/README.md 'README.md')

## Important !!!
1. When deploying to Rancher desktop, it is mandatory to add events to the bucket
2. if you use graylog, you need to manually add the GELF udp input

## Local Kuberntes
### Setup harbor
1. wsl Ubuntu add `echo "127.0.0.1     harbor.localhost" | sudo tee -a /etc/hosts`
2. create password registry `echo -n 'user:password' | base64`
3. add registry `~/.docker/config.json`
```json
{
        "auths": {
                "http://harbor.localhost": {
                        "auth": "cm9ib3QkbGlicm****"
                },
                "https://index.docker.io/v1/": {
                        "auth": "YXJ0dXJtb246bTl***"
                }
        }
}
```
4. login `docker login harbor.localhost`