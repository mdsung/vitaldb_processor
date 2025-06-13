package vital

import (
	"compress/gzip"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

// NewOptimizedVitalFileV3Fixed implements Phase 3 optimizations (FIXED):
// - Smart pre-allocation based on file size estimation
// - Optimized I/O patterns with larger buffers
// - **USES ORIGINAL PARSING LOGIC** for 100% accuracy
func NewOptimizedVitalFileV3Fixed(path string) (*VitalFile, error) {
	// Get file size for smart pre-allocation
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	fileSize := fileInfo.Size()

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	// Smart buffer sizing based on file size
	bufferSize := 512 * 1024  // Default 512KB
	if fileSize < 1024*1024 { // < 1MB
		bufferSize = 256 * 1024 // 256KB for smaller files
	} else if fileSize > 10*1024*1024 { // > 10MB
		bufferSize = 1024 * 1024 // 1MB for larger files
	}

	buffer := make([]byte, bufferSize)

	// Read initial chunk
	n, err := io.ReadFull(gz, buffer[:32])
	if err != nil && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	if n < 10 {
		return nil, errors.New("file too short")
	}

	pos := 0

	// Verify magic number (inline)
	if string(buffer[pos:pos+4]) != "VITA" {
		return nil, errors.New("not a vital file")
	}
	pos += 4

	// Read version and header length (inline)
	version := binary.LittleEndian.Uint32(buffer[pos:])
	pos += 4
	_ = version

	headerlen := binary.LittleEndian.Uint16(buffer[pos:])
	pos += 2

	// Read header if not already in buffer
	headerEnd := pos + int(headerlen)
	if headerEnd > n {
		if headerEnd > len(buffer) {
			newBuffer := make([]byte, headerEnd)
			copy(newBuffer, buffer[:n])
			buffer = newBuffer
		}

		additionalBytes, err := io.ReadFull(gz, buffer[n:headerEnd])
		if err != nil && err != io.ErrUnexpectedEOF {
			return nil, err
		}
		n += additionalBytes
	}

	// Parse header data directly from buffer
	header := buffer[pos:headerEnd]
	pos = headerEnd

	var dgmt int16
	var dtstart, dtend float64

	if len(header) >= 2 {
		dgmt = int16(binary.LittleEndian.Uint16(header[:2]))
	}
	if len(header) >= 18 {
		dtstart = bytesToFloat64(header[10:18])
	}
	if len(header) >= 26 {
		dtend = bytesToFloat64(header[18:26])
	}

	// Smart pre-allocation based on file size estimation
	estimatedDevices := int(fileSize/(100*1024)) + 10 // Rough estimate
	estimatedTracks := int(fileSize/(50*1024)) + 20   // Rough estimate

	// Ensure minimum sizes
	if estimatedDevices < 10 {
		estimatedDevices = 10
	}
	if estimatedTracks < 50 {
		estimatedTracks = 50
	}

	// Initialize VitalFile with pre-allocated maps
	vf := &VitalFile{
		Devs:    make(map[string]Device, estimatedDevices),
		Trks:    make(map[string]Track, estimatedTracks),
		DtStart: dtstart,
		DtEnd:   dtend,
		Dgmt:    dgmt,
		Order:   make([]string, 0, estimatedTracks),
		DevIDs:  make(map[uint32]string, estimatedDevices),
		TrkIDs:  make(map[uint16]string, estimatedTracks),
	}

	// Optimized streaming processor
	reader := &fixedStreamingReader{
		gz:     gz,
		buffer: buffer,
		pos:    pos,
		size:   n,
	}

	for {
		// Read packet header (5 bytes)
		hdr, err := reader.readBytes(5)
		if err != nil {
			break
		}

		pktType := hdr[0]
		pktLen := binary.LittleEndian.Uint32(hdr[1:5])

		// Read packet data
		pktData, err := reader.readBytes(int(pktLen))
		if err != nil {
			break
		}

		// USE ORIGINAL PARSING FUNCTIONS for 100% accuracy
		switch pktType {
		case 9:
			parseDevInfo(pktData, vf) // ORIGINAL function
		case 0:
			parseTrkInfo(pktData, vf) // ORIGINAL function
		case 1:
			parseRec(pktData, vf) // ORIGINAL function
		case 6:
			parseCmd(pktData, vf) // ORIGINAL function
		}
	}

	return vf, nil
}

type fixedStreamingReader struct {
	gz     *gzip.Reader
	buffer []byte
	pos    int
	size   int
	eof    bool
}

func (r *fixedStreamingReader) readBytes(n int) ([]byte, error) {
	// Ensure we have enough data
	for r.pos+n > r.size && !r.eof {
		// Move remaining data to start
		if r.pos > 0 {
			copy(r.buffer, r.buffer[r.pos:r.size])
			r.size -= r.pos
			r.pos = 0
		}

		// Expand buffer if needed (with growth strategy)
		if r.size+n > len(r.buffer) {
			// Smart growth: 1.5x or needed size, whichever is larger
			newSize := maxFixed(len(r.buffer)*3/2, r.size+n)
			newBuffer := make([]byte, newSize)
			copy(newBuffer, r.buffer[:r.size])
			r.buffer = newBuffer
		}

		// Read more data in larger chunks
		readSize := maxFixed(n, 64*1024) // At least 64KB
		if r.size+readSize > len(r.buffer) {
			readSize = len(r.buffer) - r.size
		}

		bytesRead, err := r.gz.Read(r.buffer[r.size : r.size+readSize])
		r.size += bytesRead
		if err != nil {
			if err == io.EOF {
				r.eof = true
			} else {
				return nil, err
			}
		}
	}

	if r.pos+n > r.size {
		return nil, io.EOF
	}

	result := r.buffer[r.pos : r.pos+n]
	r.pos += n
	return result, nil
}

func maxFixed(a, b int) int {
	if a > b {
		return a
	}
	return b
}
 