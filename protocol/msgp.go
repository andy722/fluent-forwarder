package protocol

import "github.com/tinylib/msgp/msgp"

// ReadRaw transforms raw message pack bytes to string
func ReadRaw(rawMsg msgp.Raw) string {
	raw := rawMsg

	switch raw[0] {
	case 0xd9:
		raw = raw[2:]
	case 0xda:
		raw = raw[3:]
	case 0xdb:
		raw = raw[5:]
	default:
		// fixstr?
		raw = raw[1:]
	}

	return msgp.UnsafeString(raw)
}
