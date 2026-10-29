# Dota 2 GSI Client (Go) — compilación con Docker Compose

Cliente inicial multiplataforma para recibir eventos **Game State Integration (GSI)** de Dota 2 en la computadora donde se ejecuta el juego y reenviarlos a un API Node.js en tu NAS Synology.

## Arquitectura

`Dota 2 → cliente local Go (127.0.0.1:3000) → API Node.js del NAS → IA / web para TV`

El cliente escucha solo en `127.0.0.1` por defecto: Dota 2 puede conectarse localmente y otros equipos de la red no pueden enviarle peticiones directamente.

## Requisitos

- Docker Desktop instalado y en ejecución (Windows o macOS).
- Dota 2 instalado en la computadora donde se ejecutará el cliente.
- Un API accesible desde esa computadora en el NAS. El API debe aceptar un `POST` con el JSON GSI.

## 1. Compilar dentro de Docker Compose

Descomprime el ZIP y abre una terminal en la carpeta `dota-gsi-client`.

```bash
docker compose run --rm builder
```

Los ejecutables se generan en `output/`:
- `dota-gsi-windows-amd64.exe`
- `dota-gsi-darwin-amd64` (Mac Intel)
- `dota-gsi-darwin-arm64` (Mac Apple Silicon)

Esto genera **ejecutables**, no instaladores firmados. En macOS, Gatekeeper puede mostrar advertencias para binarios propios; Windows también puede mostrar advertencias de SmartScreen.

## 2. Configurar el cliente

Crea una copia de `.env.example` llamada `.env` y configura:

```dotenv
GSI_LISTEN_ADDR=127.0.0.1:3000
API_URL=http://192.168.68.XX:3200/api/dota/gsi
API_TOKEN=cambia-este-token
HTTP_TIMEOUT_SECONDS=3
```

Cambia `API_URL` por la dirección y ruta reales de tu API Node.js en el NAS. `API_TOKEN` se envía como `Authorization: Bearer ...`. Si el API aún no valida ese encabezado, puedes dejar `API_TOKEN` vacío durante una prueba local.

> Usa la IP real de tu NAS y el puerto publicado por Docker. No expongas el puerto del cliente GSI a Internet.

## 3. Ejecutar el cliente

Windows PowerShell:

```powershell
$env:GSI_LISTEN_ADDR="127.0.0.1:3000"
$env:API_URL="http://192.168.68.XX:3200/api/dota/gsi"
$env:API_TOKEN="cambia-este-token"
.\output\dota-gsi-windows-amd64.exe
```

macOS (Apple Silicon):

```bash
export GSI_LISTEN_ADDR=127.0.0.1:3000
export API_URL=http://192.168.68.XX:3200/api/dota/gsi
export API_TOKEN=cambia-este-token
chmod +x output/dota-gsi-darwin-arm64
./output/dota-gsi-darwin-arm64
```

El cliente muestra en consola el estado de las peticiones y los errores de envío. Mantén la aplicación ejecutándose mientras juegas.

## 4. Configurar Dota 2

Copia `gamestate_integration_dota2.cfg` a la carpeta `game/dota/cfg/gamestate_integration` de la instalación de Dota 2.

Rutas habituales:
- Windows: `...\Steam\steamapps\common\dota 2 beta\game\dota\cfg\gamestate_integration\`
- macOS: `~/Library/Application Support/Steam/steamapps/common/dota 2 beta/game/dota/cfg/gamestate_integration/` (la ruta exacta puede variar según dónde esté instalada la biblioteca de Steam).

Si la carpeta `gamestate_integration` no existe, créala. El nombre del archivo debe empezar por `gamestate_integration_` y terminar en `.cfg`.

El archivo de ejemplo envía GSI a `http://127.0.0.1:3000/gsi`. Reinicia Dota 2 después de añadirlo.

## 5. Probar sin Dota 2

Con el cliente ejecutándose, puedes enviar un JSON de prueba desde otra terminal de la misma computadora:

```bash
curl -X POST http://127.0.0.1:3000/gsi \
  -H "Content-Type: application/json" \
  -d '{"map":{"name":"test"},"player":{"activity":"playing"}}'
```

El cliente intentará reenviar el JSON a `API_URL`. Si no has iniciado el API del NAS, verás un error de envío; eso es normal para esta prueba.

## Contrato esperado del API Node.js

El API del NAS debería aceptar:
- Método: `POST`
- Ruta: la configurada en `API_URL`
- `Content-Type: application/json`
- Opcional: `Authorization: Bearer <API_TOKEN>`
- Cuerpo: JSON original de GSI

El cliente devuelve `200 OK` a Dota cuando el JSON se recibió correctamente, aunque el API del NAS esté caído. Los fallos de reenvío se registran en la consola; así, un problema temporal del NAS no bloquea la recepción local de GSI.

## Limitaciones de esta versión inicial

- Es un cliente de consola, no una aplicación con interfaz gráfica ni un servicio de inicio automático.
- No instala automáticamente el archivo `.cfg`; debes copiarlo manualmente.
- No incluye todavía el API Node.js del NAS ni la lógica de IA.
- El endpoint del NAS debe existir y ser accesible desde la PC.
- Para jugar desde dos computadoras a la vez, ejecuta un cliente en cada una y añade una identidad de jugador en una versión posterior del protocolo/API.
