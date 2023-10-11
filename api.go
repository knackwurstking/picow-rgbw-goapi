package api

import (
	"bytes"
	"fmt"
	"net"
)

var (
	/*
	 * Handler Events
	 */

	// EventDevicesUpdate contains no message or data
	EventDevicesUpdated = "devices updated"
	// EventDeviceError event contains a message
	EventDeviceError = "device error"
	// EventDeviceOnline event will contain `*Device` data
	EventDeviceOnline = "device online"
	// EventDeviceOffline event will contain `*Device` data
	EventDeviceOffline = "device offline"
	// EventColorChanged event will contain `*Device` data
	EventColorChanged = "color changed"

	/*
	 * PicoW Commands
	 */

	TCPCommandGetColor = "rgbw color get;"
	TCPCommandSetColor = "rgbw color set %d %d %d %d;"

	TCPCommandGetPins = "rgbw gp get;"
	TCPCommandSetPins = "rgbw gp set %d %d %d %d;"
)

/*
 * Interfaces
 */

// EventHandler interface to use
// Events in use from the Handler and Device structs:
//   - `DevicesUpdateEvent` is dispatched if the handler devices private field was updated
type EventHandler interface {
	Dispatch(eventName string)
	DispatchWithData(eventName string, data any)
	AddListener(eventName string, listener func(data any)) error
	RemoveListener(eventName string)
}

/*
 * Device
 */

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

/*
 * Handler
 */

// Handler handles all `picow-rgbw-micropython` devices
type Handler struct {
	eventHandler EventHandler
	devices      []*Device
}

func NewHandler(eventHandler EventHandler, devices ...*Device) *Handler {
	return &Handler{
		eventHandler: eventHandler,
		devices:      devices,
	}
}

func (h *Handler) GetEventHandler() EventHandler {
	return h.eventHandler
}

func (h *Handler) SetEventHandler(eventHandler EventHandler) {
	h.eventHandler = eventHandler
	for _, device := range h.devices {
		device.eventHandler = h.eventHandler
	}
}

func (h *Handler) GetDevicePerAddr(addr string) *Device {
	for _, device := range h.devices {
		if device.GetAddr() == addr {
			return device
		}
	}

	return nil
}

func (h *Handler) AddDevice(d *Device) {
	// Check if device exists (host and port should be unique)
	for i, device := range h.devices {
		if device.GetAddr() == d.GetAddr() {
			// Replace device
			d.eventHandler = h.eventHandler
			h.devices[i] = d
			return
		}
	}

	d.eventHandler = h.eventHandler
	h.devices = append(h.devices, d)
}

func (h *Handler) GetDevices() []*Device {
	return h.devices
}

func (h *Handler) SetDevices(devices ...*Device) {
	for _, device := range devices {
		device.eventHandler = h.eventHandler
	}

	h.devices = devices

	if h.eventHandler != nil {
		h.eventHandler.Dispatch(EventDevicesUpdated)
	}
}

/*
 * Command
 */

// Command for picow control
type Command struct {
}

func NewCommand() *Command {
	return &Command{}
}

func (c *Command) Run(device *Device, command string, doRead bool) (data []byte, err error) {
	conn, err := net.Dial("tcp", device.GetAddr())
	if err != nil {
		return data, NewDialError(device.Host, device.Port, err)
	}
	defer conn.Close()

	n, err := conn.Write([]byte(command))
	if err != nil {
		return data, NewConnectionError(device.Host, device.Port, err)
	} else if n == 0 {
		return data, NewConnectionError(device.Host, device.Port, fmt.Errorf("no data written to %s", device.GetAddr()))
	}

	// Read the (optional) response
	if !doRead {
		return data, err
	}

	return c.readUntilEnd(conn, device)
}

func (c *Command) GetPins() string {
	return TCPCommandGetPins
}

func (c *Command) SetPins(r, g, b, w int) string {
	return fmt.Sprintf(TCPCommandSetPins, r, g, b, w)
}

func (c *Command) GetColor() string {
	return TCPCommandGetColor
}

func (c *Command) SetColor(r, g, b, w int) string {
	return fmt.Sprintf(TCPCommandSetColor, r, g, b, w)
}

func (c *Command) readUntilEnd(conn net.Conn, device *Device) (data []byte, err error) {
	for {
		chunk := make([]byte, 1024)
		n, err := conn.Read(chunk)
		if err != nil {
			return data, NewConnectionError(device.Host, device.Port, err)
		} else if n == 0 {
			break
		}
		data = append(data, chunk...)
	}

	return bytes.Trim(data, " \r\n"), nil
}
