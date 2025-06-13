package vital

import (
	"encoding/binary"
)

// parseRec parses REC packet (packet type 1)
// Contains actual data records: WAVE, NUMERIC, or STRING data
func parseRec(pkt []byte, vf *VitalFile) {
	if len(pkt) < 12 {
		return
	}
	pos := 0
	infolen := binary.LittleEndian.Uint16(pkt[pos : pos+2])
	pos += 2
	if pos+10 > len(pkt) {
		return
	}
	dt := bytesToFloat64(pkt[pos : pos+8])
	pos += 8
	trkid := binary.LittleEndian.Uint16(pkt[pos : pos+2])
	pos += 2

	// infolen이 패킷 크기보다 클 수 없음
	if int(infolen) > len(pkt) {
		return
	}

	// Python과 같은 방식으로 dtstart/dtend 업데이트
	if vf.DtStart == 0 || (dt > 0 && dt < vf.DtStart) {
		vf.DtStart = dt
	}
	if dt > vf.DtEnd {
		vf.DtEnd = dt
	}

	// tid를 사용하여 트랙 찾기 (Python의 tid_dtnames와 동일)
	trackName, exists := vf.TrkIDs[trkid]
	if !exists {
		return
	}

	track, exists := vf.Trks[trackName]
	if !exists {
		return
	}

	// 데이터 타입별 파싱
	switch track.Type {
	case 1: // WAVE 타입
		parseWaveData(pkt, pos, dt, trkid, &track, vf)
	case 2: // NUMERIC 타입
		parseNumericData(pkt, pos, dt, &track)
	case 5: // STRING 타입
		parseStringData(pkt, pos, dt, &track)
	}

	// 업데이트된 트랙을 다시 저장
	vf.Trks[trackName] = track
}

// parseWaveData parses WAVE type data from REC packet
func parseWaveData(pkt []byte, pos int, dt float64, trkid uint16, track *Track, vf *VitalFile) {
	if pos+4 > len(pkt) {
		return
	}
	nsamples := binary.LittleEndian.Uint32(pkt[pos : pos+4])
	pos += 4

	// 포맷별 샘플 크기 계산
	var sampleSize int
	switch track.Fmt {
	case 1: // float32
		sampleSize = 4
	case 2: // float64
		sampleSize = 8
	case 3: // int8
		sampleSize = 1
	case 4: // uint8
		sampleSize = 1
	case 5: // int16
		sampleSize = 2
	case 6: // uint16
		sampleSize = 2
	case 7: // int32
		sampleSize = 4
	case 8: // uint32
		sampleSize = 4
	default:
		return
	}

	totalBytes := int(nsamples) * sampleSize
	if pos+totalBytes > len(pkt) {
		return
	}

	// fmt에 따라 적절한 타입의 샘플 배열 생성
	var samples any

	switch track.Fmt {
	case 1: // float32 - 원본 타입 유지
		float32Samples := make([]float32, nsamples)
		for i := 0; i < int(nsamples); i++ {
			samplePos := pos + i*sampleSize
			if samplePos+sampleSize > len(pkt) {
				break
			}
			sample := bytesToFloat32(pkt[samplePos : samplePos+4])
			float32Samples[i] = sample
		}
		samples = float32Samples

	case 2: // float64 - 원본 타입 유지
		float64Samples := make([]float64, nsamples)
		for i := 0; i < int(nsamples); i++ {
			samplePos := pos + i*sampleSize
			if samplePos+sampleSize > len(pkt) {
				break
			}
			sample := bytesToFloat64(pkt[samplePos : samplePos+8])
			float64Samples[i] = sample
		}
		samples = float64Samples

	default: // 정수 타입들 - gain/offset 적용 후 float32로 저장 (메모리 효율성)
		float32Samples := make([]float32, nsamples)
		for i := 0; i < int(nsamples); i++ {
			samplePos := pos + i*sampleSize
			if samplePos+sampleSize > len(pkt) {
				break
			}

			var rawValue float32
			switch track.Fmt {
			case 3: // int8
				rawValue = float32(int8(pkt[samplePos]))
			case 4: // uint8
				rawValue = float32(pkt[samplePos])
			case 5: // int16
				rawValue = float32(int16(binary.LittleEndian.Uint16(pkt[samplePos : samplePos+2])))
			case 6: // uint16
				rawValue = float32(binary.LittleEndian.Uint16(pkt[samplePos : samplePos+2]))
			case 7: // int32
				rawValue = float32(int32(binary.LittleEndian.Uint32(pkt[samplePos : samplePos+4])))
			case 8: // uint32
				rawValue = float32(binary.LittleEndian.Uint32(pkt[samplePos : samplePos+4]))
			}

			// gain/offset 적용
			sample := rawValue*float32(track.Gain) + float32(track.Offset)
			float32Samples[i] = sample
		}
		samples = float32Samples
	}

	track.Recs = append(track.Recs, Rec{Dt: dt, Val: samples})

	// 웨이브 타입의 경우 샘플률을 고려하여 dtend 업데이트 (Python과 동일)
	if track.SRate > 0 {
		recDtend := dt + float64(nsamples)/float64(track.SRate)
		if recDtend > vf.DtEnd {
			vf.DtEnd = recDtend
		}
	}
}

// parseNumericData parses NUMERIC type data from REC packet
func parseNumericData(pkt []byte, pos int, dt float64, track *Track) {
	// 포맷에 따른 데이터 크기 및 타입 결정 - 원본 타입 유지
	var val any
	switch track.Fmt {
	case 1: // float32 - 원본 타입 유지
		if pos+4 > len(pkt) {
			return
		}
		val = bytesToFloat32(pkt[pos : pos+4])
	case 2: // float64 - 원본 타입 유지
		if pos+8 > len(pkt) {
			return
		}
		val = bytesToFloat64(pkt[pos : pos+8])
	case 3: // int8 - 원본 타입 유지
		if pos+1 > len(pkt) {
			return
		}
		val = int8(pkt[pos])
	case 4: // uint8 - 원본 타입 유지
		if pos+1 > len(pkt) {
			return
		}
		val = uint8(pkt[pos])
	case 5: // int16 - 원본 타입 유지
		if pos+2 > len(pkt) {
			return
		}
		val = int16(binary.LittleEndian.Uint16(pkt[pos : pos+2]))
	case 6: // uint16 - 원본 타입 유지
		if pos+2 > len(pkt) {
			return
		}
		val = binary.LittleEndian.Uint16(pkt[pos : pos+2])
	case 7: // int32 - 원본 타입 유지
		if pos+4 > len(pkt) {
			return
		}
		val = int32(binary.LittleEndian.Uint32(pkt[pos : pos+4]))
	case 8: // uint32 - 원본 타입 유지
		if pos+4 > len(pkt) {
			return
		}
		val = binary.LittleEndian.Uint32(pkt[pos : pos+4])
	default:
		// 기본값으로 8바이트 읽기
		if pos+8 > len(pkt) {
			return
		}
		val = bytesToFloat64(pkt[pos : pos+8])
	}
	track.Recs = append(track.Recs, Rec{Dt: dt, Val: val})
}

// parseStringData parses STRING type data from REC packet
func parseStringData(pkt []byte, pos int, dt float64, track *Track) {
	if pos+8 > len(pkt) {
		return
	}
	pos += 4 // 예약 필드 스킵
	if pos+4 > len(pkt) {
		return
	}
	strlen := binary.LittleEndian.Uint32(pkt[pos : pos+4])
	pos += 4
	if pos+int(strlen) > len(pkt) {
		return
	}
	val := string(pkt[pos : pos+int(strlen)])
	track.Recs = append(track.Recs, Rec{Dt: dt, Val: val})
}
