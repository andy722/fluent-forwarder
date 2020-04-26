package main

import (
	"context"
	"github.com/andy722/fluent-forwarder/metrics"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

import _ "net/http/pprof"

func init() {
	runtime.GOMAXPROCS(1)
}

func main() {
	var forwarderOptions = ForwarderOptions{}

	var showHelp = false
	var httpBindAddress string
	var logLevel = "info"

	flag.BoolVarP(&showHelp, "help", "h", showHelp, "Show help message")
	flag.StringVarP(&logLevel, "log", "", logLevel, "Logging level (trace,debug,info,warn,error,fatal)")

	flag.StringVarP(&forwarderOptions.InputFwd, "input-fwd", "", "", "Listen for incoming fluentd traffic. \nEither tcp://HOST:PORT or unix://SOCKET_PATH")
	flag.Lookup("input-fwd").NoOptDefVal = "tcp://0.0.0.0:24224"

	flag.StringVarP(&forwarderOptions.WalPath, "wal", "", "/opt/logs/.buffer", "Buffer directory")

	flag.StringVarP(&forwarderOptions.LogPath, "target-log", "", "", "Target log directory")
	flag.Lookup("target-log").NoOptDefVal = "/opt/logs"

	flag.StringVarP(&forwarderOptions.ForwardAddr, "target-fwd", "", "", "Target forwarder, tcp://FLUENT_HOST:PORT")

	flag.StringVarP(&httpBindAddress, "http", "", "", "Profiling and monitoring URI (ex. 0.0.0.0:24225)")
	flag.Lookup("http").NoOptDefVal = "0.0.0.0:24225"

	flag.Parse()

	if showHelp {
		flag.Usage()
		os.Exit(0)
	}

	initLogging(logLevel)

	if httpBindAddress != "" {
		go startHTTP(httpBindAddress)
	}

	log.Infof("Starting forwarder: %+v", forwarderOptions)
	app, err := NewForwarder(forwarderOptions)
	if err != nil {
		log.Fatalf("Startup failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go waitOnSignal(cancel)

	app.Run(ctx)
}

func initLogging(loglevel string) {
	{
		actualLevel := log.InfoLevel
		if parsedLevel, err := log.ParseLevel(loglevel); err != nil {
			log.Warnf("Invalid log level: %v, using default %v", loglevel, actualLevel)

		} else {
			actualLevel = parsedLevel
			log.StandardLogger().Logf(actualLevel, "%v logging enabled", actualLevel)
		}

		log.SetLevel(actualLevel)
	}

	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		//ForceColors:   true,
		FullTimestamp: true,
	})
}

func startHTTP(bindAddress string) {
	http.Handle("/metrics", metrics.Handler)

	log.Info("Serving metrics at ", bindAddress)
	if err := http.ListenAndServe(bindAddress, nil); err != nil {
		log.Fatalf("Failed binding to [%s]: %+v\n", bindAddress, err)
	}
}

func waitOnSignal(cancel context.CancelFunc) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)

	sig := <-signals
	log.Printf("Caught %v", sig)
	cancel()
}
