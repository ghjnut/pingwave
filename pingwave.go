package pingwave

import (
	"context"
	"fmt"
	"net"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/cactus/go-statsd-client/statsd"
	"github.com/tatsushid/go-fastping"
)

func NewStatsdPinger(statter statsd.Statter, prefix string) (sp *StatsdPinger) {
	pinger := fastping.NewPinger()
	targets := make(map[string]string)
	responded := make(map[string]bool)

	return &StatsdPinger{
		statsdClient: statter.NewSubStatter(prefix),

		//prefix:    prefix,
		pinger:    pinger,
		targets:   targets,
		responded: responded,
	}
}

type StatsdPinger struct {
	statsdClient statsd.SubStatter

	//prefix    string
	pinger    *fastping.Pinger
	targets   map[string]string
	responded map[string]bool
}

func (sp *StatsdPinger) SetInterval(i int) {
	sp.pinger.MaxRTT = time.Duration(i) * time.Second
}

func (sp *StatsdPinger) Interval() time.Duration {
	return sp.pinger.MaxRTT
}

// TODO make this take an IPaddr
func (sp *StatsdPinger) AddTarget(ip *net.IPAddr, name string) error {
	if err := statsd.CheckName(name); err != nil {
		return err
	}
	sp.pinger.AddIPAddr(ip)
	sp.targets[ip.String()] = name
	return nil
}

func (sp *StatsdPinger) RunLoop(ctx context.Context) error {
	// single response received
	sp.pinger.OnRecv = func(ip *net.IPAddr, t time.Duration) {
		sp.onRecv(ip, t)
	}

	// end of loop
	sp.pinger.OnIdle = func() {
		sp.onIdle()
	}

	sp.pinger.RunLoop()

	// wait for pinger or ctx done
	select {
	case <-sp.pinger.Done():
	case <-ctx.Done():
		sp.pinger.Stop()
		// wait until it's done
		<-sp.pinger.Done()
	}

	return sp.pinger.Err()
}

func (sp *StatsdPinger) onRecv(ip *net.IPAddr, t time.Duration) error {
	log.WithFields(log.Fields{
		"ip":       ip.String(),
		"duration": t,
		"name":     sp.targets[ip.String()],
	}).Debug("response")

	sp.responded[ip.String()] = true
	return sp.recordResponse(ip.String(), t)
}

func (sp *StatsdPinger) onIdle() {
	for ip, name := range sp.targets {
		if !sp.responded[ip] {
			log.WithFields(log.Fields{
				"name": name,
				"ip":   ip,
			}).Debug("timeout")

			if err := sp.recordTimeout(ip); err != nil {
				log.WithFields(log.Fields{
					"name": name,
					"ip":   ip,
				}).Error("failed to record timeout")
				//sp.pinger.Stop()
			}
		}
		// reset
		sp.responded[ip] = false
	}
}

func (sp *StatsdPinger) recordResponse(ip string, t time.Duration) (err error) {
	// send a zeroed failed metric, because we succeeded!
	//err := sp.statsdClient.Inc(fmt.Sprintf("%s.failed", sp.targets[ip]), 0, 1)
	if err = sp.recordFail(ip, false); err != nil {
		return err
	}

	err = sp.statsdClient.TimingDuration(sp.targets[ip], t, 1)
	//if err != nil {
	//	log.Error(fmt.Sprintf("Error sending metric: %+v", err)) }
	//err = statsdClient.TimingDuration(fmt.Sprintf("%s%s.timer", prefix, outputLabel), r.rtt, 1)
	return err
}

func (sp *StatsdPinger) recordTimeout(ip string) error {
	return sp.recordFail(ip, true)
}

func (sp *StatsdPinger) recordFail(ip string, failed bool) error {
	f := 0
	if failed {
		f = 1
	}
	return sp.statsdClient.Gauge(fmt.Sprintf("%s.failed", sp.targets[ip]), int64(f), 1)
}
