package workflow

import (
	"testing"

	"github.com/stretchr/testify/require"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	"netbox-go/internal/domain/identity"
)

func TestInterfaceRPCDispatchUsesTypedService(t *testing.T) {
	typed := &grpcInterfaceServiceSpy{}
	services := completeTypedDCIMTestServices()
	services.interfaces = typed
	server := services.server()
	ctx := identity.WithPrincipal(
		t.Context(), identity.Principal{ID: 1, Username: "typed-interface"},
	)
	_, err := server.ListInterfaces(ctx, &dcimv1.ListInterfacesRequest{})
	require.NoError(t, err)
}
