package vital

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
	Val interface{}
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
