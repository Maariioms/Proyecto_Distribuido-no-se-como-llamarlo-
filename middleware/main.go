package main

import (
	"encoding/json"
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

// serviceEntry es la forma de cada entrada del documento de service
// discovery (services.json). Los nombres de campo empiezan en mayúscula
// (exportados) porque encoding/json necesita poder verlos desde fuera del
// paquete para llenarlos — un campo no exportado nunca se decodifica.
type serviceEntry struct {
	Prefix    string   `json:"prefix"`
	Healthy   bool     `json:"healthy"`
	Instances []string `json:"instances"`
}

// balancer reparte peticiones entre las instancias de UN servicio por
// round-robin. No Mutex — atomic alcanza para un contador simple.
type balancer struct {
	prefix    string
	instances []*url.URL
	next      atomic.Uint64
}

func newBalancer(prefix string, rawURLs []string) *balancer {
	instances := make([]*url.URL, 0, len(rawURLs))
	for _, raw := range rawURLs {
		u, err := url.Parse(strings.TrimSpace(raw))
		if err != nil {
			log.Fatalf("invalid instance URL %q: %v", raw, err)
		}
		instances = append(instances, u)
	}
	return &balancer{prefix: prefix, instances: instances}
}

func (b *balancer) pick() *url.URL {
	i := b.next.Add(1) - 1
	return b.instances[i%uint64(len(b.instances))]
}

// loadServices lee el documento de service discovery y arma un balancer
// por cada servicio ahí definido. Solo se llama UNA vez, en main().
func loadServices(path string) []*balancer {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("reading service discovery file: %v", err)
	}

	var raw map[string]serviceEntry
	if err := json.Unmarshal(data, &raw); err != nil {
		log.Fatalf("parsing service discovery file: %v", err)
	}

	balancers := make([]*balancer, 0, len(raw))
	for name, entry := range raw {
		if !entry.Healthy {
			log.Printf("skipping unhealthy service %q", name)
			continue
		}
		balancers = append(balancers, newBalancer(entry.Prefix, entry.Instances))
	}
	return balancers
}

// findBalancer decide a qué servicio pertenece una petición mirando el
// prefijo de su path. Devuelve el primero que calce.
func findBalancer(balancers []*balancer, path string) *balancer {
	for _, b := range balancers {
		if strings.HasPrefix(path, b.prefix) {
			return b
		}
	}
	return nil
}

func main() {
	discoveryPath := os.Getenv("SERVICE_DISCOVERY_FILE")
	if discoveryPath == "" {
		discoveryPath = "./services.json"
	}

	balancers := loadServices(discoveryPath)
	log.Printf("loaded %d services from %s", len(balancers), discoveryPath)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		b := findBalancer(balancers, req.URL.Path)
		if b == nil {
			http.Error(w, "no service registered for this path", http.StatusNotFound)
			return
		}

		instance := b.pick()
		log.Printf("dispatching %s %s -> %s", req.Method, req.URL.Path, instance)

		proxy := httputil.NewSingleHostReverseProxy(instance)
		proxy.ServeHTTP(w, req)
	}))

	port := os.Getenv("PORT")
	if port == "" {
		port = "9000"
	}

	log.Printf("middleware listening on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}
