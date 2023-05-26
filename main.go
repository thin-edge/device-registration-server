package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/exp/slog"
)

// Build data
var buildVersion string = "0.0.0"
var buildBranch string = "unknown"

type RegistrationRequest struct {
	Name                string   `json:"name"`
	SupportedOperations []string `json:"supportedOperations"`
}

type RegistrationResponse struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Parent string `json:"parent"`
}

type ErrorResponse struct {
	Message string `json:"message"`
	Details string `json:"details"`
}

func writeJSONResponse(w http.ResponseWriter, statusCode int, response any) {
	body, err := json.Marshal(response)
	if err != nil {
		slog.Error("Could not marshal response", err)
	}
	w.WriteHeader(statusCode)
	if _, err := w.Write(body); err != nil {
		log.Printf("Could not write to body. %s", err)
	}
}

func registerHandler(id string, baseDir string, sep string, usePrefix bool) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		request := &RegistrationRequest{}
		if err := decoder.Decode(request); err != nil {
			writeJSONResponse(w, http.StatusUnsupportedMediaType, &ErrorResponse{
				Message: "Could not parse registration request",
				Details: err.Error(),
			})
			return
		}

		deviceID := id
		if deviceID == "" {
			value, err := GetDeviceID()
			if err != nil {
				writeJSONResponse(w, http.StatusBadRequest, ErrorResponse{
					Message: "Could not get device id",
					Details: err.Error(),
				})
				return
			}
			deviceID = strings.TrimSpace(value)
		}

		slog.Info("Received registration request.", "name", request.Name, "supportedOperations", request.SupportedOperations, "deviceID", deviceID)

		if deviceID == "" {
			// Fail early if the device.id has not been set
			writeJSONResponse(w, http.StatusNotFound, ErrorResponse{
				Message: "tedge device.id is not set",
				Details: "Most likely the device certificate has not been created for tedge just yet.",
			})
			return
		}

		// Disallow short device names (technically not an issue, but it is usually a sign that the user is being lazy and using names which will cause future problems)
		if len(request.Name) < 3 {
			writeJSONResponse(w, http.StatusUnprocessableEntity, ErrorResponse{
				Message: "Invalid name for a child",
				Details: fmt.Sprintf("Child device names must be at least 4 characters long. got=%s (len=%d)", request.Name, len(request.Name)),
			})
			return
		}

		// Set child name. Either prefix it with the device ID or just use the value as is
		childID := request.Name
		if usePrefix {
			childID = strings.Join([]string{deviceID, request.Name}, sep)
		}

		if err := RegisterDevice(childID, baseDir, request.SupportedOperations); err != nil {
			writeJSONResponse(w, http.StatusNotFound, ErrorResponse{
				Message: "Could not register device",
				Details: err.Error(),
			})
			return
		}

		writeJSONResponse(w, http.StatusOK, RegistrationResponse{
			Name:   request.Name,
			ID:     childID,
			Parent: deviceID,
		})
	}
}

func touchFile(name string) error {
	_, err := os.Stat(name)
	if os.IsExist(err) {
		// Do nothing
		return nil
	}

	if os.IsNotExist(err) {
		file, fileErr := os.Create(name)
		if fileErr != nil {
			return fileErr
		}
		defer file.Close()
		return nil
	}

	return err
}

func RegisterDevice(id string, baseDir string, supportedOperations []string) error {
	operationsDir := filepath.Join(baseDir, "operations", "c8y")
	if _, err := os.Stat(operationsDir); os.IsNotExist(err) {
		slog.Error("operations directory does not exist.", "path", baseDir)
		return err
	}

	childDIR := filepath.Join(operationsDir, id)
	if err := os.MkdirAll(childDIR, os.ModePerm); err != nil {
		slog.Error("Could not create operations directory for child", slog.String("error", err.Error()), slog.String("path", childDIR))
		return err
	}

	for _, opType := range supportedOperations {
		if err := touchFile(path.Join(childDIR, opType)); err != nil {
			return err
		}
	}

	return nil
}

// Get device id via the tedge cli
func GetDeviceID() (string, error) {
	cmd, err := exec.Command("tedge", "config", "get", "device.id").Output()
	if err != nil {
		slog.Error("Could not get device.id from tedge.", "error", err)
	}
	return string(bytes.TrimSpace(cmd)), err
}

func main() {
	var port int
	var bindAddr string
	var deviceID string
	var configDir string
	var nameSeparator string
	var showVersion bool
	var usePrefix bool

	flag.IntVar(&port, "port", 9000, "Port")
	flag.StringVar(&bindAddr, "bind", "", "Bind address to which the http server should attach to. It listens on all adapters by default.")
	flag.StringVar(&deviceID, "device-id", "", "Use static device id instead of using the tedge cli")
	flag.StringVar(&configDir, "config-dir", "/etc/tedge", "thin-edge.io base configuration directory")
	flag.StringVar(&nameSeparator, "separator", "_", "Device name separator")
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.BoolVar(&usePrefix, "use-prefix", true, "Prefix the child id with the main device id")

	// Support setting flags via environment variables
	flag.VisitAll(func(f *flag.Flag) {
		envName := "REGISTRATION_" + strings.ReplaceAll(strings.ToUpper(f.Name), "-", "_")
		if value := os.Getenv(envName); value != "" {
			f.Value.Set(value)
		}
	})
	flag.Parse()

	if showVersion {
		fmt.Fprintf(os.Stdout, "%s (%s)\n", buildVersion, buildBranch)
		os.Exit(0)
	}

	listenOn := fmt.Sprintf("%s:%d", bindAddr, port)
	slog.Info("Starting registration service.", "listen", listenOn, "deviceID", deviceID)
	http.HandleFunc("/register", registerHandler(deviceID, configDir, nameSeparator, usePrefix))
	http.ListenAndServe(listenOn, nil)
}
