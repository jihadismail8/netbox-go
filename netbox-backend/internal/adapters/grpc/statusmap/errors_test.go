package statusmap

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"netbox-go/internal/domain/shared"
)

func TestTypedApplicationReasonsMapToCanonicalGRPCCodes(t *testing.T) {
	tests := map[shared.ErrorReason]codes.Code{
		shared.ErrorReasonValidation:      codes.InvalidArgument,
		shared.ErrorReasonUnauthenticated: codes.Unauthenticated,
		shared.ErrorReasonForbidden:       codes.PermissionDenied,
		shared.ErrorReasonNotFound:        codes.NotFound,
		shared.ErrorReasonConflict:        codes.AlreadyExists,
		shared.ErrorReasonProtected:       codes.FailedPrecondition,
		shared.ErrorReasonRateLimited:     codes.ResourceExhausted,
		shared.ErrorReasonInternal:        codes.Internal,
	}
	for reason, expected := range tests {
		t.Run(string(reason), func(t *testing.T) {
			mapped := Error(shared.NewError(reason, "typed failure"))
			require.Equal(t, expected, status.Code(mapped))
			require.Equal(t, "typed failure", status.Convert(mapped).Message())
		})
	}
}

func TestUnknownErrorsAreSanitized(t *testing.T) {
	t.Parallel()

	mapped := Error(errors.New("database details must stay private"))
	require.Equal(t, codes.Internal, status.Code(mapped))
	require.Equal(t, "an internal error occurred", status.Convert(mapped).Message())
}
