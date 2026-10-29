# Dota Live Tactical Assistant

Dashboard táctico responsive para Smart TV y smartphone. Incluye frontend web, API Node.js, recepción GSI, autenticación por token de instalación, WebSocket, recomendaciones heurísticas y estadísticas. **MariaDB es remota y no se ejecuta en Docker Compose.**

## Despliegue

1. En el servidor MariaDB remoto, crea la base y usuario:
   ```sql
   CREATE DATABASE dota_live CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
   CREATE USER 'dota_api'@'IP_O_HOST_API' IDENTIFIED BY 'CAMBIA_LA_CLAVE';
   GRANT SELECT, INSERT, UPDATE, DELETE, CREATE, ALTER, INDEX ON dota_live.* TO 'dota_api'@'IP_O_HOST_API';
   FLUSH PRIVILEGES;
   ```
2. Ejecuta `backend/sql/001_schema.sql` contra esa base:
   `mariadb -h DB_HOST -P 3306 -u dota_api -p dota_live < backend/sql/001_schema.sql`
3. Copia `.env.example` a `.env`, configura los datos de MariaDB remota y cambia los secretos.
4. Ejecuta `docker compose up -d --build`.
5. Abre `http://IP_DEL_NAS:8080`. La API está en `http://IP_DEL_NAS:3210/health`.

## Registrar cliente GSI

```bash
curl -X POST http://IP_DEL_NAS:3210/api/v1/clients \
 -H 'Content-Type: application/json' -H 'X-Admin-Key: TU_API_ADMIN_KEY' \
 -d '{"clientName":"PC Sala","platform":"windows","label":"Jugador 1"}'
```

Guarda el `token` que devuelve; solo se muestra una vez. Configura el cliente GSI para enviar `POST http://IP_DEL_NAS:3210/api/v1/gsi` con `Authorization: Bearer TOKEN_INDIVIDUAL`. Si tu cliente actual tiene variables `API_URL` y `API_TOKEN`, usa esos valores.

## Seguridad

- `API_ADMIN_KEY` permite registrar instalaciones; `DASHBOARD_KEY` autoriza el dashboard. Deben ser secretos diferentes, largos y aleatorios.
- La API espera acceso desde una red local confiable. Para acceso fuera de casa configura HTTPS/WSS y un proxy inverso.
- Restringe el puerto 3306 de MariaDB al host de la API; no lo expongas públicamente.
- GSI cambia según el estado del juego y no entrega necesariamente todos los datos de mapa/enemigos. Los consejos del motor inicial son heurísticos y no sustituyen un análisis completo.
- Iconos desde CDN de Steam requieren Internet en la TV/móvil.

## Endpoints

- `GET /health`
- `POST /api/v1/clients` (X-Admin-Key)
- `POST /api/v1/gsi` (Bearer token del cliente)
- `GET /api/v1/dashboard` (X-Dashboard-Key)
- `GET /api/v1/players/:id` (X-Dashboard-Key)
- `GET /api/v1/stats/progress?days=30` (X-Dashboard-Key)
- `WS /ws` (primer mensaje JSON: `{"type":"auth","key":"DASHBOARD_KEY"}`)
