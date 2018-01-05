package masterpi

import (
	"fmt"
	"time"
)

type Updater struct {
	display     *Display
	stopChan    chan bool
	stoppedChan chan bool
}

func (u *Updater) run() {
	defer close(u.stoppedChan)

	// Create a timer for updating the clock
	t := time.NewTicker(time.Second)
	defer t.Stop()

	for {
		select {
		case n := <-t.C:
			u.display.Clear()
			u.display.DrawText(
				fmt.Sprintf("%02d:%02d:%02d", n.Hour(), n.Minute(), n.Second()),
				0,
				32,
				32,
			)
			u.display.Flip()
		case <-u.stopChan:
			return
		}
	}
}

func NewUpdater() (*Updater, error) {
	d, err := NewDisplay()
	if err != nil {
		return nil, err
	}
	u := &Updater{
		display:     d,
		stopChan:    make(chan bool),
		stoppedChan: make(chan bool),
	}
	go u.run()
	return u, nil
}

func (u *Updater) Close() {
	close(u.stopChan)
	<-u.stoppedChan
	u.display.Close()
}
