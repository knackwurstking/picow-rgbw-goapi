package api

type EventHandler interface {
	Handler(eventName string)
}

type Device struct {
}

func NewDevice() *Device {
	return &Device{}
}

type Handler struct {
	EventHandler EventHandler

	devices []*Device
}

func NewHandler(eventHandler EventHandler, devices ...*Device) *Handler {
	return &Handler{
		EventHandler: eventHandler,
		devices:      devices,
	}
}

func (h *Handler) GetDevices() []*Device {
	return h.devices
}

func (h *Handler) SetDevices(devices ...*Device) {
	// TODO: passing the EventHandler pointer to each device first
	h.devices = devices
}
