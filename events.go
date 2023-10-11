package api

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
)

// EventHandler interface to use
// Events in use from the Handler and Device structs:
//   - `DevicesUpdateEvent` is dispatched if the handler devices private field was updated
type EventHandler interface {
	Dispatch(eventName string)
	DispatchWithData(eventName string, data any)
	AddListener(eventName string, listener func(data any)) error
	RemoveListener(eventName string)
}
