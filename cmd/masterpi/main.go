package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nathan-osman/masterpi"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "sprise"
	app.Usage = "Temperature and clock display"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "influxdb-addr",
			Value:  "http://localhost:8086",
			EnvVar: "INFLUXDB_ADDR",
			Usage:  "address of InfluxDB host",
		},
		cli.StringFlag{
			Name:   "influxdb-username",
			EnvVar: "INFLUXDB_USERNAME",
			Usage:  "username for connecting to InfluxDB",
		},
		cli.StringFlag{
			Name:   "influxdb-password",
			EnvVar: "INFLUXDB_PASSWORD",
			Usage:  "password for connecting to InfluxDB",
		},
		cli.StringFlag{
			Name:   "storage-dir",
			Value:  "/data",
			EnvVar: "STORAGE_DIR",
			Usage:  "directory for internal storage",
		},
	}
	app.Action = func(c *cli.Context) error {

		// Create the monitor
		m, err := masterpi.NewMonitor()
		if err != nil {
			return err
		}
		defer m.Close()

		// Create the updater
		u, err := masterpi.NewUpdater(m)
		if err != nil {
			return err
		}
		defer u.Close()

		// Create the uploader
		i, err := masterpi.NewUploader(
			c.String("influxdb-addr"),
			c.String("influxdb-username"),
			c.String("influxdb-password"),
			m,
		)
		if err != nil {
			return err
		}
		defer i.Close()

		// Create the relay controller
		r, err := masterpi.NewRelay()
		if err != nil {
			return err
		}
		defer r.Close()

		// Create the switch monitor
		s, err := masterpi.NewSwitch(r)
		if err != nil {
			return err
		}
		defer s.Close()

		// Create the sensor monitor
		se, err := masterpi.NewSensor(r)
		if err != nil {
			return err
		}
		defer se.Close()

		// Create the light timer
		t, err := masterpi.NewTimer(
			c.String("storage-dir"),
			r,
		)
		if err != nil {
			return err
		}
		defer t.Close()

		// Create the HTTP server
		h, err := masterpi.NewServer(r, t)
		if err != nil {
			return err
		}
		defer h.Close()

		// Wait for SIGINT or SIGTERM
		sigChan := make(chan os.Signal)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		return nil
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal: %s\n", err.Error())
	}
}
