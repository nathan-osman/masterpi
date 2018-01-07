package masterpi

import (
	"sync"

	"github.com/nathan-osman/go-rpigpio"
)

type Relay struct {
	mutex sync.Mutex
	pin   *rpi.Pin
	on    bool
}

func initialValue() (rpi.Value, error) {
	p, err := rpi.OpenPin(17, rpi.IN)
	if err != nil {
		return 0, err
	}
	defer p.Close()
	v, err := p.Read()
	if err != nil {
		return 0, err
	}
	return v, nil
}

func NewRelay() (*Relay, error) {
	v, err := initialValue()
	if err != nil {
		return nil, err
	}
	p, err := rpi.OpenPin(17, rpi.OUT)
	if err != nil {
		return nil, err
	}
	return &Relay{
		pin: p,
		on:  v == rpi.HIGH,
	}, nil
}

func (r *Relay) IsOn() bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.on
}

func (r *Relay) Toggle() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.on {
		r.pin.Write(rpi.LOW)
	} else {
		r.pin.Write(rpi.HIGH)
	}
	r.on = !r.on
}

func (r *Relay) Close() {
	r.pin.Close()
}
