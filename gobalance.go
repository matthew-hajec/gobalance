package gobalance

import (
	"log"
	"net/http"
	"net/http/httputil"
	"sync"
)

type Server struct {
	URL   string
	Up    bool
	Proxy *httputil.ReverseProxy
}

type ServerPool struct {
	mu      sync.Mutex
	Servers []*Server
	Offset  int
}

type GoBalancerInstance struct {
	Pool *ServerPool
}

func CreateGoBalancer() *GoBalancerInstance {
	return &GoBalancerInstance{
		Pool: &ServerPool{
			Servers: []*Server{},
			Offset:  0,
		},
	}
}

func (g *GoBalancerInstance) getCurrentProxy() *httputil.ReverseProxy {
	g.Pool.mu.Lock()
	defer g.Pool.mu.Unlock()

	g.Pool.Offset = (g.Pool.Offset + 1) % len(g.Pool.Servers)

	return g.Pool.Servers[g.Pool.Offset].Proxy
}

func (g *GoBalancerInstance) balanceRequest(w http.ResponseWriter, r *http.Request) {
	proxy := g.getCurrentProxy()
	proxy.ServeHTTP(w, r)
}

func (g *GoBalancerInstance) Start(addr string) {
	http.HandleFunc("/", g.balanceRequest)
	log.Fatal(http.ListenAndServe(addr, nil))
}
