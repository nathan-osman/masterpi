package masterpi

import (
	"fmt"
	"time"
)

type Updater struct {
	display     *Display
	monitor     *Monitor
	location    *time.Location
	stopChan    chan bool
	stoppedChan chan bool
}

func (u *Updater) refresh(t time.Time) {
	t = t.In(u.location)
	u.display.Clear()
	u.display.DrawText(
		fmt.Sprintf("%02d:%02d:%02d", t.Hour(), t.Minute(), t.Second()),
		FontRegular,
		0,
		32,
		32,
	)
	u.display.DrawText(
		fmt.Sprintf("%.1f° C", u.monitor.Value(SensorOutdoor)),
		FontThin,
		8,
		64,
		16,
	)
	u.display.DrawText(
		fmt.Sprintf("%.1f° C", u.monitor.Value(SensorIndoor)),
		FontThin,
		72,
		64,
		16,
	)
	u.display.Flip()
}

func (u *Updater) run() {
	defer close(u.stoppedChan)
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		select {
		case n := <-t.C:
			u.refresh(n)
		case <-u.stopChan:
			return
		}
	}
}

func NewUpdater(m *Monitor) (*Updater, error) {
	d, err := NewDisplay()
	if err != nil {
		return nil, err
	}
	l, err := time.LoadLocation("America/Vancouver")
	if err != nil {
		return nil, err
	}
	u := &Updater{
		display:     d,
		monitor:     m,
		location:    l,
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
