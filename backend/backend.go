package backend

import (
	"container/ring"
	etcdutil "github.com/yukaary/go-docker-dns/etcd"
	"net/url"
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

func (self *Backend) Fetch() {
	self.apps = make(map[string]*url.URL)
	for _, node := range self.etcdcli.GetDir("services", self.serviceName) {
		self.apps[node.Key] = genUrl(self.scheme, node.Value, self.port)
	}
	self.nextRing = self.NewRing()
}

func (self *Backend) NewRing() *ring.Ring {
	backendRing := ring.New(len(self.apps))

	// when load balance routine (maybe it's a director function?) detects
	// new ring, it should switch it soon.
	for _, v := range self.apps {
		backendRing.Value = v
		backendRing = backendRing.Next()
	}

	return backendRing
}

func (self *Backend) SwitchApps() {
	if self.nextRing != nil {
		self.currentRing = self.nextRing
		self.nextRing = nil
	}
}

func (self *Backend) Scheme() string {
	return self.scheme
}

func (self *Backend) NextHost() string {
	host := self.currentRing.Value.(*url.URL).Host
	self.currentRing = self.currentRing.Next()
	return host
}

func genUrl(scheme, body, port string) *url.URL {
	url, _ := url.Parse(scheme + "://" + body + ":" + port)
	return url
}
