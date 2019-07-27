package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/cactus/go-statsd-client/statsd"
	"github.com/ghjnut/pingwave"
	"github.com/ghjnut/pingwave/config"
	log "github.com/sirupsen/logrus"
	"gopkg.in/urfave/cli.v1"
)

func init() {
	log.SetOutput(os.Stdout)

	log.SetLevel(log.InfoLevel)
}

func main() {
	app := cli.NewApp()

	app.Name = "pingwave"
	app.Version = "0.1"
	app.Usage = "Ping a list of endpoints and send the resulting metrics to statsd"
	app.Authors = []cli.Author{
		cli.Author{
			Name: "Lee Briggs",
		},
		cli.Author{
			Name: "Jake Pelletier",
		},
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "config-file, c",
			Value:  "config.hcl",
			Usage:  "Path to configuration file",
			EnvVar: "CONFIG_FILE",
		},
		cli.StringFlag{
			Name:   "statsd-host, sh",
			Value:  "statsd",
			Usage:  "Address of statsd listener",
			EnvVar: "STATSD_HOST",
		},
		cli.StringFlag{
			Name:   "statsd-port, sp",
			Value:  "8125",
			Usage:  "Port of statsd listener",
			EnvVar: "STATSD_PORT",
		},
		cli.StringFlag{
			Name:   "statsd-prefix, p",
			Value:  "pingwave",
			Usage:  "Top-level statsd prefix",
			EnvVar: "STATSD_PREFIX",
		},
		cli.IntFlag{
			Name:   "default-interval, i",
			Value:  10,
			Usage:  "Default interval",
			EnvVar: "DEFAULT_INTERVAL",
		},
		cli.BoolFlag{
			Name:   "debug, d",
			Usage:  "Output metrics in logs",
			EnvVar: "DEBUG",
		},
	}

	app.Action = func(c *cli.Context) (err error) {
		ctx, cancel := context.WithCancel(context.Background())

		cfg_path := c.String("config-file")

		default_interval := c.Int("default-interval")

		statsd_host := c.String("statsd-host")
		statsd_port := c.String("statsd-port")
		statsd_pfx := c.String("statsd-prefix")

		debug := c.Bool("debug")

		log.WithFields(log.Fields{
			"config-file":      cfg_path,
			"defualt-interval": default_interval,
			"statsd-host":      statsd_host,
			"statsd-port":      statsd_port,
			"statsd-prefix":    statsd_pfx,
			"debug":            debug,
		}).Info("config")

		if debug {
			log.SetLevel(log.DebugLevel)
		}

		config, err := config.Parse(cfg_path)
		if err != nil {
			return err
		}

		statsd, err := statsdClient(statsd_host, statsd_port, statsd_pfx)
		if err != nil {
			return err
		}
		defer statsd.Close()

		var wg sync.WaitGroup
		for _, tg := range config.TargetGroups {
			sp, err := statsdPinger(statsd, tg, default_interval)
			if err != nil {
				return err
			}
			go func(c context.Context) {
				defer wg.Done()
				if err := sp.RunLoop(c); err != nil {
					log.Error("failure: ", err)
				} else {
					log.Debug("success")
				}
			}(ctx)
			wg.Add(1)
		}

		// handle signals
		listen := make(chan os.Signal, 1)
		signal.Notify(listen, os.Interrupt)
		signal.Notify(listen, syscall.SIGTERM)

		sig := <-listen

		cancel()
		wg.Wait()

		log.WithFields(log.Fields{
			"signal": sig,
		}).Info("captured")

		return err
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func statsdClient(host string, port string, prefix string) (statsd.Statter, error) {
	h := fmt.Sprintf("%s:%s", host, port)
	return statsd.NewClient(h, prefix)
}

func statsdPinger(statsd statsd.Statter, tg config.TargetGroup, intvl int) (sp *pingwave.StatsdPinger, err error) {
	sp = pingwave.NewStatsdPinger(statsd, tg.Prefix)
	// if no group interval, use global
	if tg.Interval != 0 {
		sp.SetInterval(tg.Interval)
	} else {
		sp.SetInterval(intvl)
	}

	for _, t := range tg.Targets {
		ip, err := resolveIPAddr(t.Address)
		if err != nil {
			return nil, err
		}
		if err := sp.AddTarget(ip, t.Name); err != nil {
			return nil, err
		}
	}
	return sp, err
}

func resolveIPAddr(addr string) (*net.IPAddr, error) {
	return net.ResolveIPAddr("ip4:icmp", addr)
}
