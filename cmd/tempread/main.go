package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/amimof/huego"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (
	currentTemp = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "temperature",
		Help: "Temperature in hundreds of degrees Celsius",
	}, []string{"uid"})
)

func init() {
	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the info severity or above.
	log.SetLevel(log.InfoLevel)
}

func main() {
	var (
		hueHost     = flag.String("bridge", "", "Hue bridge hostname/IP address")
		hueUser     = flag.String("user", "", "Hue bridge authorized user (or give HUE_USER through env)")
		metricsPort = flag.Int("metrics-port", 2112, "Prometheus metrics port")
		metricsPath = flag.String("metrics-path", "/metrics", "Prometheus metrics path")
		debug       = flag.Bool("debug", false, "Enable debug logging")
	)

	flag.Parse()
	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	if *hueUser == "" {
		if envUser := os.Getenv("HUE_USER"); envUser != "" {
			log.Debug("using user from environment")
			*hueUser = envUser
		} else {
			log.Fatal("-user required")
		}
	}

	promHost := fmt.Sprintf(":%d", *metricsPort)

	log.WithFields(log.Fields{
		"host": promHost,
		"path": *metricsPath,
	}).Info("starting prometheus metrics")

	// set up prometheus metrics
	http.Handle(*metricsPath, promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(promHost, nil); err != nil {
			log.WithError(err).Fatal("could not start prometheus")
		}
	}()

	var bridge *huego.Bridge
	if *hueHost == "" {
		b, err := huego.Discover()
		if err != nil {
			log.WithError(err).Fatal("could not discover bridge")
		}

		bridge = b.Login(*hueUser)
	} else {
		bridge = huego.New(*hueHost, *hueUser)
	}

	firstRun := true
	for {
		if !firstRun {
			time.Sleep(time.Minute)
		}

		firstRun = false

		sensors, err := bridge.GetSensors()
		if err != nil {
			log.WithError(err).Fatal("could not get sensors")
		}

		for _, sensor := range sensors {
			if sensor.Type != "ZLLTemperature" {
				continue
			}

			sLog := log.WithFields(log.Fields{
				"sensor": sensor.Name,
				"uid":    sensor.UniqueID,
			})

			temp := -1.0
			for name, state := range sensor.State {
				if name == "temperature" {
					if t, ok := state.(float64); ok {
						temp = t
						break
					}
				}
			}

			if temp == -1.0 {
				sLog.Warning("could not get temperature reading")
				continue
			}

			currentTemp.With(prometheus.Labels{
				"uid": sensor.UniqueID,
			}).Set(temp)

			sLog.WithField("temp", temp).Info("registered temperature")
		}
	}
}
