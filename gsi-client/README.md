# Cliente Dota 2 GSI (Go)

Cliente multiplataforma de consola para recibir Game State Integration (GSI) de Dota 2 y reenviarlo al backend Dota Live. Al abrirlo pide `API_URL` y `API_TOKEN`, descarga la configuración GSI autenticándose contra el backend, la instala en la carpeta detectada de Dota 2 y muestra en el terminal el estado de recepción y reenvío.

## Compilar

Con Docker Desktop:

```bash
docker compose run --rm builder
```

Los ejecutables se generan en `output/`: `Dota GSI.exe` para Windows x64 y `Dota GSI-amd64.app` / `Dota GSI-arm64.app` para macOS Intel y Apple Silicon. El `.exe` incorpora el icono de Windows. En Mac, abre el paquete `.app`; este abre Terminal para conservar los prompts y el feedback del cliente. También se dejan los binarios macOS sueltos dentro de `output/`.

La compilación local con `go build` solo produce el binario y no incluye estos iconos ni el paquete `.app`; usa Docker Compose para generar los artefactos listos con icono.

Para compilar solo el binario en una máquina con Go:

```bash
go build -o output/dota-gsi-client ./cmd/client
```

## Ejecutar y configurar

Abre el ejecutable desde un terminal. Ingresa:

- `API_URL`: URL completa del endpoint GSI, por ejemplo `http://IP_SRV:3210/api/v1/gsi`.
- `API_TOKEN`: token individual creado en el backend. Puede dejarse vacío solo si el backend no requiere autenticación.

El cliente obtiene el archivo desde `GET /api/v1/gsi/config` usando ese mismo token. Por ello, `API_URL` debe terminar en `/gsi`. No necesita archivo `.env` ni variables de entorno. Después de descargar e instalar la configuración con éxito, guarda la URL y el token localmente para sugerirlos la próxima vez; basta con pulsar Enter para reutilizar cada valor. El archivo queda en la carpeta de configuración del usuario (`dota-gsi-client/config.json`) con permisos privados cuando el sistema los admite. El token se guarda en texto claro dentro de ese archivo, así que mantenlo protegido. Los valores locales por defecto son `127.0.0.1:3000` y timeout HTTP de 10 segundos.

El ejecutable busca Dota 2 en las bibliotecas Steam habituales, crea `game/dota/cfg/gamestate_integration` si hace falta e instala `gamestate_integration_dota2.cfg`. Si no encuentra Dota 2, revisa que esté instalado y que Steam tenga actualizado `steamapps/libraryfolders.vdf`.

Ejemplo de URL para Dota Live:

```text
http://IP_SRV:3210/api/v1/gsi
```

Después de instalar el archivo, inicia o reinicia Dota 2. Mantén el cliente abierto durante la partida. El listener local solo acepta conexiones en `127.0.0.1:3000`.

## Contrato del backend

El backend debe exponer:

- `GET /api/v1/gsi/config` con `Authorization: Bearer <API_TOKEN>` para entregar el archivo Valve `.cfg` como texto plano.
- `POST /api/v1/gsi` con el mismo encabezado para recibir el JSON GSI.

El cliente devuelve `200 OK` a Dota al recibir un JSON válido, aunque el reenvío al backend falle; el error se muestra en la consola.

## Uso de Red

Estimaciones referenciales del tráfico que el cliente reenvía al backend. Se calculan con la frecuencia y el tamaño promedio indicados; el uso real cambia según las actualizaciones GSI que genere la partida. La configuración usa `throttle = 0.1 s` como referencia para un máximo aproximado de 10 envíos por segundo, aunque Dota puede comportarse de forma distinta.

| Escenario | Supuesto | Tráfico aproximado |
|---|---:|---:|
| Menú o poca actividad | 2 envíos/s × 2 KB | 14 MB por hora |
| Partida, referencia intermedia | 5 envíos/s × 5 KB | 90 MB por hora |
| Partida, escenario alto | Hasta 10 envíos/s × 10 KB | 360 MB por hora |

Fórmula de referencia: `envíos por segundo × tamaño promedio del paquete × tiempo`. Si el backend está en la intranet, este tráfico circula por la red local y no consume datos de Internet. Solo saldrá a Internet si se accede al backend mediante una dirección pública, una VPN o una ruta externa. La descarga del archivo de configuración al iniciar el cliente es pequeña.

## Recursos locales

Estimaciones referenciales del cliente en el computador donde se ejecuta Dota 2; no incluyen CPU ni memoria del backend. El cliente no renderiza gráficos ni utiliza la GPU de forma significativa.

| Recurso local | Uso aproximado |
|---|---:|
| CPU | Cerca de 0–1 % en promedio en un equipo moderno |
| Memoria | Aproximadamente 10–30 MB |
| GPU | Prácticamente 0 % |
| Disco durante el juego | Mínimo; los mensajes se muestran en el terminal y no se guardan continuamente en un archivo |

Estas cifras son estimaciones, no resultados de un benchmark en cada computador. No se espera una diferencia apreciable en los FPS de Dota 2.
