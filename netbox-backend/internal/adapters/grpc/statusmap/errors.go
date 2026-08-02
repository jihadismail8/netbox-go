// Package statusmap maps the canonical application error taxonomy to gRPC.
package statusmap

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"netbox-go/internal/domain/shared"
)

func Error(err error) error {
	if err == nil {
		return nil
	}
	var sharedErr *shared.Error
	if !errors.As(err, &sharedErr) {
		return status.Error(codes.Internal, "an internal error occurred")
	}
	code := map[shared.ErrorReason]codes.Code{
		shared.ErrorReasonValidation:      codes.InvalidArgument,
		shared.ErrorReasonUnauthenticated: codes.Unauthenticated,
		shared.ErrorReasonForbidden:       codes.PermissionDenied,
		shared.ErrorReasonNotFound:        codes.NotFound,
		shared.ErrorReasonConflict:        codes.AlreadyExists,
		shared.ErrorReasonProtected:       codes.FailedPrecondition,
		shared.ErrorReasonRateLimited:     codes.ResourceExhausted,
		shared.ErrorReasonInternal:        codes.Internal,
	}[sharedErr.Reason]
	if code == codes.OK {
		code = codes.Internal
	}
	return status.Error(code, sharedErr.Message)
}
