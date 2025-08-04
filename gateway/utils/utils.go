package utils

import (
	"strings"

	grpc_runtime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/protobuf/encoding/protojson"
)

// NewGatewayMux returns a new gRPC-Gateway ServeMux with custom serialization
// and header handling.
//
// Among other things, it forwards all headers starting with "X-"
// by converting them to lowercase and removing the "X-" prefix.
func NewGatewayMux() *grpc_runtime.ServeMux {
	gwMux := grpc_runtime.NewServeMux(
		grpc_runtime.WithMarshalerOption(
			grpc_runtime.MIMEWildcard,
			&grpc_runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					UseProtoNames:   true,
					EmitUnpopulated: true,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{},
			},
		),
		// Forward all X-Headers
		grpc_runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
			if strings.HasPrefix(key, "X-") {
				return strings.ToLower(strings.TrimPrefix(key, "X-")), true
			}
			return grpc_runtime.DefaultHeaderMatcher(key)
		}),
	)

	return gwMux
}
