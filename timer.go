package masterpi

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// TODO: permanent storage for this information

type Timer struct {
	mutex          sync.Mutex
	relay          *Relay
	location       *time.Location
	timeOnEntries  []string
	timeOffEntries []string
	log            *logrus.Entry
	triggerChan    chan bool
	stopChan       chan bool
	stoppedChan    chan bool
}

func (t *Timer) parseTime(timeStr string, now time.Time) (time.Time, error) {
	var parsedTime time.Time
	v, err := time.Parse("15:04", timeStr)
	if err != nil {
		return time.Time{}, err
	}
	for {
		parsedTime = time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			v.Hour(),
			v.Minute(),
			0,
			0,
			t.location,
		)
		if parsedTime.Before(now) {
			now.Add(24 * time.Hour)
		} else {
			return parsedTime, nil
		}
	}
}

func (t *Timer) findSoonest(timeStrs []string, now time.Time) time.Time {
	var nextTimeOn time.Time
	for _, v := range timeStrs {
		v, err := t.parseTime(v, now)
		if err != nil {
			t.log.Error(err.Error())
			continue
		}
		if v.Before(nextTimeOn) {
			nextTimeOn = v
		}
	}
	return nextTimeOn
}

func (t *Timer) run() {
	defer close(t.stoppedChan)
	defer t.log.Info("timer shut down")
	t.log.Info("timer started")
	for {
		var (
			now         = time.Now()
			nextTimeOn  = t.findSoonest(t.timeOnEntries, now)
			nextTimeOff = t.findSoonest(t.timeOffEntries, now)
			turnOnChan  <-chan time.Time
			turnOffChan <-chan time.Time
		)
		if !nextTimeOn.IsZero() {
			if nextTimeOn.After(now) {
				t.log.Printf("next on at %s", nextTimeOn.String())
				turnOnChan = time.After(nextTimeOn.Sub(now))
			}
		}
		if !nextTimeOff.IsZero() {
			if nextTimeOff.After(now) {
				t.log.Printf("next off at %s", nextTimeOff.String())
				turnOffChan = time.After(nextTimeOff.Sub(now))
			}
		}
		select {
		case <-turnOnChan:
			t.log.Print("turning on lamp")
			t.relay.SetOn(true)
		case <-turnOffChan:
			t.log.Print("turning off lamp")
			t.relay.SetOn(false)
		case <-t.triggerChan:
		case <-t.stopChan:
			return
		}
	}
}

func NewTimer(r *Relay) (*Timer, error) {
	l, err := time.LoadLocation("America/Vancouver")
	if err != nil {
		return nil, err
	}
	t := &Timer{
		relay:          r,
		location:       l,
		timeOnEntries:  []string{},
		timeOffEntries: []string{},
		log:            logrus.WithField("context", "timer"),
		triggerChan:    make(chan bool, 1),
		stopChan:       make(chan bool),
		stoppedChan:    make(chan bool),
	}
	go t.run()
	return t, nil
}

func (t *Timer) GetTimes() ([]string, []string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.timeOnEntries, t.timeOffEntries
}

func (t *Timer) SetTimes(timeOnEntries, timeOffEntries []string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.timeOnEntries = timeOnEntries
	t.timeOffEntries = timeOffEntries
	select {
	case t.triggerChan <- true:
	default:
	}
}

func (t *Timer) Close() {
	close(t.stopChan)
	<-t.stoppedChan
}
