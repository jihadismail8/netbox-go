package identity

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUnaryAuthenticatorAllowsStandardHealthCheckWithoutCredentials(t *testing.T) {
	interceptor := UnaryAuthenticator(nil)
	want := struct{}{}
	got, err := interceptor(
		context.Background(),
		struct{}{},
		&grpc.UnaryServerInfo{FullMethod: "/grpc.health.v1.Health/Check"},
		func(context.Context, any) (any, error) { return want, nil },
	)
	if err != nil {
		t.Fatalf("health check: %v", err)
	}
	if got != want {
		t.Fatalf("health result = %#v, want %#v", got, want)
	}
}

func TestUnaryAuthenticatorStillProtectsCanonicalMethods(t *testing.T) {
	interceptor := UnaryAuthenticator(nil)
	_, err := interceptor(
		context.Background(),
		struct{}{},
		&grpc.UnaryServerInfo{FullMethod: "/netbox.dcim.v1.DCIMService/ListSites"},
		func(context.Context, any) (any, error) { return struct{}{}, nil },
	)
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("status = %v, want %v (error %v)", status.Code(err), codes.Unauthenticated, err)
	}
}
