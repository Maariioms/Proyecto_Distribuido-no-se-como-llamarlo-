# Primer log

-   XSIAM-SITO (nombre en proceso)
-   Definitivamente tengo que hacer un commit actualizando la documentacion xd

## Nginx

Esta el cntenedor `nginx-gateway/` que es la imagen `nginx:alpine` + `nginx.conf` en el puerto `8000`.
Va entre el frontend y los backends, y corre en paralelo al middleware en Go que esta en el puerto `9000`:
los dos hacen lo mismo(?) [tengo que entender biene esto aun xd]

Claude explanation: 

```
cliente -> nginx-gateway (:8000) -> accounts / events / status
cliente -> middleware Go (:9000) -> accounts / events / status
```

-   Un bloque `upstream` por servicio -> `accounts_pool` `events_pool` y `status_pool`. Cada uno con `max_fails=3 fail_timeout=30s` que es health check pasivo
-   Un `location` por prefijo -> `/api/accounts` `/api/events` y `/health`
-   `GET /nginx-gateway-heartbeat` responde "nginx gateway alive"
-   Pa balancear solo se agregan mas lineas `server` dentro del `upstream` y por default es round-robin

Probar:

```bash
docker compose up --build
curl localhost:8000/nginx-gateway-heartbeat
curl localhost:8000/api/events
curl localhost:8000/health
```

## gRPC + protobuf

-   Cada backend ahora tiene DOS puertos -> el http de siempre en el `8080` y uno grpc en el `50051`
-   Los dos usan el mismo repositorio y solo cambia como llegan las peticiones
-   El middleware abre su propio puerto grpc `9001` y reenvia a los backends

```
grpcurl / cliente -> middleware (:9001) -> accounts / events / status   LB <-> servicios
accounts --gRPC--> events                                                servicio <-> servicio
frontend / curl   -> middleware (:9000) -> http igual que antes
```

### Como esta armado

-   `proto/<servicio>/v1/<servicio>.proto` -> los contratos `accounts.v1` `events.v1` y `status.v1`. Solo rpc unarios osea una peticion y una respuesta
-   `*/gen/` -> codigo generado con `protoc`. Ya esta commiteado asi que no hace falta instalar nada. Pa regenerarlo es `make proto` o `sh proto.sh` y corre todo dentro de docker
-   `*/internal/grpcserver/` -> el "controller" de grpc de cada backend. Tambien registra el health de grpc que usa consul y reflection que usa `grpcurl` pa no necesitar el `.proto`
-   `backend/internal/events/publisher.go` -> accounts le avisa a events cuando hay login o registro. Es una interfaz `Publisher` asi luego se cambia grpc por kafka sin tocar el controller. Si events esta caido el login sigue funcionando
-   `middleware/grpcproxy.go` -> el proxy grpc. Elige una instancia sana con round-robin y hace la misma llamada al backend
-   Consul -> cada backend se registra dos veces, `<nombre>` pa http y `<nombre>-grpc` con check tipo `GRPC`. El middleware busca los `*-grpc`

| Puerto host | Que es |
|---|---|
| `50051` `50052` `50053` | grpc directo de accounts, events y status |
| `9001` | grpc del middleware que es el load balancer |

### Probar con grpcurl

No hace falta instalarlo. Se corre con docker dentro de la red del compose:

```bash
docker compose up -d --build
alias grpcurl='docker run --rm --network distribuido_default fullstorydev/grpcurl -plaintext'
```

Si ya lo tienes instalado queda como `grpcurl -plaintext localhost:50052 list`.

**1. Reflection, describe y health directo a los servicios**

```
$ grpcurl events:50051 list
events.v1.EventService
grpc.health.v1.Health
grpc.reflection.v1.ServerReflection
grpc.reflection.v1alpha.ServerReflection

$ grpcurl events:50051 describe events.v1.EventService
service EventService {
  rpc CreateEvent ( .events.v1.CreateEventRequest ) returns ( .events.v1.CreateEventResponse );
  rpc GetEvent ( .events.v1.GetEventRequest ) returns ( .events.v1.GetEventResponse );
  rpc ListEvents ( .events.v1.ListEventsRequest ) returns ( .events.v1.ListEventsResponse );
}

$ grpcurl accounts:50051 grpc.health.v1.Health/Check
{ "status": "SERVING" }

$ grpcurl status:50051 status.v1.StatusService/GetStatus
{ "status": "ok", "serviceId": "status-1", "uptime": "36s", "checksDone": "1" }
```

**2. Cuentas y codigos de error**

Equivalencias con http -> `AlreadyExists` es 409, `Unauthenticated` es 401, `InvalidArgument` es 400, `NotFound` es 404 y `Unavailable` es 503

```
$ grpcurl -d '{"username":"grpc.demo","password":"demo1234"}' accounts:50051 accounts.v1.AccountService/Register
{ "account": { "id": "4", "username": "grpc.demo", "createdAt": "2026-10-19T00:59:01.273259516Z" } }

# mismo usuario otra vez
ERROR: Code: AlreadyExists   Message: username already taken

# password de menos de 6 caracteres
ERROR: Code: InvalidArgument   Message: username is required and password must be at least 6 characters

$ grpcurl -d '{"username":"grpc.demo","password":"mala"}' accounts:50051 accounts.v1.AccountService/Login
ERROR: Code: Unauthenticated   Message: invalid username or password
```

**3. Eventos**

```
$ grpcurl -d '{"type":"prueba_grpcurl","severity":"high"}' events:50051 events.v1.EventService/CreateEvent
{ "event": { "id": "2354", "type": "prueba_grpcurl", "severity": "high", "createdAt": "2026-10-19T01:00:41.569472038Z" } }

$ grpcurl -d '{"id":99999}' events:50051 events.v1.EventService/GetEvent
ERROR: Code: NotFound   Message: event not found

$ grpcurl -d '{}' events:50051 events.v1.EventService/CreateEvent
ERROR: Code: InvalidArgument   Message: type and severity are required
```

**4. Comunicacion entre servicios, accounts -> events**

Hacer login en accounts por grpc o por http crea un evento en events

```
$ grpcurl -d '{"username":"grpc.demo","password":"demo1234"}' accounts:50051 accounts.v1.AccountService/Login   # grpc
$ grpcurl -d '{"username":"grpc.demo","password":"mala"}'     accounts:50051 accounts.v1.AccountService/Login   # grpc y falla
$ curl -X POST localhost:9000/api/accounts/login -H "Content-Type: application/json" -d '{"username":"grpc.demo","password":"demo1234"}'   # http

$ grpcurl events:50051 events.v1.EventService/ListEvents   # filtrado a los login_*
{"id":"2359","type":"login_success","severity":"low", ...}   # el del http
{"id":"2358","type":"login_failed","severity":"medium", ...} # el del password malo
{"id":"2357","type":"login_success","severity":"low", ...}   # el del grpc
```

**5. Load balancer, middleware `:9001`**

Las mismas llamadas pero pasando por el LB

```
$ grpcurl middleware:9001 list
accounts.v1.AccountService
events.v1.EventService
grpc.reflection.v1.ServerReflection
grpc.reflection.v1alpha.ServerReflection
status.v1.StatusService

$ grpcurl middleware:9001 status.v1.StatusService/GetStatus
{ "status": "ok", "serviceId": "status-1", "uptime": "3m10s", "checksDone": "2" }

$ grpcurl -d '{"id":99999}' middleware:9001 events.v1.EventService/GetEvent
ERROR: Code: NotFound   Message: event not found      # el error del backend pasa tal cual
```

Logs del middleware con `docker compose logs middleware`:

```
grpc dispatching status.v1.StatusService/GetStatus -> status:50051
grpc dispatching accounts.v1.AccountService/Login -> accounts:50051
grpc dispatching events.v1.EventService/CreateEvent -> events:50051
```

**6. Round-robin del LB con 2 instancias**

Se levanto una segunda instancia de events, `events-2`, temporal. Se registro sola en consul como `events-2-grpc`. Se hicieron 6 llamadas `GetEvent` por el LB y reparte alternando:

```
grpc dispatching events.v1.EventService/GetEvent -> events:50051
grpc dispatching events.v1.EventService/GetEvent -> events-2:50051
grpc dispatching events.v1.EventService/GetEvent -> events:50051
grpc dispatching events.v1.EventService/GetEvent -> events-2:50051
grpc dispatching events.v1.EventService/GetEvent -> events:50051
grpc dispatching events.v1.EventService/GetEvent -> events-2:50051
```

**7. Falla de un servicio con `docker compose stop events`**

```
$ grpcurl -d '{"id":1}' middleware:9001 events.v1.EventService/GetEvent
ERROR: Code: Unavailable   Message: no hay instancias sanas de events-grpc

$ grpcurl -d '{"username":"grpc.demo","password":"demo1234"}' middleware:9001 accounts.v1.AccountService/Login
{ "account": { "id": "4", "username": "grpc.demo", ... } }     # el login sigue funcionando

# en los logs de accounts:
no se pudo avisar a events (login_success): rpc error: code = Unavailable ...
```

### Nota

Pa que `Register` regresara `AlreadyExists` y el http un 409 hubo que activar `TranslateError: true` en `backend/internal/db/db.go`. Sin eso gorm no convertia el error del username repetido y salia un 500.
