package api

import (
	"fmt"
	"sync"
)

var (
	// Handler Events

	EventDevicesUpdate = "devices update"
	EventDeviceError   = "device error"

	// PicoW Commands

	TCPCommandGetColor = "rgbw color get;"
	TCPCommandSetColor = "rgbw color set %d %d %d %d;"

	TCPCommandGetPins = "rgbw gp get;"
	TCPCommandSetPins = "rgbw gp set %d %d %d %d;"
)

// EventHandler interface to use
// Events in use from the Handler and Device structs:
//   - `DevicesUpdateEvent` is dispatched if the handler devices private field was updated
type EventHandler interface {
	Dispatch(eventName string)
	DispatchWithMessage(eventName, message string)
}

// Device handles a picow device
type Device struct {
	eventHandler EventHandler
}

func NewDevice() *Device {
	return &Device{}
}

// Update sets duty and pins on the device, if missing, get it from the device
func (d *Device) Sync() {
	// TODO: Device update...
}

// Handler handles all `picow-rgbw-micropython` devices
type Handler struct {
	eventHandler EventHandler
	autoUpdate   bool
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

func (h *Handler) SetAutoUpdate(state bool) {
	h.autoUpdate = state
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
		h.eventHandler.Dispatch(EventDevicesUpdate)
	}

	go h.InitializeDevices()
}

func (h *Handler) InitializeDevices() {
	var wg sync.WaitGroup
	for _, device := range h.devices {
		wg.Add(1)
		go func(device *Device, wg *sync.WaitGroup) {
			defer wg.Done()
			device.Sync()
		}(device, &wg)
	}
}

// Command for picow control
type Command struct {
}

func NewCommand() *Command {
	return &Command{}
}

func (c *Command) Run(device *Device, command string, doRead bool) {
	// TODO: ...
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
