package api

import (
	"bytes"
	"fmt"
	"net"
)

var (
	/*
	 * PicoW Commands
	 */

	TCPCommandGetColor = "rgbw color get;"
	TCPCommandSetColor = "rgbw color set %d %d %d %d;"

	TCPCommandGetPins = "rgbw gp get;"
	TCPCommandSetPins = "rgbw gp set %d %d %d %d;"
)

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
