package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Valentin-Kaiser/go-dbase-export/pkg/config"
)

// Start the server
func Start() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", IndexHandler)
	mux.HandleFunc("/export/", ExportHandler)
	mux.HandleFunc("/download/", DownloadHandler)

	server := http.Server{
		Addr:              fmt.Sprintf(":%v", config.GetConfig().Port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Start the server
	log.Fatal(server.ListenAndServe())
}