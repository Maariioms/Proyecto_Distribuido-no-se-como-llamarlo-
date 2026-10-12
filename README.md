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

- Un bloque `upstream` por servicio (`accounts_pool`, `events_pool`, `status_pool`) con
  `max_fails=3 fail_timeout=30s` (health check pasivo).
- Un `location` por prefijo: `/api/accounts`, `/api/events`, `/health`.
- `GET /nginx-gateway-heartbeat` responde "nginx gateway alive".
- Para balancear, se agregan más líneas `server` dentro del `upstream` (round-robin por default).

Probar:

```bash
docker compose up --build
curl localhost:8000/nginx-gateway-heartbeat
curl localhost:8000/api/events
curl localhost:8000/health
```
