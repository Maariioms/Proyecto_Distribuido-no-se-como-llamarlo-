package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// routeEntry es una fila de routes.json: "los paths que empiezan con Prefix
// le tocan al servicio Service de Consul". Las instancias ya NO viven aqui,
// se las pregunta el middleware a Consul (service discovery dinamico).
type routeEntry struct {
	Prefix  string `json:"prefix"`
	Service string `json:"service"`
}

// balancer reparte peticiones entre las instancias SANAS de UN servicio.
// El contador es un solo valor -> atomic. La lista de instancias la leen
// muchas peticiones a la vez y solo se reemplaza cada 5s -> RWMutex.
type balancer struct {
	prefix  string
	service string

	mu        sync.RWMutex
	instances []*url.URL
	next      atomic.Uint64
}

func (b *balancer) setInstances(instances []*url.URL) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.instances) != len(instances) {
		log.Printf("consul: %s ahora tiene %d instancia(s) sana(s)", b.service, len(instances))
	}
	b.instances = instances
}

// pick devuelve la siguiente instancia (round-robin) o nil si no hay ninguna sana.
func (b *balancer) pick() *url.URL {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if len(b.instances) == 0 {
		return nil
	}
	i := b.next.Add(1) - 1
	return b.instances[i%uint64(len(b.instances))]
}

// consulEntry es la parte de la respuesta de Consul que nos interesa.
type consulEntry struct {
	Service struct {
		Address string `json:"Address"`
		Port    int    `json:"Port"`
	} `json:"Service"`
}

// queryConsul pregunta "dame las instancias de este servicio que estan pasando
// su health check". El ?passing=true es lo que filtra a las caidas.
func queryConsul(consulAddr, service string) ([]*url.URL, error) {
	reqURL := fmt.Sprintf("%s/v1/health/service/%s?passing=true", consulAddr, service)
	client := http.Client{Timeout: 3 * time.Second}

	resp, err := client.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("consul respondio %d", resp.StatusCode)
	}

	var entries []consulEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, err
	}

	instances := make([]*url.URL, 0, len(entries))
	for _, e := range entries {
		u, err := url.Parse(fmt.Sprintf("http://%s:%d", e.Service.Address, e.Service.Port))
		if err != nil {
			continue
		}
		instances = append(instances, u)
	}
	return instances, nil
}

// refreshAll actualiza la lista de instancias de cada servicio. Si Consul no
// contesta, se queda con la lista anterior en vez de dejar todo vacio.
func refreshAll(balancers []*balancer, consulAddr string) {
	for _, b := range balancers {
		instances, err := queryConsul(consulAddr, b.service)
		if err != nil {
			log.Printf("consul: no se pudo consultar %q, sigo con la lista anterior: %v", b.service, err)
			continue
		}
		b.setInstances(instances)
	}
}

// startDiscoveryLoop refresca una vez ya y luego cada `interval`, en una
// goroutine aparte (mismo patron del ticker del generador de events).
func startDiscoveryLoop(balancers []*balancer, consulAddr string, interval time.Duration) {
	refreshAll(balancers, consulAddr)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			refreshAll(balancers, consulAddr)
		}
	}()
}

// loadRoutes lee routes.json y arma un balancer (todavia sin instancias) por ruta.
// Solo se llama UNA vez, en main().
func loadRoutes(path string) []*balancer {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("reading routes file: %v", err)
	}

	var routes []routeEntry
	if err := json.Unmarshal(data, &routes); err != nil {
		log.Fatalf("parsing routes file: %v", err)
	}

	balancers := make([]*balancer, 0, len(routes))
	for _, route := range routes {
		balancers = append(balancers, &balancer{prefix: route.Prefix, service: route.Service})
	}
	return balancers
}

// findBalancer decide a que servicio pertenece una peticion mirando el
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
	routesPath := os.Getenv("ROUTES_FILE")
	if routesPath == "" {
		routesPath = "./routes.json"
	}

	consulAddr := os.Getenv("CONSUL_HTTP_ADDR")
	if consulAddr == "" {
		consulAddr = "http://consul:8500"
	}

	balancers := loadRoutes(routesPath)
	startDiscoveryLoop(balancers, consulAddr, 5*time.Second)
	log.Printf("loaded %d routes from %s, discovery via %s", len(balancers), routesPath, consulAddr)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		b := findBalancer(balancers, req.URL.Path)
		if b == nil {
			http.Error(w, "no service registered for this path", http.StatusNotFound)
			return
		}

		instance := b.pick()
		if instance == nil {
			http.Error(w, "no healthy instance available for "+b.service, http.StatusServiceUnavailable)
			return
		}
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
