package server

import (
	"bytes"
	"embed"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/Valentin-Kaiser/go-dbase-export/pkg/config"
	"github.com/Valentin-Kaiser/go-dbase-export/pkg/job"
)

var RepositoryName string

//go:embed template/index.html
var templates embed.FS

var runningJob *job.Job

type status struct {
	Exported   bool
	Running    bool
	Error      error
	Time       time.Time
	Duration   time.Duration
	Repository []string
}

// IndexHandler handles the index page
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	// Get the index template
	index, err := templates.ReadFile("template/index.html")
	if err != nil {
		fmt.Fprintf(w, "Error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.New("index").Parse(string(index))
	if err != nil {
		fmt.Fprintf(w, "Error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if runningJob == nil {
		tmpl.Execute(w, status{Running: false})
		return
	}

	// Append Repository if an job has run
	var repository []string
	if runningJob != nil && runningJob.IsFinished() {
		files, err := os.ReadDir(config.GetConfig().RepositoryPath)
		if err != nil {
			fmt.Fprintf(w, "Error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, file := range files {
			repository = append(repository, file.Name())
		}
	}

	// Render the template
	tmpl.Execute(w, status{
		Running:    !runningJob.IsFinished(),
		Error:      runningJob.GetError(),
		Exported:   runningJob != nil && runningJob.IsFinished(),
		Repository: repository,
		Time:       runningJob.Time,
		Duration:   runningJob.Elapsed,
	})
}

func ExportHandler(w http.ResponseWriter, r *http.Request) {
	if runningJob != nil && !runningJob.IsFinished() {
		http.Error(w, "A job is already running", http.StatusInternalServerError)
		return
	}

	runningJob = job.New(bytes.NewBuffer(nil), nil)
	go runningJob.Run(
		config.GetConfig().DBPath,
		config.GetConfig().RepositoryPath,
		"json",
	)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	// Get the file name from the url arg
	fileName := r.URL.Query().Get("file")
	if len(fileName) == 0 {
		http.Error(w, "No file name provided", http.StatusBadRequest)
		return
	}

	// Check if the file exists
	if _, err := os.Stat(filepath.Join(config.GetConfig().RepositoryPath, fileName)); os.IsNotExist(err) {
		http.Error(w, "File does not exist", http.StatusNotFound)
		return
	}

	// serve the file
	http.ServeFile(w, r, filepath.Join(config.GetConfig().RepositoryPath, fileName))
}