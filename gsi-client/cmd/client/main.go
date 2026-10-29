package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func loadConfig() {
	exePath, err := os.Executable()
	if err != nil {
		log.Printf("No se pudo determinar la ruta del ejecutable: %v", err)
		return
	}

	exeDir := filepath.Dir(exePath)
	envPath := filepath.Join(exeDir, ".env")

	if err := godotenv.Load(envPath); err != nil {
		log.Printf("No se pudo cargar %s: %v", envPath, err)
		log.Println("Se utilizarán las variables de entorno del sistema.")
		return
	}

	log.Printf("Configuración cargada desde: %s", envPath)
}

func main() {
	loadConfig()

	listenAddr := env("GSI_LISTEN_ADDR", "127.0.0.1:3000")
	apiURL := strings.TrimSpace(os.Getenv("API_URL"))
	apiToken := strings.TrimSpace(os.Getenv("API_TOKEN"))
	timeoutSeconds, err := strconv.Atoi(env("HTTP_TIMEOUT_SECONDS", "3"))
	if err != nil || timeoutSeconds < 1 {
		timeoutSeconds = 3
	}

	if apiURL == "" {
		log.Println("AVISO: API_URL no está configurada; se recibirán GSI pero no se reenviarán.")
	}

	client := &http.Client{
		Timeout: time.Duration(timeoutSeconds) * time.Second,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":"ok"}`)
	})

	mux.HandleFunc("/gsi", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		defer r.Body.Close()

		body, err := io.ReadAll(io.LimitReader(r.Body, 4<<20))
		if err != nil {
			http.Error(w, "could not read request", http.StatusBadRequest)
			return
		}
		if len(body) == 0 {
			http.Error(w, "empty request body", http.StatusBadRequest)
			return
		}
		if !json.Valid(body) {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		log.Printf("GSI recibido: %d bytes", len(body))

		if apiURL != "" {
			req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, apiURL, bytes.NewReader(body))
			if err != nil {
				log.Printf("No se pudo crear petición al API: %v", err)
			} else {
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("User-Agent", "dota-gsi-client/0.1")
				if apiToken != "" {
					req.Header.Set("Authorization", "Bearer "+apiToken)
				}

				resp, err := client.Do(req)
				if err != nil {
					log.Printf("Error reenviando al API (%s): %v", apiURL, err)
				} else {
					_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
					_ = resp.Body.Close()
					if resp.StatusCode < 200 || resp.StatusCode >= 300 {
						log.Printf("El API respondió HTTP %d", resp.StatusCode)
					} else {
						log.Printf("GSI reenviado correctamente (HTTP %d)", resp.StatusCode)
					}
				}
			}
		}

		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "ok")
	})

	server := &http.Server{
		Addr:              listenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Warn if configured to listen on a non-loopback interface.
	host, _, splitErr := net.SplitHostPort(listenAddr)
	if splitErr == nil && host != "127.0.0.1" && host != "localhost" && host != "::1" {
		log.Printf("AVISO: el cliente escucha en %q; verifica que quieras exponerlo en la red local.", listenAddr)
	}

	log.Printf("Cliente Dota GSI iniciado en http://%s", listenAddr)
	log.Printf("Health check: http://%s/health", listenAddr)
	if apiURL != "" {
		log.Printf("API destino: %s", apiURL)
	}
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(fmt.Errorf("servidor HTTP: %w", err))
	}
}
