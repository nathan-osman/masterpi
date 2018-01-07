package masterpi

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yryz/ds18b20"
)

const (
	SensorOutdoor = "28-0416a4a2abff"
	SensorIndoor  = "28-0416a4a40cff"
)

type Monitor struct {
	mutex       sync.Mutex
	values      map[string]float64
	log         *logrus.Entry
	stopChan    chan bool
	stoppedChan chan bool
}

func (m *Monitor) updateSensors() {
	sensors, err := ds18b20.Sensors()
	if err != nil {
		m.log.Error(err.Error())
		return
	}
	for _, s := range sensors {
		v, err := ds18b20.Temperature(s)
		if err != nil {
			m.log.Error(err.Error())
			continue
		}
		func() {
			m.mutex.Lock()
			defer m.mutex.Unlock()
			m.values[s] = v
		}()
	}
}

func (m *Monitor) run() {
	defer close(m.stoppedChan)
	defer m.log.Info("monitor shut down")
	t := time.NewTicker(time.Minute)
	defer t.Stop()
	m.log.Info("monitor started")
	m.updateSensors()
	for {
		select {
		case <-t.C:
			m.updateSensors()
		case <-m.stopChan:
			return
		}
	}
}

func NewMonitor() (*Monitor, error) {
	var (
		l            = logrus.WithField("context", "monitor")
		sensors, err = ds18b20.Sensors()
	)
	if err != nil {
		return nil, err
	}
	for _, s := range sensors {
		l.Infof("found sensor \"%s\"", s)
	}
	m := &Monitor{
		values:      make(map[string]float64),
		log:         l,
		stopChan:    make(chan bool),
		stoppedChan: make(chan bool),
	}
	go m.run()
	return m, nil
}

func (m *Monitor) Value(name string) float64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.values[name]
}

func (m *Monitor) Close() {
	close(m.stopChan)
	<-m.stoppedChan
}
