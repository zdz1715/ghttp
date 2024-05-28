package encoding

// Codec defines the interface Transport uses to encode and decode messages.  Note
// that implementations of this interface must be thread safe; a Codec's
// methods can be called from concurrent goroutines.
type Codec interface {
	// Marshal returns the wire format of v.
	Marshal(v interface{}) ([]byte, error)
	// Unmarshal parses the wire format into v.
	Unmarshal(data []byte, v interface{}) error
}

var registeredCodecs = make(map[string]Codec)

func RegisterCodec(name string, codec Codec) {
	if codec == nil {
		panic("cannot register a nil Codec")
	}
	if name == "" {
		panic("cannot register Codec with empty name")
	}
	registeredCodecs[name] = codec
}

func GetCodec(name string) Codec {
	return registeredCodecs[name]
}
