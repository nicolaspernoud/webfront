package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/nicolaspernoud/ninicobox-v3-server/pkg/appserver"

	"github.com/nicolaspernoud/ninicobox-v3-server/pkg/common"
	"github.com/nicolaspernoud/ninicobox-v3-server/pkg/log"
	"github.com/nicolaspernoud/ninicobox-v3-server/pkg/security"
	"github.com/nicolaspernoud/webfront/internal/types"

	"golang.org/x/crypto/acme/autocert"
)

var (
	appsFile     = flag.String("apps", "", "apps definition `file`")
	letsCacheDir = flag.String("letsencrypt_cache", "./letsencrypt_cache", "let's encrypt cache `directory`")
	mainHostName = flag.String("hostname", "localhost", "Main hostname, defaults to localhost")
	logFile      = flag.String("log_file", "", "Optional file to log to, defaults to no file logging")
	httpsPort    = flag.Int("https_port", 443, "HTTPS port to serve on (defaults to 443)")
	httpPort     = flag.Int("http_port", 80, "HTTP port to serve on (defaults to 80), only used to get let's encrypt certificates")
	debugMode    = flag.Bool("debug", false, "Debug mode, disable let's encrypt, enable CORS and more logging")
	frameSource  = "localhost"
)

func main() {

	// Parse the flags
	flag.Parse()

	// Initialize logger
	if *logFile != "" {
		log.SetFile(*logFile)
		// Properly close the log on exit
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigs
			log.Logger.Println("--- Closing log ---")
			log.CloseFile()
			os.Exit(0)
		}()
	}
	log.Logger.Println("--- Server is starting ---")
	log.Logger.Printf("Main hostname is %v\n", *mainHostName)

	// Create the server
	rootMux, hostPolicy := createRootMux(*httpsPort, &frameSource, *mainHostName, *appsFile)

	// Serve locally with https on debug mode or with let's encrypt on production mode
	if *debugMode {
		log.Logger.Fatal(http.ListenAndServeTLS(":"+strconv.Itoa(*httpsPort), "./dev_certificates/localhost.crt", "./dev_certificates/localhost.key", security.CorsMiddleware(log.Middleware(rootMux), &frameSource)))
	} else {
		certManager := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			Cache:      autocert.DirCache(*letsCacheDir),
			HostPolicy: hostPolicy,
		}

		server := &http.Server{
			Addr:    ":" + strconv.Itoa(*httpsPort),
			Handler: rootMux,
			TLSConfig: &tls.Config{
				GetCertificate: certManager.GetCertificate,
			},
			ReadTimeout:  30 * time.Minute, // in case of upload
			WriteTimeout: 5 * time.Hour,    // in case of download
			IdleTimeout:  120 * time.Second,
		}

		go http.ListenAndServe(":"+strconv.Itoa(*httpPort), certManager.HTTPHandler(nil))
		server.ListenAndServeTLS("", "")
	}
}

func createRootMux(port int, frameSource *string, mainHostName string, appsFile string) (http.Handler, func(ctx context.Context, host string) error) {
	// Create the app handler
	appServer, err := appserver.NewServer(appsFile, port, *frameSource, mainHostName)
	if err != nil {
		log.Logger.Fatal(err)
	}
	var appHandler http.Handler = appServer

	// Create the main handler
	mainMux := http.NewServeMux()
	// Expose apps API
	mainMux.HandleFunc("/apps/", types.ProcessApps)
	mainMux.Handle("/reload", reloadApps(appServer))
	// Serve static files falling back to serving index.html
	mainMux.Handle("/", http.FileServer(&common.FallBackWrapper{Assets: http.Dir("web")}))

	// Put it together into the main handler
	rootMux := http.NewServeMux()
	rootMux.Handle(mainHostName+"/", mainMux)
	rootMux.Handle("/", appHandler)
	return rootMux, appServer.HostPolicy
}

func reloadApps(appServer *appserver.Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := appServer.LoadApps("./apps.json")
		if err != nil {
			http.Error(w, err.Error(), 400)
		} else {
			fmt.Fprintf(w, "apps reloaded")
		}

	})
}
