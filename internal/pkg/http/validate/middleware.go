package validate

import (
	"context"
	nethttp "net/http"

	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	http2 "origadmin/application/origstudio/internal/pkg/http"
)

// ---------------------------------------------------------------------------
// Kratos gRPC server interceptor: validates the incoming request protobuf
// before passing control to the business handler. Returns a gRPC status with
// code=InvalidArgument (3) so that the gateway normalises it into HTTP 400.
// ---------------------------------------------------------------------------

// ServerUnaryInterceptor returns a gRPC unary server interceptor that runs
// validate.Validate on the single incoming request message. Errors are
// reported as gRPC status codes:
//
//	*ErrValidation   → codes.InvalidArgument  (gateway → HTTP 400)
//	other            → passed through unchanged
//
// Messages that do not implement Validatable/ValidatableFirst are skipped
// silently (no-op) so progressive migration is possible.
func ServerUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := Validate(req); err != nil {
			return nil, invalidArgumentStatus(err)
		}
		return handler(ctx, req)
	}
}

// ServerStreamInterceptor mirrors the unary variant for server-streaming RPCs.
// While most of our RPCs are unary, installing this interceptor prevents the
// stream path from silently bypassing validation for future streaming APIs.
func ServerStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, &validatingServerStream{ServerStream: ss})
	}
}

// validatingServerStream wraps the incoming stream so that each RecvMsg()
// immediately validates the decoded proto before the handler inspects it.
// For client-streaming this means the first (and only, in our current APIs)
// message still gets validated before the business logic runs.
type validatingServerStream struct {
	grpc.ServerStream
}

func (s *validatingServerStream) RecvMsg(m interface{}) error {
	if err := s.ServerStream.RecvMsg(m); err != nil {
		return err
	}
	if err := Validate(m); err != nil {
		return invalidArgumentStatus(err)
	}
	return nil
}

// invalidArgumentStatus wraps a validate error into gRPC status so that
// kratos HTTP gateway → frontend receives a 400 Bad Request with a readable
// message.
func invalidArgumentStatus(err error) error {
	return status.Error(3 /* codes.InvalidArgument */, err.Error())
}

// ---------------------------------------------------------------------------
// Kratos HTTP Filter: validates request bodies that were decoded via the
// kratos DecodeRequest hook by intercepting the handler call. This filter
// runs as a Kratos HTTP transport Filter, mirroring the BindJSON-in-adapter
// validation path for routes registered through the framework-agnostic
// http2.Router adapter.
// ---------------------------------------------------------------------------

// HTTPFilter is a pass-through. Validation for framework-agnostic http2.Router
// routes happens inside the BindJSON step of each adapter (gin/kratos/std).
// For native kratos DecodeRequest routes the caller can install an explicit
// decode hook. Kept exported so that future code paths have a uniform name.
func HTTPFilter(next nethttp.Handler) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		next.ServeHTTP(w, r)
	})
}

// KratosHTTPMiddleware returns an http2.MiddlewareFunc that can be applied to
// any http2.Router group. It acts as a no-op at the middleware layer because
// the actual validation is triggered inside each adapter's BindJSON method.
// Exported so that callers can explicitly register the "validator" on a
// per-group basis (making the validation intent visible at route setup).
func KratosHTTPMiddleware() http2.MiddlewareFunc {
	return func(next http2.HandlerFunc) http2.HandlerFunc {
		return func(ctx http2.Context) error {
			return next(ctx)
		}
	}
}

// ---------------------------------------------------------------------------
// Helpers: install onto Kratos transport options
// ---------------------------------------------------------------------------

// GRPCServerOptions returns a slice of native kratos grpc.ServerOption that
// installs the validate unary + streaming interceptors.
//
// Example:
//
//	opts := &grpc.ServerOptions{
//	    ServerOptions: validate.GRPCServerOptions(),
//	}
//	srv, err := grpc.NewServer(cfg, opts)
func GRPCServerOptions() []kgrpc.ServerOption {
	return []kgrpc.ServerOption{
		kgrpc.UnaryInterceptor(ServerUnaryInterceptor()),
		kgrpc.StreamInterceptor(ServerStreamInterceptor()),
	}
}
