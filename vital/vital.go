package vital

import (
	"compress/gzip"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

func NewVitalFile(path string) (*VitalFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gz.Close()

	magic := make([]byte, 4)
	if _, err := io.ReadFull(gz, magic); err != nil {
		return nil, fmt.Errorf("failed to read magic: %w", err)
	}
	if string(magic) != "VITA" {
		return nil, errors.New("not a vital file")
	}

	var version uint32
	if err := binary.Read(gz, binary.LittleEndian, &version); err != nil {
		return nil, fmt.Errorf("failed to read version: %w", err)
	}
	var headerlen uint16
	if err := binary.Read(gz, binary.LittleEndian, &headerlen); err != nil {
		return nil, fmt.Errorf("failed to read header length: %w", err)
	}
	header := make([]byte, headerlen)
	if _, err := io.ReadFull(gz, header); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	// 헤더 길이 체크 후 안전하게 파싱
	var dgmt int16
	var dtstart, dtend float64

	if len(header) >= 2 {
		dgmt = int16(binary.LittleEndian.Uint16(header[0:2]))
	}
	if len(header) >= 18 {
		dtstart = bytesToFloat64(header[10:18])
	}
	if len(header) >= 26 {
		dtend = bytesToFloat64(header[18:26])
	}

	vf := &VitalFile{
		Devs:    make(map[string]Device),
		Trks:    make(map[string]Track),
		DtStart: dtstart,
		DtEnd:   dtend,
		Dgmt:    dgmt,
		Order:   []string{},
		DevIDs:  make(map[uint32]string),
		TrkIDs:  make(map[uint16]string),
	}

	pktCount := 0
	for {
		hdr := make([]byte, 5)
		if _, err := io.ReadFull(gz, hdr); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to read packet header at packet %d: %w", pktCount, err)
		}
		pktType := hdr[0]
		pktLen := binary.LittleEndian.Uint32(hdr[1:5])

		// 패킷 길이 검증
		if pktLen > 100*1024*1024 { // 100MB 제한
			return nil, fmt.Errorf("packet %d has invalid length: %d bytes", pktCount, pktLen)
		}

		pkt := make([]byte, pktLen)
		if _, err := io.ReadFull(gz, pkt); err != nil {
			// 파일 끝에서 불완전한 패킷은 무시 (Python VitalDB와 동일한 방식)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return nil, fmt.Errorf("failed to read packet %d (type %d, length %d): %w", pktCount, pktType, pktLen, err)
		}

		switch pktType {
		case 9:
			parseDevInfo(pkt, vf)
		case 0:
			parseTrkInfo(pkt, vf)
		case 1:
			parseRec(pkt, vf)
		case 6:
			parseCmd(pkt, vf)
		default:
			// skip
		}
		pktCount++
	}
	return vf, nil
}
