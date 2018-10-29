package main

import (
	"./controller"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

var (
	listen   = ":9512"
	host     = ""
	userName = ""
	password = ""
	logLevel = "info"
)

func env(key, def string) string {
	if x := os.Getenv(key); x != "" {
		return x
	}

	return def
}

func init() {
	flag.StringVar(&listen, "listen", env("ESX_LISTEN", listen), "listen port")
	flag.StringVar(&host, "host", env("ESX_HOST", host), "URL ESX host ")
	flag.StringVar(&userName, "username", env("ESX_USERNAME", userName), "User for ESX")
	flag.StringVar(&password, "password", env("ESX_PASSWORD", password), "password for ESX")
	flag.StringVar(&logLevel, "log", env("ESX_LOG", logLevel), "Log levelmust be, debug or info")
	controllers.RegistredMetrics()
	getMetrics()
	flag.Parse()

	logLevel, err := log.ParseLevel(logLevel)

	if err != nil {
		fmt.Printf("log-level is bad value - `%s`: %s\n", logLevel, err)
		os.Exit(1)
	}
	log.SetLevel(logLevel)
}

func getMetrics() {
	go func() {
		loger("Start collect host metrics", "debug")
		controllers.NewVmwareHostMetrics(host, userName, password)
		loger("End collect host metrics", "debug")
	}()
	go func() {
		loger("Start collect datastore metrics", "debug")
		controllers.NewVmwareDsMetrics(host, userName, password)
		loger("End collect datastore metrics", "debug")
	}()
	go func() {
		loger("Start collect VM metrics", "debug")
		controllers.NewVmwareVmMetrics(host, userName, password)
		loger("End collect VM metrics", "debug")
	}()
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		getMetrics()
	}
	h := promhttp.Handler()
	h.ServeHTTP(w, r)
}

func loger(msg string, lvl string) {
	switch lvl {
	case "info":
		log.WithFields(log.Fields{"Message": msg}).Info()
	case "fatal":
		log.WithFields(log.Fields{"Message": msg}).Debug()
	case "debug":
		log.WithFields(log.Fields{"Message": msg}).Debug()
	}
}

func main() {
	if host == "" {
		loger("Yor must configured systemm env ESX_HOST or key -host", "fatal")
		os.Exit(1)
	}
	if userName == "" {
		loger("Yor must configured system env ESX_USERNAME or key -username", "fatal")
		os.Exit(1)
	}
	if password == "" {
		loger("Yor must configured system env ESX_PASSWORD or key -password", "debug")
		os.Exit(1)
	}

	msg := fmt.Sprintf("Exporter start on port %s", listen)
	loger(msg, "info")
	http.HandleFunc("/metrics", handler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>VMware Exporter</title></head>
			<body>
			<h1>VMware Exporter</h1>
			<p><a href="` + "/metrics" + `">Metrics</a></p>
			</body>
			</html>`))
	})

	log.Fatal(http.ListenAndServe(listen, nil))

}
