package api

import "fmt"

type Pin struct {
	Pin  int `json:"pin"`
	Duty int `json:"duty"`
}

// Device handles a picow device
type Device struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	Data []Pin  `json:"data"`

	eventHandler EventHandler
	command      Command

	Offline bool `json:"offline"`
}

func NewDevice(host string, port int) *Device {
	return &Device{
		Host: host,
		Port: port,
	}
}

func (d *Device) GetAddr() string {
	return fmt.Sprintf("%s:%d", d.Host, d.Port)
}

func (d *Device) GetEventHandler() EventHandler {
	return d.eventHandler
}

func (d *Device) SetEventHandler(eventHandler EventHandler) {
	d.eventHandler = eventHandler
}

// UpdateDevice will send the current duty and pin to the picow device
func (d *Device) Update() (err error) {
	var data []int

	// Set pins
	data = []int{-1, -1, -1, -1}
	for i, pin := range d.Data {
		if i > 3 {
			break // NOTE: ignore everything after index 3
		}

		data[i] = pin.Pin
	}

	_, err = d.command.Run(d, d.command.SetPins(data[0], data[1], data[2], data[3]), false)
	if err != nil {
		return d.handleError(err)
	}

	// Set duty
	data = []int{0, 0, 0, 0}
	for i, pin := range d.Data {
		data[i] = pin.Duty
	}

	_, err = d.command.Run(d, d.command.SetColor(data[0], data[1], data[2], data[3]), false)
	if err != nil {
		return d.handleError(err)
	}

	return nil
}

func (d *Device) SetColor(r, g, b, w int) error {
	_, err := d.command.Run(d, d.command.SetColor(r, g, b, w), false)
	if err != nil {
		return d.handleError(err)
	}
	d.SetDataColor(r, g, b, w)

	return nil
}

func (d *Device) SetDataColor(r, g, b, w int) {
	for i, c := range []int{r, g, b, w} {
		if i+1 <= len(d.Data) {
			d.Data[i].Duty = c
			continue
		}
		break
	}

	if d.eventHandler != nil {
		d.eventHandler.DispatchWithData(EventColorChanged, d)
	}
}

func (d *Device) handleError(err error) error {
	if err == nil {
		if d.Offline {
			d.Offline = false
			if d.eventHandler != nil {
				d.eventHandler.DispatchWithData(EventDeviceOnline, d)
			}
		}

		return err
	}

	switch err.(type) {
	case *DialError:
		if !d.Offline {
			d.Offline = true
			if d.eventHandler != nil {
				d.eventHandler.DispatchWithData(EventDeviceOffline, d)
			}
		}
		if d.eventHandler != nil {
			d.eventHandler.DispatchWithData(EventDeviceError, err.Error())
		}
	default:
		if d.Offline {
			d.Offline = false
			if d.eventHandler != nil {
				d.eventHandler.DispatchWithData(EventDeviceOnline, d)
			}
		}
		if d.eventHandler != nil {
			d.eventHandler.DispatchWithData(EventDeviceError, err.Error())
		}
	}

	return err
}
