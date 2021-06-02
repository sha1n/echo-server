package startcmd

import (
	"github.com/gin-gonic/gin"
	"github.com/sha1n/hako/cmd/hako/http"
	"github.com/sha1n/hako/cmd/hako/utils"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// Config start command configuration
type Config struct {
	ServerPort     int
	EchoPath       string
	Verbose        bool
	VerboseHeaders bool
	Delay          int32
}

// StartAsync starts an echo server in the background and returns immediately.
func StartAsync(config Config) {
	server := createHTTPServer(config)
	server.StartAsync()
}

// Start starts an echo server in the background and awaits shutdown signal.
func Start(config Config) {
	StartAsync(config)

	awaitShutdownSig()
}

func awaitShutdownSig() {
	quitChannel := make(chan os.Signal)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Waiting for shutdown signal...")

	<-quitChannel
}

func createHTTPServer(config Config) http.Server {
	router := createGinEngine(config)

	server := http.NewServer(config.ServerPort, router)

	stopServerAsync := func() {
		server.StopAsync()
	}

	log.Println("Registering signal listeners for graceful HTTP server shutdown..")
	utils.RegisterShutdownHook(utils.NewSignalHook(syscall.SIGTERM, stopServerAsync))
	utils.RegisterShutdownHook(utils.NewSignalHook(syscall.SIGKILL, stopServerAsync))

	return server
}

func createGinEngine(config Config) *gin.Engine {
	var router *gin.Engine
	if config.Verbose {
		router = http.NewDefaultEngine()
	} else {
		router = http.NewSilentEngine()
	}
	registerHandlers(router, "/echo", echoHandlerWith(config.Verbose, config.VerboseHeaders, config.Delay))

	if config.EchoPath != "" {
		registerHandlers(router, config.EchoPath, echoHandlerWith(config.Verbose, config.VerboseHeaders, config.Delay))
	}

	return router
}

func registerHandlers(router *gin.Engine, path string, handler func(ctx *gin.Context)) {
	router.GET(path, handler)
	router.POST(path, handler)
	router.PUT(path, handler)
	router.DELETE(path, handler)
	router.PATCH(path, handler)
	router.HEAD(path, handler)
	router.OPTIONS(path, handler)
}
