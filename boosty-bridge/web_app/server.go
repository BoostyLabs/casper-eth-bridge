package server

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/zeebo/errs"
	"golang.org/x/sync/errgroup"

	"tricorn/internal/logger"
	"tricorn/internal/server"
	"tricorn/web_app/controllers"
)

// ensures that Server implement server.Server.
var _ server.Server = (*Server)(nil)

var (
	// Error is an error class that indicates internal http server error.
	Error = errs.Class("server")
)

// Config contains configuration for console web server.
type Config struct {
	Address        string `env:"ADDRESS"`
	GatewayAddress string `env:"GATEWAY_ADDRESS"`
	StaticDir      string `env:"STATIC_DIR"`
	// TODO: remove after new version of Casper wallet be released.
	CasperNodeAddress       string `env:"CASPER_NODE_ADDRESS"`
	CasperTokenContract     string `env:"CASPER_TOKEN_CONTRACT"`
	CasperBridgeContract    string `env:"CASPER_BRIDGE_CONTRACT"`
	ETHTokenContract        string `env:"ETH_TOKEN_CONTRACT"`
	ETHBridgeContract       string `env:"ETH_BRIDGE_CONTRACT"`
	PolygonTokenContract    string `env:"POLYGON_TOKEN_CONTRACT"`
	PolygonBridgeContract   string `env:"POLYGON_BRIDGE_CONTRACT"`
	BNBTokenContract        string `env:"BNB_TOKEN_CONTRACT"`
	BNBBridgeContract       string `env:"BNB_BRIDGE_CONTRACT"`
	AvalancheTokenContract  string `env:"AVALANCHE_TOKEN_CONTRACT"`
	AvalancheBridgeContract string `env:"AVALANCHE_BRIDGE_CONTRACT"`
	ETHGasLimit             string `env:"ETH_GAS_LIMIT"`
}

// Server represents web-app server.
//
// architecture: Endpoint
type Server struct {
	log    logger.Logger
	config Config

	listener net.Listener
	server   http.Server

	index *template.Template
}

// NewServer is a constructor for web-app server.
func NewServer(config Config, log logger.Logger, listener net.Listener) *Server {
	server := &Server{
		log:      log,
		config:   config,
		listener: listener,
	}

	router := mux.NewRouter()

	router.HandleFunc("/bridge-in", controllers.New(log).BridgeIn).Methods(http.MethodPost)

	if server.config.StaticDir != "" {
		fs := http.FileServer(http.Dir(server.config.StaticDir))
		router.PathPrefix("/static/").Handler(server.brotliMiddleware(http.StripPrefix("/static", fs)))
		router.PathPrefix("/").Handler(http.HandlerFunc(server.appHandler))
	}

	router.Handle("/robots.txt", http.HandlerFunc(server.seoHandler))

	server.server = http.Server{
		Handler: router,
	}

	return server
}

// Run starts the server that host web-app and api endpoint.
func (server *Server) Run(ctx context.Context) (err error) {
	server.log.Debug(fmt.Sprintf("running golden-gate web-app server on %s", server.config.Address))

	var group errgroup.Group
	group.Go(func() error {
		<-ctx.Done()
		server.log.Debug("tricorn web-app http server gracefully exited")
		return Error.Wrap(server.server.Shutdown(ctx))
	})
	group.Go(func() error {
		err := server.server.Serve(server.listener)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		return Error.Wrap(err)
	})

	return Error.Wrap(group.Wait())
}

// Close closes server and underlying listener.
func (server *Server) Close() error {
	server.log.Debug("tricorn web-app http server closed")
	return Error.Wrap(server.server.Close())
}

// appHandler is web app http handler function.
func (server *Server) appHandler(w http.ResponseWriter, r *http.Request) {
	header := w.Header()

	cspValues := []string{
		"base-uri 'self'", // prevent an attacker to inject a <base> element in your page in order to redirect part of your traffic to another website.
		fmt.Sprintf("connect-src 'self' %s", server.config.GatewayAddress),
		"frame-ancestors 'self'",
		"frame-src 'self'",
		"img-src 'self'",
		"media-src 'self'",
	}

	header.Set("Content-Security-Policy", strings.Join(cspValues, "; "))

	header.Set("Content-Type", "text/html; charset=UTF-8")
	// allows you to avoid MIME type sniffing by saying that the MIME types are deliberately configured.
	header.Set("X-Content-Type-Options", "nosniff")
	// to prevent any frame or iframe from integrating the page. Blocks clickjacking type attacks.
	header.Set("X-Frame-Options", "DENY")
	// Only expose the referring url when navigating around the web-app itself.
	header.Set("Referrer-Policy", "same-origin")

	err := server.parseTemplates()
	if err != nil {
		server.log.Error("unable to parse templates", Error.Wrap(err))
		return
	}

	var data struct {
		GatewayAddress          string
		CasperNodeAddress       string
		CasperTokenContract     string
		CasperBridgeContract    string
		ETHBridgeContract       string
		ETHTokenContract        string
		PolygonBridgeContract   string
		PolygonTokenContract    string
		BNBBridgeContract       string
		BNBTokenContract        string
		AvalancheTokenContract  string
		AvalancheBridgeContract string
		ETHGasLimit             string
	}

	data.GatewayAddress = server.config.GatewayAddress
	data.CasperNodeAddress = server.config.CasperNodeAddress
	data.CasperTokenContract = server.config.CasperTokenContract
	data.CasperBridgeContract = server.config.CasperBridgeContract
	data.ETHBridgeContract = server.config.ETHBridgeContract
	data.ETHTokenContract = server.config.ETHTokenContract
	data.PolygonBridgeContract = server.config.PolygonBridgeContract
	data.PolygonTokenContract = server.config.PolygonTokenContract
	data.BNBBridgeContract = server.config.BNBBridgeContract
	data.BNBTokenContract = server.config.BNBTokenContract
	data.AvalancheBridgeContract = server.config.AvalancheBridgeContract
	data.AvalancheTokenContract = server.config.AvalancheTokenContract
	data.ETHGasLimit = server.config.ETHGasLimit

	if err := server.index.Execute(w, data); err != nil {
		server.log.Error("index template could not be executed", Error.Wrap(err))
		return
	}
}

// seoHandler indicates to web crawlers which URLs should be explored on your website.
func (server *Server) seoHandler(w http.ResponseWriter, r *http.Request) {
	header := w.Header()

	header.Set("Cache-Control", "public,max-age=31536000,immutable")
	header.Set("Content-Type", mime.TypeByExtension(".txt"))
	header.Set("X-Content-Type-Options", "nosniff")

	_, err := w.Write([]byte("User-agent: *\nDisallow:\nDisallow: /cgi-bin/"))
	if err != nil {
		server.log.Error("could not return robots.txt file", err)
	}
}

// brotliMiddleware is used to compress static content using brotli to minify resources if browser support such decoding.
func (server *Server) brotliMiddleware(fn http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// allows to cache two versions of the resource on proxies: one compressed, and one
		// uncompressed. So, the clients who cannot properly decompress the files are able to access your page via a proxy, using the
		// uncompressed version. The other users will get the compressed version.
		w.Header().Set("Vary", "Accept-Encoding")

		isBrotliSupported := strings.Contains(r.Header.Get("Accept-Encoding"), "br")
		if !isBrotliSupported {
			fn.ServeHTTP(w, r)
			return
		}

		info, err := os.Stat(server.config.StaticDir + strings.TrimPrefix(r.URL.Path, "/static") + ".br")
		if err != nil {
			fn.ServeHTTP(w, r)
			return
		}

		extension := filepath.Ext(info.Name()[:len(info.Name())-3])
		w.Header().Set("Content-Type", mime.TypeByExtension(extension))
		w.Header().Set("Content-Encoding", "br")

		newRequest := new(http.Request)
		*newRequest = *r
		newRequest.URL = new(url.URL)
		*newRequest.URL = *r.URL
		newRequest.URL.Path += ".br"

		fn.ServeHTTP(w, newRequest)
	})
}

func (server *Server) parseTemplates() (err error) {
	server.index, err = template.ParseFiles(filepath.Join(server.config.StaticDir, "dist", "index.html"))
	if err != nil {
		return errs.Combine(errors.New("dist folder is not generated. use 'npm run build' command"), err)
	}

	return nil
}
