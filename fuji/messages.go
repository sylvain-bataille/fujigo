package fuji

import (
	"encoding/binary"
	"fmt"
)

const ENQ = 0x05 // Enquiry, used to initiate communication
const ACK = 0x06 // Acknowledge, used to acknowledge receipt of a message
const DLE = 0x10 // Data Link Escape, used to indicate special control characters
const STX = 0x02 // Start of Text, initiate the begining of a message
const ETX = 0x03 // End of Text, indicate the end of a message
const EOT = 0x04 // End of Transmission, used to terminate communication
const ETB = 0x17 // End of Transmission Block, meaning more data will follow

type BaudRate int8

const (
	BAUD_9600   BaudRate = 0x00
	BAUD_12000  BaudRate = 0x01
	BAUD_14400  BaudRate = 0x02
	BAUD_16800  BaudRate = 0x03
	BAUD_19200  BaudRate = 0x04
	BAUD_28800  BaudRate = 0x05
	BAUD_38400  BaudRate = 0x06
	BAUD_57600  BaudRate = 0x07
	BAUD_115200 BaudRate = 0x08
)

type message struct {
	data   []byte
	pretty string
}

type Parser struct {
	verbose bool
}

var ENQUIRY_MSG = message{data: []byte{ENQ}, pretty: "ENQ"}
var BEGIN_MSG = message{data: []byte{DLE, STX}, pretty: "DLE STX"}
var GET_CAMERA_VERSION_MSG = message{data: []byte{0x00, 0x09, 0x00, 0x00}, pretty: "GET_CAMERA_VERSION"}
var COUNT_PICTURES_MSG = message{data: []byte{0x00, 0x0B, 0x00, 0x00}, pretty: "COUNT_PICTURES"}
var END_MSG = message{data: []byte{DLE, ETX}, pretty: "DLE ETX"}
var ACK_MSG = message{data: []byte{ACK}, pretty: "ACK"}
var END_OF_COMMUNICATION_MSG = message{data: []byte{EOT}, pretty: "EOT"}

// GetParser instantiate a Parser
func GetParser(verbose bool) Parser {
	return Parser{
		verbose: verbose,
	}
}

// GetEndTextMessageWithChecksum returns a 3 bytes message formated as : DLE ETX CHECKSUM
// Checksum is calculated from the data previously sent and ETX byte
func GetEndTextMessageWithChecksum(data []byte) message {
	msgData := append(END_MSG.data, xor(append(data, ETX)))
	return message{data: msgData, pretty: "DLE ETX CHECKSUM"}
}

// GetDownloadPictureMessage returns a message to download a picture by its number
// Picture number must be between 1 and 65535
func GetDownloadPictureMessage(pictureNumber int) message {
	if pictureNumber < 1 || pictureNumber > 65535 {
		panic("Picture number must be between 1 and 65535")
	}
	msgData := []byte{0x00, 0x02, 0x02, 0x00}
	countBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(countBytes, uint16(pictureNumber))
	msgData = append(msgData, countBytes...)
	return message{data: msgData, pretty: fmt.Sprintf("DOWNLOAD_PICTURE %d", pictureNumber)}
}

// GetSetBaudrateMessage returns a message to set the baudrate of the camera
func GetSetBaudrateMessage(baudRate BaudRate) message {
	msgData := []byte{0x01, 0x07, 0x01, 0x00, byte(baudRate)}
	return message{data: msgData, pretty: fmt.Sprintf("SET_BAUDRATE %d", baudRate)}
}

func GetBaudRateAsInt(baudRate BaudRate) int {
	switch baudRate {
	case BAUD_9600:
		return 9600
	case BAUD_12000:
		return 12000
	case BAUD_14400:
		return 14400
	case BAUD_16800:
		return 16800
	case BAUD_19200:
		return 19200
	case BAUD_28800:
		return 28800
	case BAUD_38400:
		return 38400
	case BAUD_57600:
		return 57600
	case BAUD_115200:
		return 115200
	default:
		return 9600
	}
}

// XOR all the bytes of the data slice
func xor(data []byte) byte {
	var result byte = 0
	for _, b := range data {
		result ^= b
	}
	return result
}

// Parse the response data from GET_CAMERA_VERSION command
func (p Parser) ParseCameraVersionPacket(data []byte) (string, error) {
	if p.verbose {
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
	if p.verbose {
		fmt.Printf("Data length: %d\n", length)
		fmt.Printf("Camera model: %s\n", model)
	}
	return model, nil
}

// Parse the response data from COUNT_PICTURES command
func (p Parser) ParseCountPicturesPacket(data []byte) (int, error) {
	if p.verbose {
		fmt.Printf("Raw response to parse as count pictures packet: % X\n", data)
		fmt.Println("Format should be 0x00 0x06 0x02 0x00 COUNT COUNT")
	}
	if len(data) < 4 || data[0] != 0x00 || data[1] != 0x06 || data[2] != 0x02 || data[3] != 0x00 {
		return 0, fmt.Errorf("Malformated message")
	}
	count := binary.LittleEndian.Uint16(data[4:6])
	return int(count), nil
}
