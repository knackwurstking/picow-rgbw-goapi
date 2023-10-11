package api

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
