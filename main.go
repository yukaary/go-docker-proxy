package main

import (
	"container/ring"
	"flag"
	"github.com/golang/glog"
	etcdutil "github.com/yukaary/go-docker-dns/etcd"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

type Backend struct {
	serviceName string
	fetchUrl    string
	etcdcli     *etcdutil.EtcdClient
	apps        map[string]*url.URL
	scheme      string
	port        string
	currentRing *ring.Ring
	nextRing    *ring.Ring
}

var (
	etcd_endpoint *string
	backend_port  *string
	scheme        *string
	service_name  *string
	backend       *Backend
	mutex         sync.Mutex
)

func NewBackend(fetchUrl, scheme, port, serviceName string) *Backend {
	backend := &Backend{
		fetchUrl:    fetchUrl,
		serviceName: serviceName,
		scheme:      scheme,
		port:        port,
	}
	backend.initialize()
	return backend
}

func (self *Backend) initialize() {
	self.etcdcli = etcdutil.NewEtcdClient([]string{self.fetchUrl})
}

func (self *Backend) fetch() {
	self.apps = make(map[string]*url.URL)
	for _, node := range self.etcdcli.GetDir("services", self.serviceName) {
		self.apps[node.Key] = genUrl(self.scheme, node.Value, self.port)
	}
}

func (self *Backend) NewRing() *ring.Ring {
	backendRing := ring.New(len(self.apps))

	// when load balance routine (maybe it's a director function?) detects
	// new ring, it should switch it soon.

	return backendRing
}

func genUrl(scheme, body, port string) *url.URL {
	url, _ := url.Parse(scheme + body + ":" + port)
	return url
}

func main() {
	glog.Infof("starting proxy.")

	etcd_endpoint := flag.String("etcd-endpoint", "http://127.0.0.1:4001", "etcd url which contains service information.")
	//backend_port := flag.String("backend-port", "80", "backend port to be listened")
	service_name := flag.String("service-name", "", "service name to be blanced.")
	flag.Parse()

	glog.Infof("etcd_endpoint: %s", *etcd_endpoint)
	glog.Infof("service_name: %s", *service_name)

	backend = NewBackend(*etcd_endpoint, "http://", "3000", *service_name)
	backend.fetch()
	for k, v := range backend.apps {
		glog.Infof("backend app: %s %s", k, v.String())
	}

	// load balancing.
	director := func(request *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()
		// switch app configuration
		if backend.nextRing != nil {
			backend.currentRing = backend.nextRing
			backend.nextRing = nil
		}
		request.URL.Scheme = backend.scheme
		request.URL.Host = backend.currentRing.Value.(*url.URL).Host
		glog.Infof("redirect to %s", request.URL.String())
		backend.currentRing = backend.currentRing.Next()
	}
	proxy := &httputil.ReverseProxy{Director: director}
	server := http.Server{
		Addr:    ":80",
		Handler: proxy,
	}
	server.ListenAndServe()

}
