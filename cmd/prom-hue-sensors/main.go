package main

import (
	"context"
	"errors"
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

const (
	// username for this app in the Hue bridge
	hueBridgeUsername = "prom-hue-sensors"
	registerInterval  = 5 * time.Second
)

var (
	sensorStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sensor_status",
		Help: "Every parsable status for each sensor",
	}, []string{"uid", "name", "type"})

	ignoreStates = map[string]bool{
		"lastupdated": true,
	}
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
		hueHost         = flag.String("bridge", "", "Hue bridge hostname/IP address")
		hueUser         = flag.String("user", "", "Hue bridge authorized user (or give HUE_USER through env)")
		metricsPort     = flag.Int("metrics-port", 2112, "Prometheus metrics port")
		metricsPath     = flag.String("metrics-path", "/metrics", "Prometheus metrics path")
		debug           = flag.Bool("debug", false, "Enable debug logging")
		register        = flag.Bool("register", false, "Register new Hue Hue bridge authorized user")
		registerTimeout = flag.Duration("register-timeout", time.Minute, "timeout for waiting on registering the user")
		userKeyPath     = flag.String("user-key-path", "", "path to stored registered user key")
	)

	flag.Parse()
	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	var bridge *huego.Bridge
	var err error
	// if no bridge address is given, try to discover it
	if *hueHost == "" {
		// div
		bridge, err = huego.Discover()
		if err != nil {
			log.WithError(err).Fatal("could not discover bridge")
		}

		log.WithField("bridge", bridge.Host).Info("discovered bridge")
	} else {
		// create a new bridge with given address
		bridge = &huego.Bridge{
			Host: *hueHost,
		}
	}

	// when registering, we're trying to create a hue user
	// this requires the operator of the hue bridge to physically press the link
	// button on the bridge
	// after the set timeout, the process fails
	if *register {
		ctx, cancel := context.WithTimeout(context.Background(), *registerTimeout)

		for {
			userKey, err := bridge.CreateUserContext(ctx, hueBridgeUsername)
			if err != nil {
				log.WithError(err).Debug("registration attempt failed")

				// if the context timed out, abort
				if errors.Is(err, context.DeadlineExceeded) {
					log.Fatal("failed to register user, try again later")
				}

				// see if the error the API is returning is about pressing the
				// button
				var apiErr *huego.APIError
				if errors.As(err, &apiErr) && apiErr.Type == 101 {
					log.Error("failed to create user, press key on Hue bridge")
					time.Sleep(registerInterval)

					continue
				}

				// we didn't catch this error, abort
				log.WithError(err).Fatal("unhandled registration error")
			}

			cancel()
			if *userKeyPath == "" {
				log.WithField("userKey", userKey).Info("registration successful, use this key as HUE_USER or with -user")
			} else {
				log.Info("registration successful")
			}

			if *userKeyPath != "" {
				if err := os.WriteFile(*userKeyPath, []byte(userKey), os.FileMode(0644)); err != nil {
					log.WithError(err).Fatalf("could not write user key to %s", *userKeyPath)
				}

				log.WithField("userKeyPath", userKeyPath).Info("saved user key to file")
			}

			break
		}

		log.Info("done registering")

		return
	}

	if *hueUser == "" {
		// if no hue user is given, try to get it from the environment variable HUE_USER
		if envUser := os.Getenv("HUE_USER"); envUser != "" {
			log.Debug("using user from environment")
			*hueUser = envUser
		} else if *userKeyPath != "" {
			// try to get the user key from the given user key path
			if data, err := os.ReadFile(*userKeyPath); err == nil {
				log.WithField("path", *userKeyPath).Debug("using user from user key path")
				*hueUser = string(data)
			}
		}

		if *hueUser == "" {
			log.Fatal("-user required")
		}
	}

	bridge.User = *hueUser

	// start prometheus server
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
			sLog := log.WithFields(log.Fields{
				"sensor": sensor.Name,
				"uid":    sensor.UniqueID,
			})

			sensorLabel := prometheus.Labels{
				"uid":  sensor.UniqueID,
				"name": sensor.Name,
			}

			for name, state := range sensor.State {
				// should we just ignore this state name
				if _, ok := ignoreStates[name]; ok {
					log.WithField("type", name).Debug("ignoring sensor state")

					continue
				}

				sensorLabel["type"] = name
				ssLog := sLog.WithFields(log.Fields{
					"type":    name,
					"reading": state,
				})

				// decide how we should interpret the value from this sensor
				switch s := state.(type) {
				case float64:
					sensorStatus.With(sensorLabel).Set(s)
				case bool:
					f := 0.0
					if s {
						f = 1.0
					}
					sensorStatus.With(sensorLabel).Set(f)
				default:
					ssLog.Debug("could not register sensor state")
					continue
				}

				ssLog.Info("registered sensor state")
			}
		}
	}
}
