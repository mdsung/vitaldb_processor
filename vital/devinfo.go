package vital

import (
	"encoding/binary"
)

// parseDevInfo parses DEVINFO packet (packet type 9)
// Contains device information: device ID, type name, device name, port
func parseDevInfo(pkt []byte, vf *VitalFile) {
	if len(pkt) < 4 {
		return
	}
	pos := 0
	did := binary.LittleEndian.Uint32(pkt[pos : pos+4])
	pos += 4
	if pos >= len(pkt) {
		return
	}
	typename, n := unpackStr(pkt, pos)
	pos += n
	if pos >= len(pkt) {
		return
	}
	name, n := unpackStr(pkt, pos)
	pos += n
	port := ""
	if pos < len(pkt) {
		port, _ = unpackStr(pkt, pos)
	}
	dev := Device{Name: name, TypeName: typename, Port: port}
	vf.Devs[name] = dev
	vf.DevIDs[did] = name // did -> device name 매핑 저장
}
 