# dockerhub-limit-exporter-go

A prometheus exporter which helps you with monitoring your Docker Hub rate limits.
This is based on Michael Friedrich's docker-hub-limit-exporter (https://gitlab.com/gitlab-com/marketing/corporate_marketing/developer-evangelism/code/docker-hub-limit-exporter), it's basically a Go rewrite of the original Python program.

## Building it

Either build it with `go build` with your local go installation or build it with docker `docker build -t dockerhub-limit-exporter-go`

## Usage

It can be configured via the environment variables
* DOCKERHUB_USERNAME
* DOCKERHUB_PASSWORD
* DOCKERHUB_EXPORTER_PORT
which should be pretty much self-explanatory.

If no username or password is given, anonymous pulls are used. The default port is 8881

## Available metrics

Currently there are two metrics: `dockerhub_limit_max_requests_total` which gives you the max number of allowed requests and `dockerhub_limit_remaining_requests_total`, which gives the currently remaining number of available requests.

## Why?

I built the docker image of the aforementioned Python based exporter and found its size at around 900 MB a little hefty. Since I wanted to do a little bit of Go programming anyways, I decided to do a quick rewrite.
