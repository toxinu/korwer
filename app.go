package main

import (
    "io/ioutil"

    "os"
    "os/user"

    "log"
    "fmt"
    "strconv"
    "net/http"
    "crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"

    "github.com/gorilla/mux"
    "github.com/BurntSushi/toml"
)

func NewJsonResponse(success bool, message string) *JsonResponse {
	return &JsonResponse{
		Success: success,
		Message: message,
	}
}

type JsonResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type JsonRequest struct {
	Site string `json:"site"`
    Secret string `json:"secret"`
}

func buildAndDeployWebsite(site *Site) {
    c := make(chan string)
    var BuildProcess *Process
    var DeployProcess *Process

    if site.BuildCommand != "" {
        BuildProcess = &Process{
            Command: fmt.Sprintf("cd %s && %s", site.Path, site.BuildCommand),
            Output: c}
    }
    if site.DeployCommand != "" {
        DeployProcess = &Process{
            Command: fmt.Sprintf("cd %s && %s", site.Path, site.DeployCommand),
            Output: c}
    }

    if site.BuildCommand != "" && site.DeployCommand != "" {
        go BuildProcess.Run(*DeployProcess)
    } else if site.BuildCommand != "" && site.DeployCommand == "" {
        go BuildProcess.Run()
    } else if site.BuildCommand == "" && site.DeployCommand != "" {
        go DeployProcess.Run()
    }
    go Collect(c)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	jsonResponse, _ := json.Marshal(NewJsonResponse(true, "Hello world"))
	fmt.Fprintf(w, string(jsonResponse))
}

func buildHandler(w http.ResponseWriter, r *http.Request) {
    jsonDecoder := json.NewDecoder(r.Body)
    var jsonRequest JsonRequest
    err := jsonDecoder.Decode(&jsonRequest)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    var site Site
    siteName := jsonRequest.Site
    siteSecret := jsonRequest.Secret
    siteExists := false
    for _, v := range CONFIG.Site {
        if siteName == v.Name {
            site = v
            siteExists = true
        }
    }
    var jsonResponse []byte
    if siteExists {
        if siteSecret != site.Secret {
            jsonResponse, _ = json.Marshal(NewJsonResponse(
                false, "Secret not valid"))
        } else {
            buildAndDeployWebsite(&site)
            jsonResponse, _ = json.Marshal(NewJsonResponse(
                true, fmt.Sprintf("Build launched for %s", site.Name)))
        }
    } else {
        jsonResponse, _ = json.Marshal(NewJsonResponse(
            false, fmt.Sprintf("Site (%s) does not exist", siteName)))
    }
	fmt.Fprintf(w, string(jsonResponse))
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
    if r.Header.Get("X-GitHub-Event") == "" {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

    vars := mux.Vars(r)
    var site Site
    siteName := vars["site"]
    siteExists := false
    for _, v := range CONFIG.Site {
        if siteName == v.Name {
            siteExists = true
            site = v
        }
    }
    var jsonResponse []byte
    if siteExists {
        sig := r.Header.Get("X-Hub-Signature")
        if sig == "" {
			http.Error(w, "403 Forbidden - Missing X-Hub-Signature required for HMAC verification", http.StatusForbidden)
			return
        }
        mac := hmac.New(sha1.New, []byte(site.Secret))
		mac.Write(body)
		expectedMAC := mac.Sum(nil)
		expectedSig := "sha1=" + hex.EncodeToString(expectedMAC)
		if !hmac.Equal([]byte(expectedSig), []byte(sig)) {
			http.Error(w, "403 Forbidden - HMAC verification failed", http.StatusForbidden)
			return
		}
        buildAndDeployWebsite(&site)
    } else {
        jsonResponse, _ = json.Marshal(NewJsonResponse(
            false, fmt.Sprintf("Site (%s) does not exist", vars["site"])))
    }
    fmt.Fprintf(w, string(jsonResponse))
}

func listHandler(w http.ResponseWriter, r *http.Request) {
    jsonResponse, _ := json.Marshal(&CONFIG.Site)
    fmt.Fprintf(w, string(jsonResponse))
}

// Globals
var CONFIG Config

func main() {
    Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)

    configurationPath := "./korwer.toml"
    _, err := toml.DecodeFile(configurationPath, &CONFIG)
    if err != nil {
        log.Fatal(err)
    }

    currentUser, _ := user.Current()
    fmt.Println("##########\n# Korwer #\n##########")
    fmt.Printf("Port         : %s\n", strconv.Itoa(CONFIG.Settings.Port))
    fmt.Printf("User         : %s\n", currentUser.Username)
    fmt.Printf("Configuration: %s\n\n", configurationPath)
    for _, s := range(CONFIG.Site) {
        fmt.Printf("- %s\n", s.Name)
        fmt.Printf("  | %s\n", s.Path)
        if s.BuildCommand != "" {
            fmt.Printf("  | %s\n", s.BuildCommand)
        }
        if s.DeployCommand != "" {
            fmt.Printf("  | %s\n", s.DeployCommand)
        }
    }

    r := mux.NewRouter()
    r.HandleFunc("/", indexHandler)
    r.HandleFunc("/list", listHandler)
    r.HandleFunc("/build", buildHandler).Methods("POST")
    r.HandleFunc("/webhook/{site:[[:alnum:].-]+}", webhookHandler).Methods(
        "POST")


	fmt.Printf("\nListening on :%s... <ctrl-c to stop>", strconv.Itoa(
        CONFIG.Settings.Port))
    http.Handle("/", r)
	http.ListenAndServe(fmt.Sprintf(
        ":%s", strconv.Itoa(CONFIG.Settings.Port)), nil)
}
