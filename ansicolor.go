package main

import "fmt"

// Used for logging messages & errors
type ANSIColor int

const (
	Black ANSIColor = iota
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
	Reset
)

func GetEscape(c ANSIColor) string {
	switch c {
	case Black:
		return "\033[0;30m"
	case Red:
		return "\033[0;31m"
	case Green:
		return "\033[0;32m"
	case Yellow:
		return "\033[0;33m"
	case Blue:
		return "\033[0;34m"
	case Magenta:
		return "\033[0;35m"
	case Cyan:
		return "\033[0;36m"
	case White:
		return "\033[0;37m"
	default:
		return "\033[0m"
	}
}

func ColorError(message string, color ANSIColor) error {
	if UseANSIColor {
		return fmt.Errorf(GetEscape(color) + message + GetEscape(Reset))
	}
	return fmt.Errorf(message)
}
