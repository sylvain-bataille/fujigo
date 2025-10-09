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

type Message struct {
	data   []byte
	pretty string
}

type Parser struct {
	verbose bool
}

var ENQUIRY_MSG = Message{data: []byte{ENQ}, pretty: "ENQ"}
var BEGIN_MSG = Message{data: []byte{DLE, STX}, pretty: "DLE STX"}
var GET_CAMERA_VERSION_MSG = Message{data: []byte{0x00, 0x09, 0x00, 0x00}, pretty: "GET_CAMERA_VERSION"}
var COUNT_PICTURES_MSG = Message{data: []byte{0x00, 0x0B, 0x00, 0x00}, pretty: "COUNT_PICTURES"}
var END_MSG = Message{data: []byte{DLE, ETX}, pretty: "DLE ETX"}
var ACK_MSG = Message{data: []byte{ACK}, pretty: "ACK"}
var END_OF_COMMUNICATION_MSG = Message{data: []byte{EOT}, pretty: "EOT"}

// GetParser instantiate a Parser
func GetParser(verbose bool) Parser {
	return Parser{
		verbose: verbose,
	}
}

// BuildEndTextMessageWithChecksum returns a 3 bytes message formated as : DLE ETX CHECKSUM
// Checksum is calculated from the data previously sent and ETX byte
func BuildEndTextMessageWithChecksum(data []byte) Message {
	msgData := append(END_MSG.data, xor(append(data, ETX)))
	return Message{data: msgData, pretty: "DLE ETX CHECKSUM"}
}

// BuildDownloadPictureMessage returns a message to download a picture by its number
// Picture number must be between 1 and 65535
func BuildDownloadPictureMessage(pictureNumber int) Message {
	if pictureNumber < 1 || pictureNumber > 65535 {
		panic("Picture number must be between 1 and 65535")
	}
	msgData := []byte{0x00, 0x02, 0x02, 0x00}
	countBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(countBytes, uint16(pictureNumber))
	msgData = append(msgData, countBytes...)
	return Message{data: msgData, pretty: fmt.Sprintf("DOWNLOAD_PICTURE %d", pictureNumber)}
}

// BuildSetBaudrateMessage returns a message to set the baudrate of the camera
func BuildSetBaudrateMessage(baudRate BaudRate) Message {
	msgData := []byte{0x01, 0x07, 0x01, 0x00, byte(baudRate)}
	return Message{data: msgData, pretty: fmt.Sprintf("SET_BAUDRATE %d", baudRate)}
}

// GetBaudRateAsInt converts a BaudRate constant to its integer value
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
