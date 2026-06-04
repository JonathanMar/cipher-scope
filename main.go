package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os/exec"
	"runtime"

	goflag "flag"

	"auditor/server"
)

//go:embed ui
var uiFiles embed.FS

func main() {
	addr := goflag.String("addr", ":8080", "Endereço do servidor web (ex: :8080 ou 0.0.0.0:8080)")
	noBrowser := goflag.Bool("no-browser", false, "Não abrir o browser automaticamente")
	goflag.Parse()

	stripped, err := fs.Sub(uiFiles, "ui")
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	// Static UI
	mux.Handle("/", http.FileServer(http.FS(stripped)))

	// API
	mux.HandleFunc("/api/info", server.HandleInfo)
	mux.HandleFunc("/api/hash", server.HandleHashExtended)
	mux.HandleFunc("/api/crack", server.HandleCrack)
	mux.HandleFunc("/api/stop", server.HandleStop)
	mux.HandleFunc("/api/progress", server.HandleSSE)
	mux.HandleFunc("/api/encode", server.HandleEncode)
	mux.HandleFunc("/api/identify", server.HandleIdentifyHash)
	mux.HandleFunc("/api/chrome", server.HandleChromeDecrypt)

	url := buildURL(*addr)
	fmt.Printf("╔══════════════════════════════════════╗\n")
	fmt.Printf("║       Cipher Scope  Dashboard        ║\n")
	fmt.Printf("╠══════════════════════════════════════╣\n")
	fmt.Printf("║  URL: %-31s║\n", url)
	fmt.Printf("╚══════════════════════════════════════╝\n")

	if !*noBrowser {
		go openBrowser(url)
	}

	log.Fatal(http.ListenAndServe(*addr, mux))
}

// buildURL monta a URL de acesso correta a partir do addr configurado.
//
//	":8080"              → "http://localhost:8080"
//	"0.0.0.0:8080"      → "http://localhost:8080"
//	"192.168.0.5:8080"  → "http://192.168.0.5:8080"
func buildURL(addr string) string {
	host := addr
	port := ""
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			host = addr[:i]
			port = addr[i:] // inclui ":"
			break
		}
	}
	if host == "" || host == "0.0.0.0" {
		host = "localhost"
	}
	return fmt.Sprintf("http://%s%s", host, port)
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return
	}
	cmd.Start()
}
