package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync/atomic"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// El balancer reparte las peticiones entre varios workers por RR.
// El contador lo incrementan varias goroutines a la vez una por request
// entrante y usamos atomic en vez de Mutex porque Mutex bloquearía a todas
// las goroutines que quieran tocar el contador al mismo tiempo y tendrían
// que esperar su turno solo para hacer el paso
type balancer struct {
	workers []*url.URL
	next    atomic.Uint64
}

func newBalancer(rawURLs []string) *balancer {
	workers := make([]*url.URL, 0, len(rawURLs))
	for _, raw := range rawURLs {
		u, err := url.Parse(strings.TrimSpace(raw))
		if err != nil {
			log.Fatalf("invalid worker URL %q: %v", raw, err)
		}
		workers = append(workers, u)
	}
	return &balancer{workers: workers}
}

// pick devuelve el siguiente worker en la rotación.
func (b *balancer) pick() *url.URL {
	i := b.next.Add(1) - 1
	return b.workers[i%uint64(len(b.workers))]
}

func main() {
	rawWorkers := os.Getenv("WORKERS")
	if rawWorkers == "" {
		rawWorkers = "http://backend-1:8080,http://backend-2:8080"
	}

	b := newBalancer(strings.Split(rawWorkers, ","))

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		worker := b.pick()
		log.Printf("dispatching %s %s -> %s", req.Method, req.URL.Path, worker)

		proxy := httputil.NewSingleHostReverseProxy(worker)
		proxy.ServeHTTP(w, req)
	}))

	port := os.Getenv("PORT")
	if port == "" {
		port = "9000"
	}

	log.Printf("middleware listening on :%s, workers: %v", port, rawWorkers)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}
