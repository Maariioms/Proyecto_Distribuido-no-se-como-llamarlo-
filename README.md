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

## Kafka

-   Cola de mensajeria asincrona entre los backends con `apache/kafka:3.8.0` en modo KRaft osea sin zookeeper, un solo nodo que es broker y controller
-   Mismo esquema del `kafka_example` del profe -> `segmentio/kafka-go` y un helper `kafkax` con `EnsureTopic` `NewWriter` `Consume` y `processWithRetry`
-   Antes accounts le avisaba a events por grpc y si events estaba caido el aviso se perdia. Ahora accounts deja el mensaje en la cola y events lo lee cuando puede

```
accounts --> [accounts.activity] --> events --> [events.created] --> status
                     |
                     +--> [accounts.activity.dlq]   los que fallan 3 veces
```

### Topicos y quien los usa

| Topico | Productor | Consumidor, group | Pa que |
|---|---|---|---|
| `accounts.activity` | accounts | events, `events-writer` | `account_created` `login_success` `login_failed` y events los guarda en postgres |
| `events.created` | events | status, `status-counter` | cada evento guardado avisa y status los cuenta |
| `accounts.activity.dlq` | events | nadie | mensajes que fallaron todos los intentos |

-   3 particiones por topico y replication 1 pq es un solo broker
-   Mensaje json `{type, severity, at}` y la key es el `type` pa que el mismo tipo caiga en la misma particion y se mantenga el orden
-   Cada group tiene su propia copia de los mensajes. Si despues otro servicio se une con otro group tambien recibe todo

### Como esta armado

-   `*/internal/kafkax/` -> el helper, igual en backend events y status como `consul.go`. `topics.go` tiene los nombres de topicos y groups
-   `backend/internal/events/kafka_publisher.go` -> implementa la interfaz `Publisher` que ya existia. Por eso el controller de accounts no cambio. Si hay `KAFKA_BROKERS` usa kafka, si no cae a grpc y si no hay nada usa `Noop`
-   `events/internal/consumer/` -> lee `accounts.activity` y llama `repo.Create`
-   `events/internal/repository/publishing_repository.go` -> envuelve al repo y despues de guardar publica en `events.created`. Asi el generador el http y el grpc tambien avisan sin saber de kafka
-   `status/internal/consumer/` -> lee `events.created` y sube contadores en el Tracker. Se ven en `events_seen` y `events_by_severity` del `/health`
-   El offset se confirma hasta que el handler termina bien. Es at-least-once y si el servicio se cae a la mitad el mensaje puede llegar dos veces. Un id unico por evento lo arreglaria pero cambia el esquema y no se hizo
-   Docker -> el servicio `kafka` tiene dos puertas. `kafka:9092` pa los contenedores y `localhost:29092` pa tu compu

### Probar

```bash
docker compose up -d --build
```

**1. Topicos creados**

```
$ docker exec kafka /opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --list
__consumer_offsets
accounts.activity
accounts.activity.dlq
events.created

$ docker exec kafka /opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --describe --topic accounts.activity
Topic: accounts.activity   PartitionCount: 3   ReplicationFactor: 1
	Partition: 0   Leader: 1   Replicas: 1   Isr: 1
	Partition: 1   Leader: 1   Replicas: 1   Isr: 1
	Partition: 2   Leader: 1   Replicas: 1   Isr: 1
```

**2. Registro y logins por el middleware, y mensajes crudos en la cola**

```
$ curl -X POST localhost:9000/api/accounts/register -H "Content-Type: application/json" -d '{"username":"kafka.demo","password":"demo1234"}'
$ curl -X POST localhost:9000/api/accounts/login ... "demo1234"    # bien
$ curl -X POST localhost:9000/api/accounts/login ... "mala"        # falla

$ docker exec kafka /opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic accounts.activity --from-beginning --property print.key=true
login_success    {"type":"login_success","severity":"low","at":"2026-10-24T00:32:56.116023281Z"}
account_created  {"type":"account_created","severity":"low","at":"2026-10-24T00:32:55.882862288Z"}
login_failed     {"type":"login_failed","severity":"medium","at":"2026-10-24T00:32:56.326313271Z"}
```

Logs de events con `docker compose logs events`:

```
kafka: accounts.activity partition=2 offset=0 -> evento account_created guardado
kafka: accounts.activity partition=0 offset=0 -> evento login_success guardado
kafka: accounts.activity partition=1 offset=0 -> evento login_failed guardado
```

Y `GET /api/events` ya los muestra como `account_created` `login_success` y `login_failed`. Notar que cada tipo cayo en una particion distinta por la key.

**3. Status cuenta lo que pasa por `events.created`**

```
$ curl localhost:9000/health
{"checks_done":1,"events_by_severity":{"low":2,"medium":2},"events_seen":4,"service_id":"status-1","status":"ok","uptime":"29s"}
```

Sube solo con el generador de events cada 15s.

**4. Lo asincrono, events caido**

Esto es lo que grpc no podia. Se para events, se hacen 3 logins y siguen respondiendo 200. Los mensajes esperan en la cola, el `LAG` del group es lo que falta por leer:

```
$ docker compose stop events
$ curl ... /login    # x3
login 200
login 200
login 200

$ docker exec kafka /opt/kafka/bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 --describe --group events-writer
GROUP          TOPIC              PARTITION  CURRENT-OFFSET  LOG-END-OFFSET  LAG
events-writer  accounts.activity  0          1               4               3
events-writer  accounts.activity  1          1               1               0
events-writer  accounts.activity  2          1               1               0

$ docker compose start events
kafka: accounts.activity partition=0 offset=1 -> evento login_success guardado
kafka: accounts.activity partition=0 offset=2 -> evento login_success guardado
kafka: accounts.activity partition=0 offset=3 -> evento login_success guardado
```

Los 3 eventos aparecieron al volver y no se perdio ninguno.

### Nota

-   La DLQ se crea pero no se probo forzando un fallo. Solo se usa si un mensaje falla 3 intentos seguidos
-   El grpc de accounts a events sigue en el codigo como fallback. En el compose esta comentado `EVENTS_GRPC_ADDR` pa que de verdad use kafka
