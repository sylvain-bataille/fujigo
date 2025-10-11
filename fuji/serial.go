package fuji

import (
	"fmt"
	"log"
	"time"

	"go.bug.st/serial"
)

// SerialClient represents a client for serial communication with the camera
// It holds configuration settings and the serial port instance
// Parameters:
//   - Verbose enables detailed logging
//   - Device is the serial port device (e.g., /dev/ttyUSB0 or COM3)
//   - BaudRate is the communication speed (e.g., 9600, 115200)
//   - Port is the opened serial port instance
//   - DefaultTimeout is the read timeout duration
type SerialClient struct {
	VerboseLvl     int
	Device         string
	BaudRate       int
	Port           serial.Port
	DefaultTimeout time.Duration
}

// NewSerialClient creates a new SerialClient with the specified settings
func NewSerialClient(verboseLvl int, device string, baudRate int) *SerialClient {
	return &SerialClient{VerboseLvl: verboseLvl, Device: device, BaudRate: baudRate, DefaultTimeout: time.Duration(1) * time.Second}
}

// setBaudrateSetting sets the baudrate setting of the SerialClient and pauses for 100ms
func (s *SerialClient) setBaudrateSetting(baudRate int) error {
	if s.VerboseLvl > 0 {
		fmt.Printf("Resetting baudrate to %d...\n", baudRate)
	}
	s.BaudRate = baudRate
	s.pause()
	return nil
}

// pause pauses execution for 100 milliseconds
func (s *SerialClient) pause() {
	time.Sleep(100 * time.Millisecond)
}

// sendBytes sends raw bytes to the serial port
func (s *SerialClient) sendBytes(data []byte) (int, error) {
	n, err := s.Port.Write(data)
	if err != nil {
		return 0, err
	}
	if s.VerboseLvl > 1 {
		fmt.Printf("Sent %v bytes\n", n)
	}
	return n, nil
}

// sendCommand sends a complete command sequence: DLE STX, data message, DLE ETX CHECKSUM
// and waits for ACK
func (s *SerialClient) sendCommand(msg Message) error {
	// Send DLE STX to indicate start of text
	err := s.sendMessage(BEGIN_MSG)
	if err != nil {
		return err
	}
	// Send data message
	err = s.sendMessage(msg)
	if err != nil {
		return err
	}
	// Send DLE ETX and checksum
	err = s.sendMessage(BuildEndTextMessageWithChecksum(msg.data))
	if err != nil {
		return err
	}

	err = s.waitForACK()
	if err != nil {
		return err
	}
	return nil
}

// sendMessage sends a raw message to the serial port and logs it if verbose is enabled
func (s *SerialClient) sendMessage(msg Message) error {
	if s.VerboseLvl > 1 {
		fmt.Printf("Sending message: ")
		fmt.Printf(" (%s)\n", msg.pretty)
	}
	n, err := s.sendBytes(msg.data)
	if err != nil {
		return err
	}
	if n != len(msg.data) {
		return fmt.Errorf("sent %d bytes, expected to send %d bytes", n, len(msg.data))
	}
	if s.VerboseLvl > 1 {
		fmt.Printf("Message sent: %s\n", msg.pretty)
	}
	return nil
}

// waitForACK waits for an ACK byte from the serial port
func (s *SerialClient) waitForACK() error {
	// Wait for ACK
	if s.VerboseLvl > 0 {
		fmt.Println("Waiting for ACK...")
	}
	b, err := readByte(s.Port)
	if err != nil {
		return err
	}
	if b != ACK {
		return fmt.Errorf("expected ACK (0x%02X), got 0x%02X", ACK, b)
	}
	if s.VerboseLvl > 0 {
		fmt.Printf("Received ACK (0x%02X)\n", b)
	}
	return nil
}

// openPort opens the serial port with the specified settings
func (s *SerialClient) openPort() error {
	mode := &serial.Mode{
		BaudRate: s.BaudRate,
		Parity:   serial.EvenParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}
	if s.VerboseLvl > 0 {
		fmt.Printf("Opening port %s with mode %+v\n", s.Device, mode)
	}
	port, err := serial.Open(s.Device, mode)
	if err != nil {
		return err
	}
	s.Port = port
	return nil
}

// openCommunication opens the port, sets read timeout, and flushes input buffer
func (s *SerialClient) openCommunication() error {
	err := s.openPort()
	if err != nil {
		return err
	}
	s.Port.SetReadTimeout(s.DefaultTimeout)
	if s.VerboseLvl > 0 {
		fmt.Println("Flushing input buffer...")
	}
	err = s.Port.ResetInputBuffer()
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

// initiateCommunication with the camera
// Opens the port, sends ENQ, and waits for ACK
func (s *SerialClient) initiateCommunication() error {
	err := s.openCommunication()
	if err != nil {
		return err
	}
	err = s.sendMessage(ENQUIRY_MSG)
	err = s.waitForACK()
	if err != nil {
		return err
	}
	return nil
}

// ListPorts lists available serial ports
func (s *SerialClient) ListPorts() ([]string, error) {
	if s.VerboseLvl > 0 {
		fmt.Println("Getting serial ports...")
	}
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, err
	}
	if s.VerboseLvl > 0 {
		fmt.Printf("Found %d ports\n", len(ports))
	}
	return ports, nil
}

// GetModel return information about camera model
func (s *SerialClient) GetModel() (string, error) {
	if s.VerboseLvl > 0 {
		fmt.Println("Initiating communication...")
	}
	err := s.initiateCommunication()
	if err != nil {
		return "", err
	}
	defer s.Close()

	err = s.sendCommand(GET_CAMERA_VERSION_MSG)
	if err != nil {
		return "", err
	}

	response, err := s.readPacket(true, defaultPacketParser{verboseLvl: s.VerboseLvl})
	if err != nil {
		return "", err
	}

	model, err := GetParser(s.VerboseLvl).ParseCameraVersionPacket(response)
	if err != nil {
		return "", fmt.Errorf("Parsing issue: %w", err)
	}
	return model, nil
}

// CountPictures returns the number of pictures stored on the camera
func (s *SerialClient) CountPictures() (int, error) {
	if s.VerboseLvl > 0 {
		fmt.Println("Initiating communication...")
	}
	err := s.initiateCommunication()
	if err != nil {
		return 0, err
	}
	defer s.Close()

	err = s.sendCommand(COUNT_PICTURES_MSG)
	if err != nil {
		return 0, err
	}

	response, err := s.readPacket(true, defaultPacketParser{verboseLvl: s.VerboseLvl})
	if err != nil {
		return 0, err
	}

	count, err := GetParser(s.VerboseLvl).ParseCountPicturesPacket(response)
	if err != nil {
		return 0, fmt.Errorf("Parsing issue: %w", err)
	}
	return count, nil
}

// DownloadPicture downloads a picture by its number (1 to 65535) and returns the raw JPEG data
func (s *SerialClient) DownloadPicture(pictureNumber int) ([]byte, error) {
	// Set higher baudrate for faster download
	initialBaudRate := s.BaudRate
	err := s.setBaudRate(BAUD_115200)
	if err != nil {
		return nil, err
	}

	if s.VerboseLvl > 0 {
		fmt.Println("Initiating communication...")
	}
	err = s.initiateCommunication()
	if err != nil {
		return nil, err
	}
	defer s.Close()
	defer s.setBaudrateSetting(initialBaudRate)

	downloadMsg := BuildDownloadPictureMessage(pictureNumber)
	if s.VerboseLvl > 0 {
		fmt.Printf("Send download picture command: %s\n", downloadMsg.pretty)
	}
	err = s.sendCommand(downloadMsg)
	if err != nil {
		return nil, err
	}

	s.Port.SetReadTimeout(time.Duration(30) * time.Second)
	pictureData, err := s.readPacket(true, picturePacketParser{verboseLvl: s.VerboseLvl})
	s.Port.SetReadTimeout(s.DefaultTimeout)
	if err != nil {
		return nil, err
	}

	if s.VerboseLvl > 0 {
		fmt.Printf("Downloaded picture %d, size: %d bytes\n", pictureNumber, len(pictureData))
	}
	// Restore initial baudrate
	return pictureData, nil
}

func (s *SerialClient) DeletePicture(pictureNumber int) error {
	if s.VerboseLvl > 0 {
		fmt.Println("Initiating communication...")
	}
	err := s.initiateCommunication()
	if err != nil {
		return err
	}
	defer s.Close()

	deleteMsg := BuildDeletePictureMessage(pictureNumber)
	if s.VerboseLvl > 0 {
		fmt.Printf("Send delete picture command: %s\n", deleteMsg.pretty)
	}
	err = s.sendCommand(deleteMsg)
	if err != nil {
		return err
	}

	response, err := s.readPacket(true, defaultPacketParser{verboseLvl: s.VerboseLvl})
	if err != nil {
		return err
	}

	success, err := GetParser(s.VerboseLvl).ParseDeletePicturePacket(response)
	if err != nil {
		return fmt.Errorf("Parsing issue: %w", err)
	}
	if !success {
		return fmt.Errorf("Failed to delete picture %d", pictureNumber)
	}
	if s.VerboseLvl > 0 {
		fmt.Printf("Picture %d deleted successfully\n", pictureNumber)
	}
	return nil
}

func (s *SerialClient) setBaudRate(baudRate BaudRate) error {
	if s.VerboseLvl > 0 {
		fmt.Println("Initiating communication...")
	}
	err := s.initiateCommunication()
	if err != nil {
		return err
	}

	if s.VerboseLvl > 0 {
		fmt.Printf("Setting baudrate to %d...\n", baudRate)
	}
	var baudRateMsg = BuildSetBaudrateMessage(baudRate)
	err = s.sendCommand(baudRateMsg)
	if err != nil {
		return err
	}
	response, err := s.readPacket(true, defaultPacketParser{verboseLvl: s.VerboseLvl})
	if err != nil {
		return err
	}
	if s.VerboseLvl > 0 {
		fmt.Printf("Response to set baudrate: % X\n", response)
	}
	s.BaudRate = GetBaudRateAsInt(baudRate)
	s.pause()
	s.Close()
	return nil

}

// Close closes the serial port after sending EOT
func (s *SerialClient) Close() error {
	err := s.sendMessage(END_OF_COMMUNICATION_MSG)
	if err != nil {
		fmt.Println("Error sending EOT:", err)
	}
	s.Port.Drain()
	if s.VerboseLvl > 0 {
		fmt.Println("Closing port...")
	}

	s.Port.Close()
	return nil
}

// readPacket reads a packet from the serial port, handling DLE stuffing
// set acknowledge to true to send ACK after reading the whole packet (ack is always send after each part of a multi-part packet)
// set checksumToVerify to true to verify the checksum at the end of the packet
func (s *SerialClient) readPacket(acknowledge bool, parser packetParser) ([]byte, error) {
	if s.VerboseLvl > 0 {
		fmt.Println("Reading packet...")
	}
	startSequence := true
	port := s.Port
	var buffer []byte
	var data []byte
	for {
		if startSequence {
			firstByte, err := readByte(port)
			if err != nil {
				return nil, err
			}
			if firstByte != DLE {
				return nil, fmt.Errorf("expected DLE (0x%02X), got 0x%02X", DLE, firstByte)
			}
			if s.VerboseLvl > 1 {
				fmt.Printf("Received DLE (0x%02X)\n", firstByte)
			}
			secondByte, err := readByte(port)
			if err != nil {
				return nil, err
			}
			if secondByte != STX {
				return nil, fmt.Errorf("expected STX (0x%02X), got 0x%02X", STX, secondByte)
			}
			if s.VerboseLvl > 1 {
				fmt.Printf("Received STX (0x%02X)\n", secondByte)
			}
			startSequence = false
		}
		b, err := readByte(port)
		if err != nil {
			return nil, err
		}

		if b == DLE {
			// Peek the next byte
			if s.VerboseLvl > 1 {
				fmt.Printf("Received DLE (0x%02X), peeking next byte...\n", b)
			}
			nextByte, err := readByte(port)
			if err != nil {
				return nil, err
			}
			if nextByte == ETX || nextByte == ETB {
				// End of text
				if s.VerboseLvl > 1 {
					fmt.Printf("Received ETX or ETB (0x%02X)\n", nextByte)
				}
				checksum, err := readByte(port)
				if err != nil {
					return nil, err
				}
				if s.VerboseLvl > 1 {
					fmt.Printf("Received checksum byte: 0x%02X\n", checksum)
				}
				computedChecksum := xor(append(buffer, nextByte))
				if checksum != computedChecksum {
					return nil, fmt.Errorf("checksum mismatch: expected 0x%02X, got 0x%02X", computedChecksum, checksum)
				}
				// End of packet
				if nextByte == ETX {
					if s.VerboseLvl > 0 {
						fmt.Println("End of packet reached with ETX.")
					}
					break
				} else {
					// ETB, more data will follow
					if s.VerboseLvl > 1 {
						fmt.Println("End of transmission block reached with ETB, more data will follow.")
					}
					// Acknowledge the block
					err = s.sendMessage(ACK_MSG)
					if err != nil {
						return nil, err
					}
					startSequence = true
					buffer, err = parser.Parse(buffer)
					if err != nil {
						return nil, err
					}
					data = append(data, buffer...)
					buffer = []byte{}
				}
			} else if nextByte == DLE {
				// Escaped DLE, add one DLE to data
				if s.VerboseLvl > 1 {
					fmt.Printf("Received escaped DLE (0x%02X), adding to data.\n", nextByte)
				}
				buffer = append(buffer, DLE)
			} else {
				return nil, fmt.Errorf("unexpected byte after DLE: 0x%02X", nextByte)
			}
		} else {
			// Regular byte, add to data
			buffer = append(buffer, b)
		}
	}
	if s.VerboseLvl > 0 {
		fmt.Println("Packet is read.")
	}
	if acknowledge {
		err := s.sendMessage(ACK_MSG)
		if err != nil {
			return nil, err
		}
	}
	buffer, err := parser.Parse(buffer)
	if err != nil {
		return nil, err
	}
	data = append(data, buffer...) // Skip the 4 bytes header
	return data, nil
}

// readByte reads a single byte from the serial port with timeout handling
func readByte(port serial.Port) (byte, error) {
	buff := make([]byte, 1)
	n, err := port.Read(buff)
	if err != nil {
		log.Fatal(err)
	}
	if n == 0 {
		return 0, fmt.Errorf("timeout")
	}
	return buff[0], nil
}
