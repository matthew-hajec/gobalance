package gobalance

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

type server struct {
	URL   string
	Up    bool
	Proxy *httputil.ReverseProxy
}

type serverPool struct {
	mu      sync.Mutex
	servers []*server
	Offset  int
}

type goBalancerInstance struct {
	Pool *serverPool
}

func CreateGoBalancer() *goBalancerInstance {
	return &goBalancerInstance{
		Pool: &serverPool{
			servers: []*server{},
			Offset:  0,
		},
	}
}

func (g *goBalancerInstance) AddServer(addr string) {
	url, err := url.Parse(addr)

	if err != nil {
		log.Fatal(err)
	}

	g.Pool.servers = append(g.Pool.servers, &server{
		URL:   addr,
		Up:    true,
		Proxy: httputil.NewSingleHostReverseProxy(url),
	})
}

func (g *goBalancerInstance) Start(addr string) {
	http.HandleFunc("/", g.balanceRequest)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func (g *goBalancerInstance) getCurrentProxy() *httputil.ReverseProxy {
	g.Pool.mu.Lock()
	defer g.Pool.mu.Unlock()

	g.Pool.Offset = (g.Pool.Offset + 1) % len(g.Pool.servers)

	return g.Pool.servers[g.Pool.Offset].Proxy
}

func (g *goBalancerInstance) balanceRequest(w http.ResponseWriter, r *http.Request) {
	proxy := g.getCurrentProxy()
	proxy.ServeHTTP(w, r)
}
