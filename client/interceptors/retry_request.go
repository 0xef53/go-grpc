package interceptors

import (
	"context"
	"time"

	"github.com/0xef53/go-grpc/utils"

	"google.golang.org/grpc"
	grpc_codes "google.golang.org/grpc/codes"
	grpc_status "google.golang.org/grpc/status"

	log "github.com/sirupsen/logrus"
)

// WithRequestsRetries returns an unary client interceptor that retries a request
// that fail due to temporary failures (such as network problems or service unavailability).
// It performs up to maxAttempts retries with a delay in seconds between attempts.
func WithRequestsRetries(maxAttempts int, delay time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req interface{}, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		logger := log.WithField("request.uid", utils.ExtractRequestID(ctx))

		var err error

		for attempt := 0; attempt < maxAttempts; attempt++ {
			err = invoker(ctx, method, req, reply, cc, opts...)

			if grpc_status.Code(err) == grpc_codes.Unavailable {
				logger.Warnf("Failed to perform request (attempt = %d): %s, %s", attempt, err, ctx.Err())
				logger.Warnf("Next try after %d seconds", delay)

				time.Sleep(delay * time.Second)

				continue
			}

			break
		}

		return err
	}
}
