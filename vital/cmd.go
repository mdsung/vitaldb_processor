package vital

import (
	"encoding/binary"
)

// parseCmd parses CMD packet (packet type 2)
// Contains commands like track ordering (TRK_ORDER)
func parseCmd(pkt []byte, vf *VitalFile) {
	// CMD 패킷 파싱 (트랙 순서 등)
	if len(pkt) < 4 {
		return
	}

	pos := 0
	cmdType := binary.LittleEndian.Uint32(pkt[pos : pos+4])
	pos += 4

	switch cmdType {
	case 1: // TRK_ORDER - 트랙 순서 명령
		// 트랙 ID 리스트를 읽어서 Order 업데이트
		if pos+2 > len(pkt) {
			return
		}
		count := binary.LittleEndian.Uint16(pkt[pos : pos+2])
		pos += 2

		newOrder := make([]string, 0, count)
		for i := 0; i < int(count); i++ {
			if pos+2 > len(pkt) {
				break
			}
			trkID := binary.LittleEndian.Uint16(pkt[pos : pos+2])
			pos += 2

			// 트랙 ID를 인덱스로 사용해서 트랙 이름 찾기
			if int(trkID-1) < len(vf.Order) {
				trackName := vf.Order[int(trkID-1)]
				newOrder = append(newOrder, trackName)
			}
		}

		// 순서 업데이트
		if len(newOrder) > 0 {
			vf.Order = newOrder
		}

	default:
		// 다른 CMD 타입들은 일단 스킵
		return
	}
}
