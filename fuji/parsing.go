package fuji

import (
	"encoding/binary"
	"fmt"
)

type Parser struct {
	verboseLvl int
}

// GetParser instantiate a Parser
func GetParser(verboseLvl int) Parser {
	return Parser{
		verboseLvl: verboseLvl,
	}
}

// Parse the response data from GET_CAMERA_VERSION command
func (p Parser) ParseCameraVersionPacket(data []byte) (string, error) {
	if p.verboseLvl > 0 {
		fmt.Printf("Raw response to parse as camera version packet: % X\n", data)
		fmt.Println("Format should be 0x00 0x05 LENGTH 0x00 DATA")
	}
	if len(data) < 4 || data[0] != 0x00 || data[1] != 0x05 || data[3] != 0x00 {
		return "", fmt.Errorf("Malformated message")
	}
	length := int(data[2])
	if length == 0 {
		return "", nil
	}
	model := string(data[4 : 4+length])
	if p.verboseLvl > 0 {
		fmt.Printf("Data length: %d\n", length)
		fmt.Printf("Camera model: %s\n", model)
	}
	return model, nil
}

// Parse the response data from COUNT_PICTURES command
func (p Parser) ParseCountPicturesPacket(data []byte) (int, error) {
	if p.verboseLvl > 0 {
		fmt.Printf("Raw response to parse as count pictures packet: % X\n", data)
		fmt.Println("Format should be 0x00 0x06 0x02 0x00 COUNT COUNT")
	}
	if len(data) < 4 || data[0] != 0x00 || data[1] != 0x06 || data[2] != 0x02 || data[3] != 0x00 {
		return 0, fmt.Errorf("Malformated message")
	}
	count := binary.LittleEndian.Uint16(data[4:6])
	return int(count), nil
}

// packetParser defines an interface for parsing different types of packets
// a packetParser can modify the raw data received from the camera when reading from serial
type packetParser interface {
	Parse(data []byte) ([]byte, error)
}

// picturePacketParser is a packet parser for picture data packets
type picturePacketParser struct {
	verboseLvl int
}

func (p picturePacketParser) Parse(data []byte) ([]byte, error) {
	if p.verboseLvl > 1 {
		fmt.Printf("data to parse: %x\n", data)
		fmt.Println("picture parser verify package and remove control bytes")
	}
	if len(data) < 4 || data[0] != 0x00 || data[1] != 0x03 {
		return nil, fmt.Errorf("Malformated picture packet")
	}
	size := binary.LittleEndian.Uint16(data[2:4])
	if len(data) < int(4+size) {
		return nil, fmt.Errorf("Incomplete picture packet")
	}
	return data[4:], nil
}

// defaultPacketParser is a packet parser that does not modify the data
type defaultPacketParser struct {
	verboseLvl int
}

func (p defaultPacketParser) Parse(data []byte) ([]byte, error) {
	if p.verboseLvl > 1 {
		fmt.Printf("data to parse: %x\n", data)
		fmt.Println("default parser, no data modified")
	}
	return data, nil
}
