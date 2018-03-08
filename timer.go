package masterpi

import (
	"sync"

	"github.com/sirupsen/logrus"
)

// TODO: permanent storage for this information

type Timer struct {
	mutex          sync.Mutex
	relay          *Relay
	timeOnEntries  []string
	timeOffEntries []string
	log            *logrus.Entry
	triggerChan    chan bool
	stopChan       chan bool
	stoppedChan    chan bool
}

func (t *Timer) run() {
	defer close(t.stoppedChan)
	for {
		//...

		select {
		case <-t.triggerChan:
		case <-t.stopChan:
			return
		}
	}
}

func NewTimer(r *Relay) *Timer {
	t := &Timer{
		relay:       r,
		log:         logrus.WithField("context", "timer"),
		triggerChan: make(chan bool, 1),
		stopChan:    make(chan bool),
		stoppedChan: make(chan bool),
	}
	go t.run()
	return t
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
