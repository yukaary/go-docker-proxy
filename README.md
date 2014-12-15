A Simple Implementation of proxy server for load balancing docker container.
============================================================================

This is the quite simple implementation of load balancer for container process. It assumes to cooperate with [go-docker-dns](https://github.com/yukaary/go-docker-dns) to accquire service information(IP addresses of scaled instances) via `etcd`. This proxy can direct newly instance without configuration reloading. The service information would be updated per 1 second.

__It does not support reducing instances completely.__ There are possible cases that you reduce number of instance 5 -> 3, but the proxy service still has a old configuration which knows 5 instances, and then it will redirect non-exist instance.

## Install

```
$ go get github.com/yukaary/go-docker-proxy
$ go install github.com/yukaary/go-docker-proxy
```

## Usage

```
$ /go/bin/go-docker-proxy -stderrthreshold=INFO -etcd-endpoint http://172.17.8.101:4001 -service-name frontend
```

where `etcd-endpoint` is a one of the etcd endpoint forming a cluster, `servie_name` is a head of a host name of scaled service which has a fullname like frontend_1, frontend_2, ..., frontend_N - this is the changed behaviour of [fig, better-recreate branch](https://github.com/yukaary/fig/commits/better_recreate).

### Using fig

Description in `fig.yml` looks like that.

```
goproxy:
    build: ./go-docker-proxy
    command: /go/bin/go-docker-proxy -stderrthreshold=INFO -etcd-endpoint http://172.17.8.101:4001 -service-name frontend
    ports:
        - "10080:80"
    dns:
        - "172.17.8.1"
```
