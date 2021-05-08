package prosody_httpupload

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"sync"
)

// Server is an HTTP server that complies with https://modules.prosody.im/mod_http_upload_external.html
type Server struct {
	config Config
	auth   *authenticator
	router *mux.Router
	mtx    *sync.Mutex // Prevent race conditions when checking for file existence
}

// Config is the configuration for the HTTP server
type Config struct {
	ListenAddress string `mapstructure:"listen-address"`
	Secret        string `mapstructure:"secret"`
	StoragePath   string `mapstructure:"storage-path"`
}

// check performs basic validation on the config
func (c *Config) check() error {
	if c.ListenAddress == "" {
		return fmt.Errorf("listen-address must be non empty, found '%s'", c.ListenAddress)
	}

	if c.Secret == "" {
		return fmt.Errorf("secret must be non empty, found '%s'", c.Secret)
	}

	err := os.MkdirAll(c.StoragePath, 0755)
	if err != nil {
		return fmt.Errorf("could not crate storage path '%s': %w", c.StoragePath, err)
	}

	return nil
}

// New creates a Server given its config
func New(c Config) (*Server, error) {
	s := &Server{
		config: c,
		router: mux.NewRouter(),
		auth:   newAuthenticator(c.Secret),
		mtx:    &sync.Mutex{},
	}

	s.setupRoutes()

	return s, c.check()
}

// setupRoutes configures mux to route requests to the desired handlers
func (s *Server) setupRoutes() {
	// 404 for `/`
	s.router.Handle("/", http.NotFoundHandler())

	// GET/HEAD route, delegate to fileServer
	s.router.NewRoute().
		Methods(http.MethodGet, http.MethodHead).
		Handler(http.FileServer(http.Dir(s.config.StoragePath)))

	// PUT route, go through authenticate middleware and then to putFile
	s.router.NewRoute().
		Methods(http.MethodPut).
		Handler(s.auth.authenticate(
			http.HandlerFunc(s.putFile),
		))

	s.router.NotFoundHandler = http.NotFoundHandler()
}

// Run starts the server an blocks
func (s *Server) Run() error {
	log.Printf("Starting up Prosody HTTP Upload server in %s", s.config.ListenAddress)
	return http.ListenAndServe(s.config.ListenAddress, s.router)
}

// putFile is a handler that creates a file in disk from the specified request path and body
func (s *Server) putFile(rw http.ResponseWriter, r *http.Request) {
	// Resolve any dots in file name
	filepath := path.Clean(r.URL.Path)

	if filepath == "." {
		// Disallow writing to root path
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	// Check for file existence, create if it does not exist
	// createFile will append the sanitized path to the storage path.
	file, err := s.createFile(filepath)
	if errors.Is(err, os.ErrExist) {
		// File already exists, return conflict
		rw.WriteHeader(http.StatusConflict)
		return
	} else if err != nil {
		// Unknown error, return 500
		log.Printf("error processing PUT %s: %v", filepath, err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Defer close body and log error
	defer func() {
		err := r.Body.Close()
		log.Printf("error closing file while handling %v: %v", r, err)
	}()

	// Write file to disk
	_, err = io.Copy(file, r.Body)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		log.Printf("error writing %s: %v", filepath, err)
		return
	}

	rw.WriteHeader(http.StatusCreated)
}

// createFile checks atomically for existence of a file, and returns a handle to a new one if it did not exist
// If the file was already found on disk, it returns nil and os.ErrExist.
func (s *Server) createFile(filepath string) (*os.File, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	// Produce full path by appending sanitized path to storage path
	fullpath := path.Join(s.config.StoragePath, path.Clean(filepath))

	var err error
	// Create any subdir that is missing
	err = os.MkdirAll(path.Dir(fullpath), 0755)
	if err != nil {
		return nil, fmt.Errorf("could not create dir for %s: %w", fullpath, err)
	}

	// Attempt to stat the specified path
	_, err = os.Stat(fullpath)
	if errors.Is(err, os.ErrNotExist) {
		// If file does not exist, create it and return happily
		return os.Create(fullpath)
	}

	// Return unknown error when attempting to stat
	if err != nil {
		return nil, fmt.Errorf("could not stat %s: %w", fullpath, err)
	}

	// stat succeeded, return os.ErrExist
	return nil, os.ErrExist
}
