package api

import "fmt"

type DialError struct {
	Host string
	Port int

	err error
}

func NewDialError(host string, port int, err error) *DialError {
	return &DialError{host, port, err}
}

func (d *DialError) Error() string {
	return fmt.Sprintf("%s: %s", d.GetAddr(), d.err.Error())
}

func (d *DialError) GetAddr() string {
	return fmt.Sprintf("%s:%d", d.Host, d.Port)
}

type ConnectionError struct {
	Host string
	Port int

	err error
}

func NewConnectionError(host string, port int, err error) *ConnectionError {
	return &ConnectionError{host, port, err}
}

func (c *ConnectionError) Error() string {
	return fmt.Sprintf("%s: %s", c.GetAddr(), c.err.Error())
}

func (c *ConnectionError) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
