package mySocket

func DecodeUint64(data []byte) uint64 {
	return (uint64(data[7]) << 56) | (uint64(data[6]) << 48) | (uint64(data[5]) << 40) | ((uint64(data[4]) << 32) | (uint64(data[3]) << 24)) | (uint64(data[2]) << 16) | (uint64(data[1]) << 8) | uint64(data[0])
}
func DecodeUint32(data []byte) uint32 {
	return (uint32(data[3]) << 24) | (uint32(data[2]) << 16) | (uint32(data[1]) << 8) | uint32(data[0])
}

func DecodeUint16(data []byte) uint16 {
	return (uint16(data[1]) << 8) | uint16(data[0])
}

//small Endian
func EncodeUint64(n uint64, b []byte) {
	b[0] = byte(n & 0xFF)
	b[1] = byte((n >> 8) & 0xFF)
	b[2] = byte((n >> 16) & 0xFF)
	b[3] = byte((n >> 24) & 0xFF)
	b[4] = byte((n >> 32) & 0xFF)
	b[5] = byte((n >> 40) & 0xFF)
	b[6] = byte((n >> 48) & 0xFF)
	b[7] = byte((n >> 56) & 0xFF)
}
func EncodeUint32(n uint32, b []byte) {
	b[0] = byte(n & 0xFF)
	b[1] = byte((n >> 8) & 0xFF)
	b[2] = byte((n >> 16) & 0xFF)
	b[3] = byte((n >> 24) & 0xFF)
}
func EncodeUint16(n uint16, b []byte) {
	b[0] = byte(n & 0xFF)
	b[1] = byte((n >> 8) & 0xFF)
}
