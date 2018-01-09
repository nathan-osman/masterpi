package masterpi

import (
	"time"

	"github.com/nathan-osman/go-rpigpio"
	"github.com/sirupsen/logrus"
)

type Switch struct {
	relay       *Relay
	pin         *rpi.Pin
	log         *logrus.Entry
	stopChan    chan bool
	stoppedChan chan bool
}

func (s *Switch) run() {
	defer close(s.stoppedChan)
	defer s.log.Info("switch shut down")
	s.log.Info("switch started")
	t := time.NewTicker(100 * time.Millisecond)
	defer t.Stop()
	var oldVal rpi.Value
	for {
		select {
		case <-t.C:
			v, err := s.pin.Read()
			if err != nil {
				s.log.Error(err.Error())
				continue
			}
			if oldVal == rpi.LOW && v == rpi.HIGH {
				s.log.Info("toggling lamp")
				s.relay.Toggle()
			}
			oldVal = v
		case <-s.stopChan:
			return
		}
	}
}

func NewSwitch(r *Relay) (*Switch, error) {
	p, err := rpi.OpenPin(27, rpi.IN)
	if err != nil {
		return nil, err
	}
	s := &Switch{
		relay:       r,
		pin:         p,
		log:         logrus.WithField("context", "switch"),
		stopChan:    make(chan bool),
		stoppedChan: make(chan bool),
	}
	go s.run()
	return s, nil
}

func (s *Switch) Close() {
	close(s.stopChan)
	<-s.stoppedChan
	s.pin.Close()
}
