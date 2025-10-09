package fuji

import (
	"fmt"
	"log"
	"time"

	"go.bug.st/serial"
)

type SerialClient struct {
	Verbose        bool
	Device         string
	BaudRate       int
	Port           serial.Port
	DefaultTimeout time.Duration
}

// NewSerialClient creates a new SerialClient with the specified settings
func NewSerialClient(verbose bool, device string, baudRate int) *SerialClient {
	return &SerialClient{Verbose: verbose, Device: device, BaudRate: baudRate, DefaultTimeout: time.Duration(1) * time.Second}
}

// sendBytes sends raw bytes to the serial port
func (s *SerialClient) sendBytes(data []byte) (int, error) {
	n, err := s.Port.Write(data)
	if err != nil {
		return 0, err
	}
	if s.Verbose {
		fmt.Printf("Sent %v bytes\n", n)
	}
	return n, nil
}

// send a command, which is a message encapsulated between delimitation messages
// once the command is sent completely, the function wait for ACK
func (s *SerialClient) SendCommand(msg message) error {
	// Send DLE STX to indicate start of text
	err := s.SendMessage(BEGIN_MSG)
	if err != nil {
		return err
	}
	// Send data message
	err = s.SendMessage(msg)
	if err != nil {
		return err
	}
	// Send DLE ETX and checksum
	err = s.SendMessage(GetEndTextMessageWithChecksum(msg.data))
	if err != nil {
		return err
	}

	err = s.waitForACK()
	if err != nil {
		return err
	}
	return nil
}

func (s *SerialClient) SendMessage(msg message) error {
	if s.Verbose {
		fmt.Printf("Sending message: ")
		for _, b := range msg.data {
			fmt.Printf("0x%02X ", b)
		}
		fmt.Printf(" (%s)", msg.pretty)
		fmt.Println()
	}
	n, err := s.sendBytes(msg.data)
	if err != nil {
		return err
	}
	if n != len(msg.data) {
		return fmt.Errorf("sent %d bytes, expected to send %d bytes", n, len(msg.data))
	}
	if s.Verbose {
		fmt.Printf("Message sent: %s\n", msg.pretty)
	}
	return nil
}

func (s *SerialClient) waitForACK() error {
	// Wait for ACK
	if s.Verbose {
		fmt.Println("Waiting for ACK...")
	}
	b, err := readByte(s.Port)
	if err != nil {
		return err
	}
	if b != ACK {
		return fmt.Errorf("expected ACK (0x%02X), got 0x%02X", ACK, b)
	}
	if s.Verbose {
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
	if s.Verbose {
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
	// if s.Verbose {
	// 	fmt.Println("Flushing input buffer...")
	// }
	// err = s.Port.ResetInputBuffer()
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
	err = s.SendMessage(ENQUIRY_MSG)
	err = s.waitForACK()
	if err != nil {
		return err
	}
	return nil
}

// ListPorts lists available serial ports
func (s *SerialClient) ListPorts() ([]string, error) {
	if s.Verbose {
		fmt.Println("Getting serial ports...")
	}
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, err
	}
	if s.Verbose {
		fmt.Printf("Found %d ports\n", len(ports))
	}
	return ports, nil
}

// GetModel return information about camera model
func (s *SerialClient) GetModel() (string, error) {
	if s.Verbose {
		fmt.Println("Initiating communication...")
	}
	err := s.initiateCommunication()
	if err != nil {
		return "", err
	}
	defer s.Close()

	err = s.SendCommand(GET_CAMERA_VERSION_MSG)
	if err != nil {
		return "", err
	}

	response, err := s.readPacket(true, 0)
	if err != nil {
		return "", err
	}

	model, err := GetParser(s.Verbose).ParseCameraVersionPacket(response)
	if err != nil {
		return "", fmt.Errorf("Parsing issue: %w", err)
	}
	return model, nil
}

// CountPictures returns the number of pictures stored on the camera
func (s *SerialClient) CountPictures() (int, error) {
	if s.Verbose {
		fmt.Println("Initiating communication...")
	}
	err := s.initiateCommunication()
	if err != nil {
		return 0, err
	}
	defer s.Close()

	err = s.SendCommand(COUNT_PICTURES_MSG)
	if err != nil {
		return 0, err
	}

	response, err := s.readPacket(true, 0)
	if err != nil {
		return 0, err
	}

	count, err := GetParser(s.Verbose).ParseCountPicturesPacket(response)
	if err != nil {
		return 0, fmt.Errorf("Parsing issue: %w", err)
	}
	return count, nil
}

func (s *SerialClient) setBaudrateSetting(baudRate int) error {
	if s.Verbose {
		fmt.Printf("Resetting baudrate to %d...\n", baudRate)
	}
	s.BaudRate = baudRate
	s.pause()
	return nil
}

func (s *SerialClient) pause() {
	time.Sleep(100 * time.Millisecond)
}

func (s *SerialClient) DownloadPicture(pictureNumber int) ([]byte, error) {
	// Set higher baudrate for faster download
	initialBaudRate := s.BaudRate
	err := s.setBaudRate(BAUD_115200)
	if err != nil {
		return nil, err
	}

	if s.Verbose {
		fmt.Println("Initiating communication...")
	}
	err = s.initiateCommunication()
	if err != nil {
		return nil, err
	}
	defer s.Close()
	defer s.setBaudrateSetting(initialBaudRate)

	downloadMsg := GetDownloadPictureMessage(pictureNumber)
	err = s.SendCommand(downloadMsg)
	if err != nil {
		return nil, err
	}

	s.Port.SetReadTimeout(time.Duration(30) * time.Second)
	pictureData, err := s.readPacket(true, 4) // Skip the 4 bytes header
	s.Port.SetReadTimeout(s.DefaultTimeout)
	if err != nil {
		return nil, err
	}

	if s.Verbose {
		fmt.Printf("Downloaded picture %d, size: %d bytes\n", pictureNumber, len(pictureData))
	}
	// Restore initial baudrate
	return pictureData, nil
}

func (s *SerialClient) setBaudRate(baudRate BaudRate) error {
	if s.Verbose {
		fmt.Println("Initiating communication...")
	}
	err := s.initiateCommunication()
	if err != nil {
		return err
	}

	if s.Verbose {
		fmt.Printf("Setting baudrate to %d...\n", baudRate)
	}
	var baudRateMsg = GetSetBaudrateMessage(baudRate)
	err = s.SendCommand(baudRateMsg)
	if err != nil {
		return err
	}
	response, err := s.readPacket(true, 0)
	if err != nil {
		return err
	}
	if s.Verbose {
		fmt.Printf("Response to set baudrate: % X\n", response)
	}
	//s.Close()
	s.BaudRate = GetBaudRateAsInt(baudRate)

	s.pause()
	// err = s.SendMessage(END_OF_COMMUNICATION_MSG)
	// if err != nil {
	// 	fmt.Println("Error sending EOT:", err)
	// }

	//Sleep 5 seconds to ensure the message is sent before closing the port
	// s.Port.Drain()
	// time.Sleep(1 * time.Second)
	// if s.Verbose {
	// 	fmt.Println("Closing port...")
	// }

	// s.Port.Close()

	//Send ENQ again to re-initiate communication at new baudrate
	s.Close()
	// s.Port.Drain()
	// s.Port.Close()

	return nil

}

// Close closes the serial port after sending EOT
func (s *SerialClient) Close() error {
	err := s.SendMessage(END_OF_COMMUNICATION_MSG)
	if err != nil {
		fmt.Println("Error sending EOT:", err)
	}
	//Sleep 5 seconds to ensure the message is sent before closing the port
	s.Port.Drain()
	if s.Verbose {
		fmt.Println("Closing port...")
	}

	s.Port.Close()
	return nil
}

// readPacket reads a packet from the serial port, handling DLE stuffing
// set acknowledge to true to send ACK after reading the whole packet (ack is always send after each part of a multi-part packet)
// set checksumToVerify to true to verify the checksum at the end of the packet
// TODO refactor the skip bytes part to handle photo downloading and other commands in a better way, because this is very spaghetti
// TODO set higher baudrate when downloading pictures
func (s *SerialClient) readPacket(acknowledge bool, skipBytes int) ([]byte, error) {
	if s.Verbose {
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
			if s.Verbose {
				fmt.Printf("Received DLE (0x%02X)\n", firstByte)
			}
			secondByte, err := readByte(port)
			if err != nil {
				return nil, err
			}
			if secondByte != STX {
				return nil, fmt.Errorf("expected STX (0x%02X), got 0x%02X", STX, secondByte)
			}
			if s.Verbose {
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
			if s.Verbose {
				fmt.Printf("Received DLE (0x%02X), peeking next byte...\n", b)
			}
			nextByte, err := readByte(port)
			if err != nil {
				return nil, err
			}
			if nextByte == ETX || nextByte == ETB {
				// End of text
				if s.Verbose {
					fmt.Printf("Received ETX or ETB (0x%02X)\n", nextByte)
				}
				checksum, err := readByte(port)
				if err != nil {
					return nil, err
				}
				if s.Verbose {
					fmt.Printf("Received checksum byte: 0x%02X\n", checksum)
				}
				computedChecksum := xor(append(buffer, nextByte))
				if checksum != computedChecksum {
					return nil, fmt.Errorf("checksum mismatch: expected 0x%02X, got 0x%02X", computedChecksum, checksum)
				}
				// End of packet
				if nextByte == ETX {
					if s.Verbose {
						fmt.Println("End of packet reached with ETX.")
					}
					break
				} else {
					// ETB, more data will follow
					if s.Verbose {
						fmt.Println("End of transmission block reached with ETB, more data will follow.")
					}
					// Acknowledge the block
					err = s.SendMessage(ACK_MSG)
					if err != nil {
						return nil, err
					}
					startSequence = true
					data = append(data, buffer[skipBytes:]...) // Skip the 4 bytes header
					buffer = []byte{}
				}
			} else if nextByte == DLE {
				// Escaped DLE, add one DLE to data
				if s.Verbose {
					fmt.Printf("Received escaped DLE (0x%02X), adding to data.\n", nextByte)
				}
				buffer = append(buffer, DLE)
			} else {
				return nil, fmt.Errorf("unexpected byte after DLE: 0x%02X", nextByte)
			}
		} else {
			if s.Verbose {
				//fmt.Printf("Received byte: 0x%02X\n", b)
			}
			// Regular byte, add to data
			buffer = append(buffer, b)
		}
	}
	if s.Verbose {
		fmt.Println("Packet is read.")
	}
	if acknowledge {
		err := s.SendMessage(ACK_MSG)
		if err != nil {
			return nil, err
		}
	}
	data = append(data, buffer[skipBytes:]...) // Skip the 4 bytes header
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
		//Timeout
		return 0, fmt.Errorf("timeout")
	}
	return buff[0], nil
}
