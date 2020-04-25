package util

import (
	"fmt"
	"strings"
)

func SplitAddress(addr string) (network string, address string, err error) {
	parts := strings.Split(addr, "://")
	if len(parts) != 2 {
		err = fmt.Errorf("%v: invalid address", addr)
		return
	}

	network = parts[0]
	address = parts[1]

	if network != "tcp" && network != "unix" {
		err = fmt.Errorf("%v: invalid address", addr)
	}

	return

}
