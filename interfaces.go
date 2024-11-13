package dotprops

// TextMarshaler interface as defined in the encoding package
type TextMarshaler interface {
	MarshalText() (text []byte, err error)
}

// TextUnmarshaler interface as defined in the encoding package
type TextUnmarshaler interface {
	UnmarshalText(text []byte) error
}

// PropMarshaler allows custom marshaling of a single property.
type PropMarshaler interface {
	MarshalProp() (key string, value string, err error)
}

// PropUnmarshaller allows custom unmarshaling of a single property.
type PropUnmarshaller interface {
	UnmarshalProp(key string, value string) error
}
