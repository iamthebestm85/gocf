package fingerprint

type H2Settings struct {
	HeaderTableSize      uint32
	EnablePush           uint32
	MaxConcurrentStreams uint32
	InitialWindowSize    uint32
	MaxFrameSize         uint32
	MaxHeaderListSize    uint32
	WindowUpdateSize     uint32
	PriorityWeight       uint32
	PriorityDependency   uint32
	PriorityExclusive    bool
}

// ChromeH2Settings returns HTTP/2 settings matching Chrome 138 on Windows.
func ChromeH2Settings() H2Settings {
	return H2Settings{
		HeaderTableSize:      65536,
		EnablePush:           0,
		MaxConcurrentStreams: 0,
		InitialWindowSize:    6291456,
		MaxFrameSize:         16384,
		MaxHeaderListSize:    262144,
		WindowUpdateSize:     15663105,
		PriorityWeight:       256,
		PriorityDependency:   0,
		PriorityExclusive:    false,
	}
}

func (s H2Settings) ToSettingsMap() map[uint32]uint32 {
	m := map[uint32]uint32{
		1: s.HeaderTableSize,
		2: s.EnablePush,
		4: s.InitialWindowSize,
		6: s.MaxHeaderListSize,
	}
	if s.MaxConcurrentStreams > 0 {
		m[3] = s.MaxConcurrentStreams
	}
	if s.MaxFrameSize > 0 {
		m[5] = s.MaxFrameSize
	}
	return m
}
