package types

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/nicolaspernoud/ninicobox-v3-server/pkg/appserver"
	"github.com/nicolaspernoud/ninicobox-v3-server/pkg/common"
)

// App represents a app serving static content proxying a web server
type App struct {
	ID int `json:"id"`
	appserver.App
}

// ProcessApps processes apps regarding of HTTP method
func ProcessApps(w http.ResponseWriter, req *http.Request) {
	switch method := req.Method; method {
	case "GET":
		SendApps(w, req)
	case "POST":
		AddApp(w, req)
	case "DELETE":
		DeleteApp(w, req)
	default:
		http.Error(w, "method not allowed", 400)
	}
}

// SendApps send apps as response from an http requests
func SendApps(w http.ResponseWriter, req *http.Request) {
	var apps []App
	err := common.Load("./apps.json", &apps)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	json.NewEncoder(w).Encode(apps)
}

// AddApp adds an app
func AddApp(w http.ResponseWriter, req *http.Request) {
	var apps []App
	err := common.Load("./apps.json", &apps)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if req.Body == nil {
		http.Error(w, "please send a request body", 400)
		return
	}
	var newApp App
	err = json.NewDecoder(req.Body).Decode(&newApp)
	if _, ok := err.(*json.UnmarshalTypeError); !ok && err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	// Add the app only if the name doesn't exists yet
	isNew := true
	for idx, val := range apps {
		if val.ID == newApp.ID {
			apps[idx] = newApp
			isNew = false
			break
		}
	}
	if isNew {
		apps = append(apps, newApp)
	}
	err = common.Save("./apps.json", &apps)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	SendApps(w, req)
}

// DeleteApp adds an app
func DeleteApp(w http.ResponseWriter, req *http.Request) {
	var apps []App
	err := common.Load("./apps.json", &apps)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	pathElements := strings.Split(req.URL.Path, "/")
	idx, err := strconv.Atoi(pathElements[len(pathElements)-1])
	if err != nil {
		http.Error(w, "please provide an app index", 400)
		return
	}
	// Add the app only if the name doesn't exists yet
	newApps := apps[:0]
	for _, app := range apps {
		if app.ID != idx {
			newApps = append(newApps, app)
		}
	}
	err = common.Save("./apps.json", &newApps)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	SendApps(w, req)
}
