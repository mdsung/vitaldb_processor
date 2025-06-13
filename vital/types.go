package vital

// Value represents the possible data types that can be stored in a Rec
type Value interface {
	~float32 | ~float64 | ~int8 | ~uint8 | ~int16 | ~uint16 | ~int32 | ~uint32 | ~string |
		~[]float32 | ~[]float64
}

// Device represents a medical device in the VitalDB file
type Device struct {
	Name     string
	TypeName string
	Port     string
}

// Track represents a data track in the VitalDB file
type Track struct {
	Name    string
	Type    uint8
	Fmt     uint8
	Unit    string
	SRate   float32
	Gain    float64
	Offset  float64
	Mindisp float32
	Maxdisp float32
	Col     uint32
	Montype uint8
	DName   string
	Recs    []Rec
}

// Rec represents a single data record within a track
type Rec struct {
	Dt  float64
	Val any // 더 명확한 표기를 위해 any 사용 (Go 1.18+에서 interface{}의 별칭)
}

// VitalFile represents the complete structure of a VitalDB file
type VitalFile struct {
	Devs    map[string]Device
	Trks    map[string]Track
	DtStart float64
	DtEnd   float64
	Dgmt    int16
	Order   []string
	DevIDs  map[uint32]string // did -> device name 매핑
	TrkIDs  map[uint16]string // tid -> track name 매핑
}
