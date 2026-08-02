package workflow

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewIPAMServerRequiresEveryTypedService(t *testing.T) {
	vrfs := &grpcVRFServiceSpy{}
	prefixes := &grpcPrefixServiceSpy{}
	addresses := &grpcIPAddressServiceSpy{}

	require.Panics(t, func() {
		NewIPAMServer(nil, prefixes, addresses)
	})
	require.Panics(t, func() {
		NewIPAMServer(vrfs, nil, addresses)
	})
	require.Panics(t, func() {
		NewIPAMServer(vrfs, prefixes, nil)
	})
	require.NotNil(t, NewIPAMServer(vrfs, prefixes, addresses))
}
