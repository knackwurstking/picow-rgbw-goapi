package api

import "sync"

var (
	EventDevicesUpdate = "devices update"
	EventDeviceError   = "device error"
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
func (d *Device) Update() {
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
			device.Update()
		}(device, &wg)
	}
}
