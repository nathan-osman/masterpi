package masterpi

import (
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/sirupsen/logrus"
)

type Uploader struct {
	monitor     *Monitor
	client      client.Client
	log         *logrus.Entry
	stopChan    chan bool
	stoppedChan chan bool
}

func (u *Uploader) uploadValue(name, location string) error {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "thsat",
		Precision: "s",
	})
	if err != nil {
		return err
	}
	p, err := client.NewPoint(
		"temperature",
		map[string]string{
			"location": location,
		},
		map[string]interface{}{
			"value": u.monitor.Value(name),
		},
	)
	if err != nil {
		return err
	}
	bp.AddPoint(p)
	return u.client.Write(bp)
}

func (u *Uploader) uploadValues() {
	for k, v := range map[string]string{
		SensorOutdoor: "balcony",
		SensorIndoor:  "bedroon",
	} {
		if err := u.uploadValue(k, v); err != nil {
			u.log.Error(err.Error())
		}
	}
}

func (u *Uploader) run() {
	defer u.log.Info("uploader shut down")
	defer close(u.stoppedChan)
	t := time.NewTicker(5 * time.Minute)
	defer t.Stop()
	u.log.Info("uploader started")
	for {
		select {
		case <-t.C:
			u.uploadValues()
		case <-u.stopChan:
			return
		}
	}
}

func NewUploader(addr, username, password string, m *Monitor) (*Uploader, error) {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     addr,
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, err
	}
	u := &Uploader{
		monitor:     m,
		client:      c,
		log:         logrus.WithField("context", "monitor"),
		stopChan:    make(chan bool),
		stoppedChan: make(chan bool),
	}
	go u.run()
	return u, nil
}

func (u *Uploader) Close() {
	close(u.stopChan)
	<-u.stoppedChan
	u.client.Close()
}
