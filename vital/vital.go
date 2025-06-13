package vital

import (
	"compress/gzip"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

func NewVitalFile(path string) (*VitalFile, error) {
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

	magic := make([]byte, 4)
	if _, err := io.ReadFull(gz, magic); err != nil {
		return nil, err
	}
	if string(magic) != "VITA" {
		return nil, errors.New("not a vital file")
	}

	var version uint32
	if err := binary.Read(gz, binary.LittleEndian, &version); err != nil {
		return nil, err
	}
	var headerlen uint16
	if err := binary.Read(gz, binary.LittleEndian, &headerlen); err != nil {
		return nil, err
	}
	header := make([]byte, headerlen)
	if _, err := io.ReadFull(gz, header); err != nil {
		return nil, err
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

	for {
		hdr := make([]byte, 5)
		if _, err := io.ReadFull(gz, hdr); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		pktType := hdr[0]
		pktLen := binary.LittleEndian.Uint32(hdr[1:5])
		pkt := make([]byte, pktLen)
		if _, err := io.ReadFull(gz, pkt); err != nil {
			return nil, err
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
	}
	return vf, nil
}
