package initial

import (
	"context"
	"testing"

	"github.com/go-dev-frame/sponge/pkg/app"
	"github.com/stretchr/testify/require"

	"netbox-go/internal/config"
	runtimeconfig "netbox-go/internal/platform/config"
	"netbox-go/internal/platform/readiness"
	"netbox-go/internal/server"
)

func TestRuntimeReadinessConstructionFailsClosedWithoutChecker(t *testing.T) {
	config.Set(&config.Config{})
	checker := &initialReadinessChecker{}
	var httpChecker readiness.Checker
	var grpcChecker readiness.Checker

	servers := createServices(
		runtimeconfig.HTTPRuntime{},
		checker,
		func(_ string, observed readiness.Checker, _ ...server.HTTPOption) app.IServer {
			httpChecker = observed
			return initialStubServer("http")
		},
		func(_ string, observed readiness.Checker, _ ...server.GrpcOption) app.IServer {
			grpcChecker = observed
			return initialStubServer("grpc")
		},
	)

	require.Len(t, servers, 2)
	require.Same(t, checker, httpChecker)
	require.Same(t, checker, grpcChecker)
	require.Same(t, httpChecker, grpcChecker)

	require.PanicsWithValue(t, "service composition requires a readiness checker", func() {
		createServices(
			runtimeconfig.HTTPRuntime{},
			nil,
			func(string, readiness.Checker, ...server.HTTPOption) app.IServer { return initialStubServer("http") },
			func(string, readiness.Checker, ...server.GrpcOption) app.IServer { return initialStubServer("grpc") },
		)
	})
	require.PanicsWithValue(t, "HTTP server requires a readiness checker", func() {
		server.NewHTTPServer("127.0.0.1:0", nil)
	})
	require.PanicsWithValue(t, "gRPC server requires a readiness checker", func() {
		server.NewGRPCServer("127.0.0.1:0", nil)
	})
}

type initialReadinessChecker struct{}

func (*initialReadinessChecker) Check(context.Context) error { return nil }

type initialStubServer string

func (initialStubServer) Start() error          { return nil }
func (initialStubServer) Stop() error           { return nil }
func (server initialStubServer) String() string { return string(server) }

var _ app.IServer = initialStubServer("")
