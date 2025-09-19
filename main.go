// This smol program is the Go HTTP server in static mode
// in order to provide the smallest docker image possible

package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	// Def of flags
	portPtr                  = flag.Int("port", 8080, "The listening port")
	basePath                 = flag.String("path", "/srv/http", "The path for the static files")
	vhostPrefix              = flag.String("vhost", "labs", "The prefix for locating lightweight virtual hosted subdomains, or vhosts. E.g. 'labs' will serve the files at /srv/http/labs/tango when someone visits http://tango.your.tld")

	logLevel                 = flag.String("log-level", "info", "default: info - What level of logging to run, info logs all requests (error, warn, info, debug)")
)

func setupLogger(logLevel string) {
	switch logLevel {
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

type NotFoundResponseWriter struct {
	http.ResponseWriter
	silenceWrites bool
	errorBody string
}

func (w *NotFoundResponseWriter) WriteHeader(status int) {
	if status == http.StatusNotFound {
		// http.server will try to force a text/plain content type
		// so overwrite that before sending the HTTP status line
		w.ResponseWriter.Header().Set("content-type", "text/html; charset=utf8")
		w.ResponseWriter.WriteHeader(status)

		log.Debug().Msgf("Sending custom 404 response")

		w.ResponseWriter.Write([]byte(w.errorBody))
		w.silenceWrites = false
		return
	}

	w.ResponseWriter.WriteHeader(status)
}

func (w *NotFoundResponseWriter) Write(b []byte) (int, error) {
    if w.silenceWrites {
        // Allow the caller to write the bytes, but do nothing
        return len(b), nil
    }

    return w.ResponseWriter.Write(b)
}

func logRequestAnd404(h http.Handler, errorBody string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info().Str("Method", r.Method).Str("Path", r.URL.Path).Msg("inbound-request")
		h.ServeHTTP(&NotFoundResponseWriter{ResponseWriter: w, errorBody: errorBody}, r)
	})
}

func main() {
	flag.Parse()

	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Set a pretty console output
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	//setting up the logger
	setupLogger(*logLevel)
	log.Debug().Str("Logging Level", zerolog.GlobalLevel().String()).Msg("Logger setup...")

	port := ":" + strconv.FormatInt(int64(*portPtr), 10)

	var fileSystem http.FileSystem = http.Dir(*basePath)
	log.Info().Str("path", *basePath).Msg("Document root")


	errorPage, err := fileSystem.Open("/404.html")
	var errorBody = "No resource found at the requested path"

	if (err != nil) {
		log.Debug().Err(err).Msg("No file found at /404.html")
	} else {
		errorBytes, err := io.ReadAll(errorPage)
		if (err != nil) {
			log.Error().Err(err).Msg("Could not read /404.html")
		}
		log.Info().Msg("Using custom 404 page /404.html")
		errorBody = string(errorBytes)
	}
	log.Debug().Str("404", errorBody).Msg("Custom error handler text")

	handler := http.FileServer(fileSystem)

	// VirtualHost handler here
	//   - this won't allow for auth/headers/etc to be
	//     configured differently on different vhosts
	//     if you want that, use nginx instead
	handler = vhostify(handler, fileSystem)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Msg("Returning Service Health")
		log.Info().Int("http_code", http.StatusOK).Str("path", r.URL.Path).Msg("healthcheck")
		fmt.Fprintf(w, "Ok")
	})

	handler = logRequestAnd404(handler, errorBody)
	http.Handle("/now", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dest := *r.URL
		dest.Path = "/now.html"

		http.Redirect(w, r, dest.String(), http.StatusFound)
		log.Info().Int("http_code", http.StatusFound).Str("Method", r.Method).Str("From", r.URL.Path).Str("To", dest.Path).Msg("Request Redirected")
	}))

	http.Handle("/", handler)

	log.Info().Msgf("Listening at http://0.0.0.0%v", port)
	if err := http.ListenAndServe(port, nil); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("Server startup failed")
	}

}
