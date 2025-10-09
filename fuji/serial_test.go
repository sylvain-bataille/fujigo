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

func TestParseCameraVersionPacket(t *testing.T) {
	parser := GetParser(0)
	data := []byte{0x00, 0x05, 0x08, 0x00, 0x46, 0x75, 0x6A, 0x69, 0x20, 0x58, 0x70, 0x72}
	version, err := parser.ParseCameraVersionPacket(data)
	if err != nil {
		t.Errorf("ParseCameraVersionPacket(%v) returned error: %v", data, err)
	}
	expected := "Fuji Xpr"
	if version != expected {
		t.Errorf("ParseCameraVersionPacket(%v) = %s; want %s", data, version, expected)
	}
}

func TestParseCountPicturesPacket(t *testing.T) {
	parser := GetParser(0)
	data := []byte{0x00, 0x06, 0x02, 0x00, 0x03, 0x00}
	count, err := parser.ParseCountPicturesPacket(data)
	if err != nil {
		t.Errorf("ParseCountPicturesPacket(%v) returned error: %v", data, err)
	}
	expected := 3
	if count != expected {
		t.Errorf("ParseCountPicturesPacket(%v) = %d; want %d", data, count, expected)
	}
}
