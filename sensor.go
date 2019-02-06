package masterpi

import (
	"time"

	"github.com/nathan-osman/go-rpigpio"
	"github.com/sirupsen/logrus"
)

type Sensor struct {
	relay       *Relay
	pin         *rpi.Pin
	log         *logrus.Entry
	stopChan    chan bool
	stoppedChan chan bool
}

func (s *Sensor) run() {
	defer close(s.stoppedChan)
	defer s.log.Info("sensor shut down")
	s.log.Info("sensor started")
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
			if v == rpi.HIGH {
				s.relay.SetOn(true)
			}
			oldVal = v
		case <-s.stopChan:
			return
		}
	}
}

func NewSensor(r *Relay) (*Sensor, error) {
	p, err := rpi.OpenPin(25, rpi.IN)
	if err != nil {
		return nil, err
	}
	s := &Sensor{
		relay:       r,
		pin:         p,
		log:         logrus.WithField("context", "sensor"),
		stopChan:    make(chan bool),
		stoppedChan: make(chan bool),
	}
	go s.run()
	return s
}

func (s *Sensor) Close() {
	close(s.stopChan)
	<-s.stoppedChan
	s.pin.Close()
}
