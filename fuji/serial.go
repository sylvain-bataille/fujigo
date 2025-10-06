package fuji

import (
	"fmt"
	"log"
	"time"

	"go.bug.st/serial"
)

const ENQ = 0x05 // Enquiry, used to initiate communication
const ACK = 0x06 // Acknowledge, used to acknowledge receipt of a message
const DLE = 0x10 // Data Link Escape, used to indicate special control characters
const STX = 0x02 // Start of Text, initiate the begining of a message
const ETX = 0x03 // End of Text, indicate the end of a message
const EOT = 0x04 // End of Transmission, used to terminate communication

func openCommunication(portName string) (serial.Port, error) {
	mode := &serial.Mode{
		BaudRate: 9600,
		Parity:   serial.EvenParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}
	port, err := serial.Open("/dev/ttyUSB0", mode)
	if err != nil {
		log.Fatal(err)
	}
	port.SetReadTimeout(time.Duration(1) * time.Second)
	// Flush any existing data
	err = port.ResetInputBuffer()
	if err != nil {
		log.Fatal(err)
	}
	// Initiate communication
	fmt.Printf("Opening port %s\n", portName)
	n, err := port.Write([]byte{ENQ})
	fmt.Printf("Sent %v bytes\n", n)
	if err != nil {
		log.Fatal(err)
	}
	// Wait for ACK
	b, err := readByte(port)
	if err != nil {
		log.Fatal(err)
	}
	if b != ACK {
		log.Fatalf("Expected ACK (0x%02X), got 0x%02X\n", ACK, b)
	}
	fmt.Printf("Received ACK (0x%02X)\n", b)
	return port, nil
}

func ListPorts() ([]string, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, err
	}
	return ports, nil
}

func GetModel() string {
	port, err := openCommunication("/dev/ttyUSB0")
	// Send DLE STX to indicate start of text
	n, err := port.Write([]byte{DLE, STX})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Sent %v bytes\n", n)

	// Send command to get model
	command := []byte{0x00, 0x09, 0x00, 0x00} // Command to get model
	n, err = port.Write(command)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Sent %v bytes\n", n)

	// Send DLE ETX and checksum
	n, err = port.Write([]byte{DLE, ETX, xor(append(command, ETX))})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Sent %v bytes\n", n)

	// Verify acknowledgment
	b, err := readByte(port)
	if err != nil {
		log.Fatal(err)
	}
	if b != ACK {
		log.Fatalf("Expected ACK (0x%02X), got 0x%02X\n", ACK, b)
	}
	fmt.Printf("Received ACK (0x%02X)\n", b)
	response, err := readPacket(port)
	fmt.Printf("Model: %s\n", string(response))

	n, err = port.Write([]byte{ACK})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Sent %v bytes\n", n)

	if err != nil {
		log.Fatal(err)
	}

	n, err = port.Write([]byte{EOT})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Sent %v bytes\n", n)
	port.Close()
	return ""
}

func readPacket(port serial.Port) ([]byte, error) {
	firstByte, err := readByte(port)
	if err != nil {
		return nil, err
	}
	if firstByte != DLE {
		return nil, fmt.Errorf("expected DLE (0x%02X), got 0x%02X", DLE, firstByte)
	}
	secondByte, err := readByte(port)
	if err != nil {
		return nil, err
	}
	if secondByte != STX {
		return nil, fmt.Errorf("expected STX (0x%02X), got 0x%02X", STX, secondByte)
	}
	var data []byte
	for {
		b, err := readByte(port)
		if err != nil {
			return nil, err
		}

		if b == DLE {
			// Peek the next byte
			nextByte, err := readByte(port)
			if err != nil {
				return nil, err
			}
			if nextByte == ETX {
				// End of text
				break
			} else if nextByte == DLE {
				// Escaped DLE, add one DLE to data
				data = append(data, DLE)
			} else {
				//not sure
				//return nil, fmt.Errorf("unexpected byte after DLE: 0x%02X", nextByte)
				data = append(data, b)
			}
		} else {
			data = append(data, b)
		}
	}
	return data, nil
}

func readResponse(port serial.Port) {
	buff := make([]byte, 100)
	for {
		n, err := port.Read(buff)
		if err != nil {
			log.Fatal(err)
			break
		}
		if n == 0 {
			fmt.Println("\nEOF")
			break
		}
		// Print the byte
		fmt.Printf("\nRead %v bytes: ", n)
		for i := 0; i < n; i++ {
			fmt.Printf("0x%02X ", buff[i])
		}
		// Print as string
		fmt.Printf("%v", string(buff[:n]))
	}
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

func xor(data []byte) byte {
	var result byte = 0
	for _, b := range data {
		result ^= b
	}
	return result
}
