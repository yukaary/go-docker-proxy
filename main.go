package main

import (
	"flag"
	"github.com/golang/glog"
	etcdutil "github.com/yukaary/go-docker-dns/etcd"
	"net/url"
)

type Backend struct {
	serviceName string
	fetchUrl    string
	etcdcli     *etcdutil.EtcdClient
	apps        map[string]url.URL
}

var (
	etcd_endpoint *string
	backend_port  *string
	service_name  *string
	backend       *Backend
)

func NewBackend(fetchUrl, serviceName string) *Backend {
	backend := &Backend{
		fetchUrl:    fetchUrl,
		serviceName: serviceName,
	}
	backend.initialize()
	return backend
}

func (self *Backend) initialize() {
	self.apps = make(map[string]url.URL)
	self.etcdcli = etcdutil.NewEtcdClient([]string{self.fetchUrl})
}

func (self *Backend) fetch() {
	for _, node := range self.etcdcli.GetDir(self.serviceName) {
		glog.Infof("key, value: %s, %s", node.Key, node.Value)
	}
}

func main() {
	glog.Infof("starting proxy.")

	etcd_endpoint := flag.String("etcd-endpoint", "http://127.0.0.1:4001", "etcd url which contains service information.")
	//backend_port := flag.String("backend-port", "80", "backend port to be listened")
	service_name := flag.String("service-name", "", "service name to be blanced.")
	flag.Parse()

	backend = NewBackend(*etcd_endpoint, *service_name)
	backend.fetch()
}
