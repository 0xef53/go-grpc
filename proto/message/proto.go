package message

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/0xef53/go-grpc/options"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

var (
	ErrUnsupportedFormat = errors.New("<unsupported format>")
)

// LoggingShow returns tags of [protoreflect.Value] and all nested elements, represented as a string.
func LoggingShow(fd protoreflect.FieldDescriptor, v protoreflect.Value) map[string]interface{} {
	m := make(map[string]interface{})
	name := fd.TextName()

	switch {
	case fd.IsList():
		for idx := 0; idx < v.List().Len(); idx++ {
			for k, v := range TagsFromValue(fmt.Sprintf("%s.[%d]", name, idx), fd.Kind(), v.List().Get(idx)) {
				m[k] = v
			}
		}
	case fd.IsMap():
		v.Map().Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
			for mk, mv := range TagsFromValue(fmt.Sprintf("%s.[%s]", name, k.String()), fd.MapValue().Kind(), v) {
				m[mk] = mv
			}
			return true
		})
	default:
		m = TagsFromValue(name, fd.Kind(), v)
	}

	return m
}

// LoggingHide returns an empty tags for the hidden field.
func LoggingHide(fd protoreflect.FieldDescriptor, v protoreflect.Value) map[string]interface{} {
	return nil
}

// LoggingObfuscate returns a map where all nested tags of [protoreflect.Value] are masked
// with a given "replacement". If no replacement is specified, the standard mask "*****" is used.
func LoggingObfuscate(fd protoreflect.FieldDescriptor, v protoreflect.Value, replacement string) map[string]interface{} {
	if len(replacement) == 0 {
		replacement = "*****"
	}

	return map[string]interface{}{
		fd.TextName(): fmt.Sprintf("(%s) %s", fd.Kind(), replacement),
	}
}

// LoggingTrimHead returns tags of [protoreflect.Value] trimmed to "tail" elements from the end.
//
// If "tail" is zero, the function behaves like LoggingHide(). If "tail" is greater than or
// equal to the number of elements in [protoreflect.Value], then the function behaves like LoggingShow.
//
// Only string values or lists are supported. For other types, an error ErrUnsupportedFormat is returned.
func LoggingTrimHead(fd protoreflect.FieldDescriptor, v protoreflect.Value, tail int64) map[string]interface{} {
	if tail == 0 {
		return LoggingHide(fd, v)
	}

	m := make(map[string]interface{})

	name := fd.TextName()

	switch {
	case fd.IsList():
		if int(tail) >= v.List().Len() {
			return LoggingShow(fd, v)
		}

		for idx := v.List().Len() - int(tail); idx < v.List().Len(); idx++ {
			for k, v := range TagsFromValue(fmt.Sprintf("%s.[%d]", name, idx), fd.Kind(), v.List().Get(idx)) {
				m[k] = v
			}
		}
	case fd.Kind() == protoreflect.StringKind:
		vs := v.String()

		if len(vs) <= int(tail) {
			m[name] = v
		} else {
			m[name] = "<...>" + vs[len(vs)-int(tail):]
		}
	default:
		m[name] = ErrUnsupportedFormat.Error()
	}

	return m
}

// LoggingTrimTail returns tags of [protoreflect.Value] trimmed to "head" elements from the start.
//
// If "head" is zero, the function behaves like LoggingHide(). If "head" is greater than or
// equal to the number of elements in [protoreflect.Value], then the function behaves like LoggingShow.
//
// Only string values or lists are supported. For other types, an error ErrUnsupportedFormat is returned.
func LoggingTrimTail(fd protoreflect.FieldDescriptor, v protoreflect.Value, head int64) map[string]interface{} {
	m := make(map[string]interface{})

	name := fd.TextName()

	if head == 0 {
		return LoggingHide(fd, v)
	}

	switch {
	case fd.IsList():
		if int(head) > v.List().Len() {
			return LoggingShow(fd, v)
		}
		for idx := 0; idx < v.List().Len() && idx < int(head); idx++ {
			for k, v := range TagsFromValue(fmt.Sprintf("%s.[%d]", name, idx), fd.Kind(), v.List().Get(idx)) {
				m[k] = v
			}
		}
	case fd.Kind() == protoreflect.StringKind:
		vs := v.String()

		if len(vs) <= int(head) {
			m[name] = v
		} else {
			m[name] = vs[:head] + "<...>"
		}
	default:
		m[name] = ErrUnsupportedFormat.Error()
	}

	return m
}

// LoggingTrimMiddle returns tags of [protoreflect.Value] trimmed to "head" elements from the start
// and to "tail" elements from the end.
//
// If "head" is zero, the function behaves like LoggingTrimHead(). If "tail" is zero, the function behaves
// like LoggingTrimTail(). If both "head" and "tail" are zero, the function behaves like LoggingHide().
// If the sum of "head" and "tail" is greater than or equal to the number of elements in [protoreflect.Value],
// the function behaves like LoggingShow().
//
// Only string values or lists are supported. For other types, an error ErrUnsupportedFormat is returned.
func LoggingTrimMiddle(fd protoreflect.FieldDescriptor, v protoreflect.Value, head, tail int64) map[string]interface{} {
	m := make(map[string]interface{})

	name := fd.TextName()

	if int(tail+head) == 0 {
		return LoggingHide(fd, v)
	}

	switch {
	case fd.IsList():
		if int(tail+head) >= v.List().Len() {
			return LoggingShow(fd, v)
		}

		for k, v := range LoggingTrimHead(fd, v, tail) {
			m[k] = v
		}

		for k, v := range LoggingTrimTail(fd, v, head) {
			m[k] = v
		}
	case fd.Kind() == protoreflect.StringKind:
		vs := v.String()

		if int(tail+head) >= len(vs) {
			return LoggingShow(fd, v)
		}

		m[name] = vs[:head] + "<...>" + vs[len(vs)-int(tail):]
	default:
		m[name] = ErrUnsupportedFormat.Error()
	}

	return m
}

// TagsFromValue parses non-lists and non-maps values of type [protoreflect.Value] into tags.
func TagsFromValue(name string, kind protoreflect.Kind, v protoreflect.Value) map[string]interface{} {
	m := make(map[string]interface{})

	switch kind {
	case protoreflect.MessageKind:
		msg := v.Message()

		switch msg.Descriptor().FullName() {
		case "google.protobuf.Timestamp":
			m[name] = strings.Trim(protojson.Format(msg.Interface()), `"`)
		default:
			for k, v := range TagsFromMessage(msg) {
				m[fmt.Sprintf("%s.%s", name, k)] = v
			}
		}
	case protoreflect.GroupKind:
		// do nothing with groups
	case protoreflect.BytesKind:
		m[name] = base64.StdEncoding.EncodeToString(v.Bytes())
	default:
		// try to convert to a string
		m[name] = v.String()
	}

	return m
}

// TagsFromMessage parses [protoreflect.Message] into tags.
// Parsing options are defined by the field option of type [options.FieldLogging].
func TagsFromMessage(msg protoreflect.Message) (tags map[string]interface{}) {
	defer func() {
		if r := recover(); r != nil {
			tags = map[string]interface{}{
				"_xxx_tags_extracting_error": fmt.Sprintf("[UNEXPECTED] error when tags extracting: %v", r),
			}
		}
	}()

	tags = make(map[string]interface{})

	msg.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		opts := fd.Options().(*descriptorpb.FieldOptions)

		logf := proto.GetExtension(opts, options.E_LogFormatting).(*options.FieldLogging)
		if logf == nil {
			logf = &options.FieldLogging{}
		}

		switch logf.Display {
		case options.FieldLogging_Hide:
			for k, val := range LoggingHide(fd, v) {
				tags[k] = val
			}
		case options.FieldLogging_Obfuscate:
			for k, val := range LoggingObfuscate(fd, v, logf.Replacement) {
				tags[k] = val
			}
		case options.FieldLogging_TrimHead:
			for k, val := range LoggingTrimHead(fd, v, logf.TailChars) {
				tags[k] = val
			}
		case options.FieldLogging_TrimTail:
			for k, val := range LoggingTrimTail(fd, v, logf.HeadChars) {
				tags[k] = val
			}
		case options.FieldLogging_TrimMiddle:
			for k, val := range LoggingTrimMiddle(fd, v, logf.HeadChars, logf.TailChars) {
				tags[k] = val
			}
		default:
			for k, val := range LoggingShow(fd, v) {
				tags[k] = val
			}
		}

		return true
	})

	if len(tags) == 0 {
		return nil
	}

	return tags
}
