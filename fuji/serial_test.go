package fuji

import "testing"

func TestXor(t *testing.T) {
	numbers := []byte{0x00, 0x09, 0x00, 0x00, 0x03}
	expected := byte(0x0a)
	result := xor(numbers)
	if result != expected {
		t.Errorf("xor(%v) = 0x%02X; want 0x%02X", numbers, result, expected)
	}
}
