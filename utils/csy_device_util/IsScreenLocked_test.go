package csy_device_util

import (
	"fmt"
	"testing"
)

func TestIsScreenLocked(t *testing.T) {
	locked, err := IsScreenLocked()
	fmt.Println("TestIsScreenLocked", locked, err)
	fmt.Println("\r\n1")
}
