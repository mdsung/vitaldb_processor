package vital

import (
	"encoding/binary"
)

// parseTrkInfo parses TRKINFO packet (packet type 0)
// Contains track metadata: type, format, name, unit, sample rate, gain, etc.
func parseTrkInfo(pkt []byte, vf *VitalFile) {
	if len(pkt) < 51 { // 최소 필요 길이
		return
	}
	pos := 0
	tid := binary.LittleEndian.Uint16(pkt[pos : pos+2]) // tid (사용 안함)
	pos += 2
	trktype := pkt[pos]
	pos++
	fmtcode := pkt[pos]
	pos++
	name, n := unpackStr(pkt, pos)
	pos += n
	if pos >= len(pkt) {
		return
	}
	unit, n := unpackStr(pkt, pos)
	pos += n
	if pos+20 > len(pkt) {
		return
	}
	mindisp := bytesToFloat32(pkt[pos : pos+4])
	pos += 4
	maxdisp := bytesToFloat32(pkt[pos : pos+4])
	pos += 4
	col := binary.LittleEndian.Uint32(pkt[pos : pos+4])
	pos += 4
	srate := bytesToFloat32(pkt[pos : pos+4])
	pos += 4
	gain := bytesToFloat64(pkt[pos : pos+8])
	pos += 8
	offset := bytesToFloat64(pkt[pos : pos+8])
	pos += 8
	if pos >= len(pkt) {
		return
	}
	montype := pkt[pos]
	pos++
	if pos+4 > len(pkt) {
		return
	}
	did := binary.LittleEndian.Uint32(pkt[pos : pos+4])

	// 디바이스 ID로 디바이스 이름 찾기
	deviceName := ""
	if dname, exists := vf.DevIDs[did]; exists {
		deviceName = dname
	}

	// 트랙 이름을 "디바이스명/트랙명" 형태로 구성
	fullTrackName := name
	if deviceName != "" {
		fullTrackName = deviceName + "/" + name
	}

	// tid와 트랙 이름 매핑 저장 (Python의 tid_dtnames와 동일)
	vf.TrkIDs[tid] = fullTrackName

	trk := Track{
		Name:    fullTrackName, // Python VitalDB 호환성: "Device/TrackName" 형식
		Type:    trktype,
		Fmt:     fmtcode,
		Unit:    unit,
		SRate:   srate,
		Gain:    gain,
		Offset:  offset,
		Mindisp: mindisp,
		Maxdisp: maxdisp,
		Col:     col,
		Montype: montype,
		DName:   deviceName,
		Recs:    []Rec{},
	}
	vf.Trks[fullTrackName] = trk
	vf.Order = append(vf.Order, fullTrackName)
}
