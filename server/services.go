package server

import (
	"sync"

	grpc_runtime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

var (
	defaultServiceBucket = "default"

	pool = struct {
		sync.Mutex
		services map[string][]Service
	}{
		services: map[string][]Service{
			defaultServiceBucket: make([]Service, 0),
		},
	}
)

type Service interface {
	Name() string

	RegisterGRPC(*grpc.Server)
	RegisterGW(*grpc_runtime.ServeMux, string, []grpc.DialOption)
}

// ServiceOption is a common interface type for optional parameters for [Service].
type ServiceOption interface{}

// BucketServiceOption is an option containing the name of some bucket.
type BucketServiceOption struct {
	bucket string
}

func WithServiceBucket(bucket string) ServiceOption {
	return &BucketServiceOption{
		bucket: bucket,
	}
}

// Register appends a given service to the common bucket with name "default".
// If additional bucket names are specified as [BucketServiceOption],
// the function will also add a given service to the specified buckets.
func Register(svc Service, options ...ServiceOption) {
	pool.Lock()
	defer pool.Unlock()

	for _, opt := range options {
		switch o := opt.(type) {
		case *BucketServiceOption:
			if _, ok := pool.services[o.bucket]; !ok {
				pool.services[o.bucket] = []Service{svc}
			} else {
				pool.services[o.bucket] = append(pool.services[o.bucket], svc)
			}
		}
	}

	pool.services[defaultServiceBucket] = append(pool.services[defaultServiceBucket], svc)
}

// Services returns a list of services associated with the given bucket names.
// If no buckets are provided, the function returns all registered services.
func Services(buckets ...string) []Service {
	pool.Lock()
	defer pool.Unlock()

	if len(buckets) == 0 {
		buckets = []string{defaultServiceBucket}
	}

	services := make([]Service, 0)

	for _, bucket := range buckets {
		if len(pool.services[bucket]) > 0 {
			services = append(services, pool.services[bucket]...)
		}
	}

	return services
}
