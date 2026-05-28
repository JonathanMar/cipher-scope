package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os/exec"
	"runtime"

	"auditor/server"

	goflag "flag"
	"fmt"
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
	mux.HandleFunc("/api/hash", server.HandleHash)
	mux.HandleFunc("/api/crack", server.HandleCrack)
	mux.HandleFunc("/api/stop", server.HandleStop)
	mux.HandleFunc("/api/progress", server.HandleSSE)

	url := fmt.Sprintf("http://localhost%s", *addr)
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
