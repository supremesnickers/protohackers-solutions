package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

var (
	validChallenges = map[string]bool{
		"smoke-test":      true,
		"primetime":       true,
		"pest-control":    true,
		"means-to-an-end": true,
	}

	mu             sync.Mutex
	currentProcess *exec.Cmd
	currentName    string
)

func stopCurrentServer() error {
	if currentProcess != nil && currentProcess.Process != nil {
		log.Printf("Stopping current server: %s (PID: %d)", currentName, currentProcess.Process.Pid)
		if err := currentProcess.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}
		currentProcess.Wait()
		currentProcess = nil
		currentName = ""
		log.Println("Server stopped")
	}
	return nil
}

func startServer(name string) error {
	// Get the project root directory
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	projectRoot := filepath.Dir(filepath.Dir(execPath))

	// If running via go run, use working directory instead
	if wd, err := os.Getwd(); err == nil {
		projectRoot = wd
	}

	challengeDir := filepath.Join(projectRoot, name)

	// Build the challenge
	log.Printf("Building %s...", name)
	buildCmd := exec.Command("go", "build", "-o", "server", ".")
	buildCmd.Dir = challengeDir
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("failed to build %s: %w", name, err)
	}

	// Start the server
	log.Printf("Starting %s on port 8080...", name)
	serverPath := filepath.Join(challengeDir, "server")
	currentProcess = exec.Command(serverPath)
	currentProcess.Dir = challengeDir
	currentProcess.Stdout = os.Stdout
	currentProcess.Stderr = os.Stderr
	if err := currentProcess.Start(); err != nil {
		return fmt.Errorf("failed to start %s: %w", name, err)
	}

	currentName = name
	log.Printf("Server %s started (PID: %d)", name, currentProcess.Process.Pid)
	return nil
}

func setChallengeHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "missing 'name' parameter", http.StatusBadRequest)
		return
	}

	if !validChallenges[name] {
		http.Error(w, fmt.Sprintf("invalid challenge name: %s. Valid options: smoke-test, primetime, pest-control", name), http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	// Stop current server if running
	if err := stopCurrentServer(); err != nil {
		http.Error(w, fmt.Sprintf("failed to stop current server: %v", err), http.StatusInternalServerError)
		return
	}

	// Start the new server
	if err := startServer(name); err != nil {
		http.Error(w, fmt.Sprintf("failed to start server: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Challenge '%s' is now running on port 8080\n", name)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	if currentName == "" {
		fmt.Fprintln(w, "No challenge currently running")
	} else {
		fmt.Fprintf(w, "Current challenge: %s (PID: %d)\n", currentName, currentProcess.Process.Pid)
	}
}

func main() {
	http.HandleFunc("/setchallenge", setChallengeHandler)
	http.HandleFunc("/status", statusHandler)

	port := "9000"
	log.Printf("Controller listening on port %s", port)
	log.Printf("Usage: GET /setchallenge?name=<challenge>")
	log.Printf("Valid challenges: smoke-test, primetime, pest-control")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start controller: %v", err)
	}
}
