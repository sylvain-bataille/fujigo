package fuji

// XOR all the bytes of the data slice
func xor(data []byte) byte {
	var result byte = 0
	for _, b := range data {
		result ^= b
	}
	return result
}
