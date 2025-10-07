package fuji

import (
	"fmt"
	"log"
	"time"

	"go.bug.st/serial"
)

type SerialClient struct {
	Verbose  bool
	Device   string
	BaudRate int
	Port     serial.Port
}

// NewSerialClient creates a new SerialClient with the specified settings
func NewSerialClient(verbose bool, device string, baudRate int) *SerialClient {
	return &SerialClient{Verbose: verbose, Device: device, BaudRate: baudRate}
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
	s.Port.SetReadTimeout(time.Duration(1) * time.Second)
	if s.Verbose {
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

	response, err := s.readPacket(true)
	if err != nil {
		return "", err
	}

	model, err := GetParser(s.Verbose).ParseCameraVersionPacket(response)
	if err != nil {
		return "", fmt.Errorf("Parsing issue: %w", err)
	}
	return model, nil
}

func (s *SerialClient) Close() error {
	err := s.SendMessage(END_OF_COMMUNICATION_MSG)
	if err != nil {
		fmt.Println("Error sending EOT:", err)
	}
	if s.Verbose {
		fmt.Println("Closing port...")
	}
	s.Port.Close()
	return nil
}

func (s *SerialClient) readPacket(acknowledge bool) ([]byte, error) {
	if s.Verbose {
		fmt.Println("Reading packet...")
	}
	port := s.Port
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
	var data []byte
	for {
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
			if nextByte == ETX {
				// End of text
				if s.Verbose {
					fmt.Printf("Received ETX (0x%02X), end of packet.\n", nextByte)
				}
				break
			} else if nextByte == DLE {
				// Escaped DLE, add one DLE to data
				if s.Verbose {
					fmt.Printf("Received escaped DLE (0x%02X), adding to data.\n", nextByte)
				}
				data = append(data, DLE)
			} else {
				//not sure
				//return nil, fmt.Errorf("unexpected byte after DLE: 0x%02X", nextByte)
				data = append(data, b)
			}
		} else {
			if s.Verbose {
				fmt.Printf("Received byte: 0x%02X\n", b)
			}
			// Regular byte, add to data
			data = append(data, b)
		}
	}
	if s.Verbose {
		fmt.Println("Packet is read. Data: ")
		for _, b := range data {
			fmt.Printf("0x%02X ", b)
		}
		fmt.Println()
	}
	if acknowledge {
		err = s.SendMessage(ACK_MSG)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

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
