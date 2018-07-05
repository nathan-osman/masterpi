package masterpi

import (
	"encoding/json"
	"os"
	"path"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type timerConfig struct {
	fileName       string
	TimeOnEntries  []string `json:"time_on_entries"`
	TimeOffEntries []string `json:"time_off_entries"`
}

func (t *timerConfig) Load() error {
	f, err := os.Open(t.fileName)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil {
		defer f.Close()
		if err := json.NewDecoder(f).Decode(t); err != nil {
			return err
		}
	}
	return nil
}

func (t *timerConfig) Save() error {
	f, err := os.Create(t.filename)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(t); err != nil {
		return err
	}
	return nil
}

type Timer struct {
	mutex       sync.Mutex
	relay       *Relay
	location    *time.Location
	config      *timerConfig
	log         *logrus.Entry
	triggerChan chan bool
	stopChan    chan bool
	stoppedChan chan bool
}

func (t *Timer) parseTime(timeStr string, now time.Time) (time.Time, error) {
	v, err := time.Parse("15:04", timeStr)
	if err != nil {
		return time.Time{}, err
	}
	curDay := now
	for {
		parsedTime := time.Date(
			curDay.Year(),
			curDay.Month(),
			curDay.Day(),
			v.Hour(),
			v.Minute(),
			0,
			0,
			t.location,
		)
		if parsedTime.Before(now) {
			curDay = curDay.Add(24 * time.Hour)
		} else {
			return parsedTime, nil
		}
	}
}

func (t *Timer) findSoonest(timeStrs []string, now time.Time) time.Time {
	var soonestTime time.Time
	for _, timeStr := range timeStrs {
		v, err := t.parseTime(timeStr, now)
		if err != nil {
			t.log.Error(err.Error())
			continue
		}
		if soonestTime.IsZero() || v.Before(soonestTime) {
			soonestTime = v
		}
	}
	return soonestTime
}

func (t *Timer) run() {
	defer close(t.stoppedChan)
	defer t.log.Info("timer shut down")
	t.log.Info("timer started")
	for {
		var (
			now                           = time.Now()
			timeOnEntries, timeOffEntries = t.GetTimes()
			nextTimeOn                    = t.findSoonest(timeOnEntries, now)
			nextTimeOff                   = t.findSoonest(timeOffEntries, now)
			turnOnChan                    <-chan time.Time
			turnOffChan                   <-chan time.Time
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

func NewTimer(storageDir string, r *Relay) (*Timer, error) {
	l, err := time.LoadLocation("America/Vancouver")
	if err != nil {
		return nil, err
	}
	config := &timerConfig{
		fileName: path.Join(storageDir, "config.json"),
	}
	if err := config.Load(); err != nil {
		return nil, err
	}
	t := &Timer{
		relay:       r,
		location:    l,
		config:      config,
		log:         logrus.WithField("context", "timer"),
		triggerChan: make(chan bool, 1),
		stopChan:    make(chan bool),
		stoppedChan: make(chan bool),
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
	if err := t.config.Save(); err != nil {
		t.log.Error(err.Error())
	}
	select {
	case t.triggerChan <- true:
	default:
	}
}

func (t *Timer) Close() {
	close(t.stopChan)
	<-t.stoppedChan
}
