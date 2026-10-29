package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

type savedConfig struct {
	APIURL   string `json:"api_url"`
	APIToken string `json:"api_token"`
}

func configFilePath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil { return "", err }
	return filepath.Join(base, "dota-gsi-client", "config.json"), nil
}

func loadSavedConfig() savedConfig {
	path, err := configFilePath()
	if err != nil { return savedConfig{} }
	data, err := os.ReadFile(path)
	if err != nil { return savedConfig{} }
	var config savedConfig
	if json.Unmarshal(data, &config) != nil { return savedConfig{} }
	return config
}

func saveConfig(config savedConfig) error {
	path, err := configFilePath()
	if err != nil { return err }
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil { return err }
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil { return err }
	return os.WriteFile(path, data, 0600)
}

func prompt(reader *bufio.Reader, label string, required bool) string {
	for {
		fmt.Printf("%s: ", label)
		value, err := reader.ReadString('\n')
		if err != nil && len(value) == 0 {
			log.Fatal("No se pudo leer la configuración desde el terminal.")
		}
		value = strings.TrimSpace(value)
		if value != "" || !required {
			return value
		}
		fmt.Println("Este valor es obligatorio.")
	}
}

func configURL(apiURL string) (string, error) {
	u, err := url.Parse(apiURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", fmt.Errorf("API_URL debe ser una URL HTTP o HTTPS válida")
	}
	path := strings.TrimSuffix(u.Path, "/")
	if !strings.HasSuffix(path, "/gsi") {
		return "", fmt.Errorf("API_URL debe terminar en /gsi para localizar el endpoint de configuración")
	}
	u.Path = strings.TrimSuffix(path, "/gsi") + "/gsi/config"
	return u.String(), nil
}

var libraryPath = regexp.MustCompile(`"path"\s*"([^"]+)"`)

func steamRoots() []string {
	var roots []string
	if runtime.GOOS == "windows" {
		for _, p := range []string{os.Getenv("ProgramFiles(x86)"), os.Getenv("ProgramFiles"), os.Getenv("LOCALAPPDATA")} {
			if p != "" { roots = append(roots, filepath.Join(p, "Steam")) }
		}
	} else if runtime.GOOS == "darwin" {
		home, _ := os.UserHomeDir()
		roots = append(roots, filepath.Join(home, "Library", "Application Support", "Steam"))
	} else {
		home, _ := os.UserHomeDir()
		roots = append(roots, filepath.Join(home, ".steam", "steam"), filepath.Join(home, ".local", "share", "Steam"))
	}
	seen := map[string]bool{}
	for i := 0; i < len(roots); i++ {
		root := roots[i]
		if seen[root] { continue }
		seen[root] = true
		data, err := os.ReadFile(filepath.Join(root, "steamapps", "libraryfolders.vdf"))
		if err != nil { continue }
		for _, match := range libraryPath.FindAllStringSubmatch(string(data), -1) {
			p := filepath.FromSlash(match[1])
			if p != "" && !seen[p] { roots = append(roots, p) }
		}
	}
	return roots
}

func dotaConfigPath() (string, error) {
	for _, root := range steamRoots() {
		game := filepath.Join(root, "steamapps", "common", "dota 2 beta")
		if info, err := os.Stat(game); err == nil && info.IsDir() {
			return filepath.Join(game, "game", "dota", "cfg", "gamestate_integration", "gamestate_integration_dota2.cfg"), nil
		}
	}
	return "", fmt.Errorf("no se encontró Dota 2 en las bibliotecas de Steam detectadas")
}

func installDotaConfig(client *http.Client, endpoint, token string) error {
	path, err := dotaConfigPath()
	if err != nil { return err }
	log.Printf("Descargando configuración GSI desde %s", endpoint)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil { return err }
	if token != "" { req.Header.Set("Authorization", "Bearer "+token) }
	resp, err := client.Do(req)
	if err != nil { return fmt.Errorf("no se pudo descargar la configuración GSI: %w", err) }
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 { return fmt.Errorf("el endpoint de configuración respondió HTTP %d", resp.StatusCode) }
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil { return fmt.Errorf("no se pudo leer la configuración GSI: %w", err) }
	if len(data) == 0 || !strings.Contains(string(data), `"uri"`) || !strings.Contains(string(data), "/gsi") {
		return fmt.Errorf("el endpoint no devolvió un archivo de configuración GSI válido")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil { return fmt.Errorf("no se pudo crear la carpeta de configuración: %w", err) }
	if err := os.WriteFile(path, data, 0644); err != nil { return fmt.Errorf("no se pudo instalar el archivo GSI: %w", err) }
	log.Printf("Configuración de Dota 2 instalada: %s", path)
	return nil
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("=== Cliente Dota 2 Game State Integration ===")
	previous := loadSavedConfig()
	apiURL := ""
	if previous.APIURL != "" {
		fmt.Printf("API_URL [%s] (Enter reutiliza): ", previous.APIURL)
		value, err := reader.ReadString('\n')
		if err != nil && len(value) == 0 { log.Fatal("No se pudo leer la configuración desde el terminal.") }
		apiURL = strings.TrimSpace(value)
		if apiURL == "" { apiURL = previous.APIURL }
	} else {
		apiURL = prompt(reader, "API_URL (endpoint GSI del backend)", true)
	}
	apiToken := ""
	if previous.APIToken != "" {
		fmt.Print("API_TOKEN [guardado] (Enter reutiliza; escribe otro para cambiar): ")
		value, err := reader.ReadString('\n')
		if err != nil && len(value) == 0 { log.Fatal("No se pudo leer la configuración desde el terminal.") }
		apiToken = strings.TrimSpace(value)
		if apiToken == "" { apiToken = previous.APIToken }
	} else {
		apiToken = prompt(reader, "API_TOKEN (puede quedar vacío si el backend no lo requiere)", false)
	}
	configEndpoint, err := configURL(apiURL)
	if err != nil { log.Fatal(err) }

	listenAddr := "127.0.0.1:3000"
	client := &http.Client{Timeout: 10 * time.Second}
	if err := installDotaConfig(client, configEndpoint, apiToken); err != nil {
		log.Fatalf("No se pudo configurar Dota 2: %v", err)
	}
	if err := saveConfig(savedConfig{APIURL: apiURL, APIToken: apiToken}); err != nil {
		log.Printf("AVISO: no se pudo guardar la configuración local: %v", err)
	} else {
		log.Println("Configuración guardada para la próxima ejecución.")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":"ok"}`)
	})
	mux.HandleFunc("/gsi", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
		defer r.Body.Close()
		body, err := io.ReadAll(io.LimitReader(r.Body, 4<<20))
		if err != nil { http.Error(w, "could not read request", http.StatusBadRequest); return }
		if len(body) == 0 { http.Error(w, "empty request body", http.StatusBadRequest); return }
		if !json.Valid(body) { http.Error(w, "invalid JSON", http.StatusBadRequest); return }
		log.Printf("GSI recibido: %d bytes", len(body))
		req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, apiURL, bytes.NewReader(body))
		if err != nil { log.Printf("No se pudo crear petición al API: %v", err) } else {
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("User-Agent", "dota-gsi-client/0.1")
			if apiToken != "" { req.Header.Set("Authorization", "Bearer "+apiToken) }
			resp, err := client.Do(req)
			if err != nil { log.Printf("Error reenviando al API (%s): %v", apiURL, err) } else {
				_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10)); _ = resp.Body.Close()
				if resp.StatusCode < 200 || resp.StatusCode >= 300 { log.Printf("El API respondió HTTP %d", resp.StatusCode) } else { log.Printf("GSI reenviado correctamente (HTTP %d)", resp.StatusCode) }
			}
		}
		w.WriteHeader(http.StatusOK); _, _ = io.WriteString(w, "ok")
	})

	server := &http.Server{Addr: listenAddr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	host, _, splitErr := net.SplitHostPort(listenAddr)
	if splitErr == nil && host != "127.0.0.1" && host != "localhost" && host != "::1" { log.Printf("AVISO: el cliente escucha en %q; verifica que quieras exponerlo en la red local.", listenAddr) }
	log.Printf("Cliente Dota GSI iniciado en http://%s", listenAddr)
	log.Printf("Health check: http://%s/health", listenAddr)
	log.Printf("API destino: %s", apiURL)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed { log.Fatal(fmt.Errorf("servidor HTTP: %w", err)) }
}
