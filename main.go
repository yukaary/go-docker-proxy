package main

import (
	"flag"
	"github.com/golang/glog"
	be "github.com/yukaary/go-docker-proxy/backend"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"
)

var (
	backend *be.Backend
	mutex   sync.Mutex
)

func main() {
	glog.Infof("starting proxy.")

	etcd_endpoint := flag.String("etcd-endpoint", "http://127.0.0.1:4001", "etcd url which contains service information.")
	scheme := flag.String("scheme", "http", "scheme, maybe http/https. but currently only support http.")
	backend_port := flag.String("backend-port", "80", "backend port to be listened")
	service_name := flag.String("service-name", "", "service name to be blanced.")
	flag.Parse()

	backend = be.NewBackend(*etcd_endpoint, *scheme, *backend_port, *service_name)
	backend.Fetch()

	// load balancing.
	director := func(request *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()
		// switch app configuration
		backend.SwitchApps()
		request.URL.Scheme = backend.Scheme()
		request.URL.Host = backend.NextHost()
		glog.Infof("redirect to %s", request.URL.String())
	}
	proxy := &httputil.ReverseProxy{Director: director}
	server := http.Server{
		Addr:    ":80",
		Handler: proxy,
	}

	// update service information per 1 second.
	go func() {
		for {
			time.Sleep(1 * time.Second)
			backend.Fetch()
		}
	}()
	server.ListenAndServe()
}
