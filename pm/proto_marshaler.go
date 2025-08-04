package pm

import (
	"os"

	grpc_runtime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	// pmWithEmpty is the standard instance of [grpc_runtime.JSONPb] marshaler
	pmAsJSON = MarshalerAsJSON()

	// pmWithEmpty is the standard instance of [grpc_runtime.JSONPb] marshaler with "empty" property
	pmWithEmpty = MarshalerWithEmpty()
)

// MarshalerAsJSON returns a configured [grpc_runtime.JSONPb] instance for serializing/deserializing
// protocol buffer messages with the following properties:
//
//   - When marshaling, empty fields with default values will be skipped.
//
//   - When unmarshaling, unknown fields and enum name values are ignored.
func MarshalerAsJSON() *grpc_runtime.JSONPb {
	return &grpc_runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			Indent:            "    ",
			UseProtoNames:     true,
			EmitUnpopulated:   false,
			EmitDefaultValues: false,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}
}

// MarshalerAsJSON returns a configured [grpc_runtime.JSONPb] instance for serializing/deserializing
// protocol buffer messages with the following properties:
//
//   - When marshaling, empty fields with default values will be populated in the resulting JSON.
//
//   - When unmarshaling, unknown fields and enum name values are ignored.
func MarshalerWithEmpty() *grpc_runtime.JSONPb {
	return &grpc_runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			Indent:          "    ",
			UseProtoNames:   true,
			EmitUnpopulated: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}
}

// Marshal marshals "v" into JSON.
func Marshal(v interface{}) ([]byte, error) {
	return pmAsJSON.Marshal(v)
}

// MarshalFile marshals "v" into JSON and writes the result to a given file.
func MarshalFile(filename string, v interface{}) error {
	b, err := Marshal(v)
	if err != nil {
		return err
	}

	fd, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fd.Close()

	if _, err := fd.Write(b); err != nil {
		return err
	}

	return nil
}

// Unmarshal unmarshals JSON-encoded data into the non-nil value pointed to by "v".
func Unmarshal(data []byte, v interface{}) error {
	return pmAsJSON.Unmarshal(data, v)
}

// UnmarshalFile reads JSON-encoded data from a given file and unmarshals it
// into the non-nil value pointed to by "v".
func UnmarshalFile(filename string, v interface{}) error {
	b, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return Unmarshal(b, v)
}

// MarshalWithEmpty marshals "v" into JSON. Empty fields with default values will be omitted.
func MarshalWithEmpty(v interface{}) ([]byte, error) {
	return pmWithEmpty.Marshal(v)
}

// MarshalFileWithEmpty marshals "v" into JSON and writes the result to a given file.
// Empty fields with default values will be omitted.
func MarshalFileWithEmpty(filename string, v interface{}) error {
	b, err := MarshalWithEmpty(v)
	if err != nil {
		return err
	}

	fd, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fd.Close()

	if _, err := fd.Write(b); err != nil {
		return err
	}

	return nil
}

// UnmarshalWithEmpty unmarshals JSON-encoded data into the non-nil value pointed to by "v".
// Unknown fields and enum name values are ignored.
func UnmarshalWithEmpty(data []byte, v interface{}) error {
	return pmWithEmpty.Unmarshal(data, v)
}

// UnmarshalFileWithEmpty reads JSON-encoded data from a given file and unmarshals it
// into the non-nil value pointed to by "v".
// Unknown fields and enum name values are ignored.
func UnmarshalFileWithEmpty(filename string, v interface{}) error {
	b, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return UnmarshalWithEmpty(b, v)
}
