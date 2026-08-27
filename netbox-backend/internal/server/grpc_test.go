package server

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"netbox-go/internal/platform/composition"
	"netbox-go/internal/platform/readiness"
)

func TestRuntimeGRPCServerRegistersOnlyCanonicalAndOperationalServices(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	server := grpc.NewServer()
	registerRuntimeServices(server, composition.NewCore(db), serverReadyChecker())
	services := server.GetServiceInfo()
	want := []string{
		"grpc.health.v1.Health",
		"grpc.reflection.v1.ServerReflection",
		"grpc.reflection.v1alpha.ServerReflection",
		"netbox.dcim.v1.DCIMService",
		"netbox.identity.v1.IdentityService",
		"netbox.ipam.v1.IPAMService",
	}
	if len(services) != len(want) {
		t.Fatalf("services = %v, want %v", services, want)
	}
	for _, name := range want {
		if _, ok := services[name]; !ok {
			t.Errorf("canonical service %q is not registered", name)
		}
	}
}

func TestRuntimeGRPCHealthTracksPostgreSQLReadiness(t *testing.T) {
	checker := &serverScriptedReadiness{results: []error{
		nil,
		errors.New("private PostgreSQL endpoint failure"),
		nil,
	}}
	healthServer := newReadinessHealthServer(checker)
	want := []healthv1.HealthCheckResponse_ServingStatus{
		healthv1.HealthCheckResponse_SERVING,
		healthv1.HealthCheckResponse_NOT_SERVING,
		healthv1.HealthCheckResponse_SERVING,
	}
	for index, wantStatus := range want {
		response, err := healthServer.Check(t.Context(), &healthv1.HealthCheckRequest{})
		require.NoError(t, err, "Check call %d", index+1)
		require.Equal(t, wantStatus, response.GetStatus(), "Check call %d", index+1)
	}
	require.Equal(t, 3, checker.Calls())
}

func TestRuntimeGRPCHealthRejectsNamedServicesWithoutProbing(t *testing.T) {
	checker := &serverScriptedReadiness{results: []error{nil}}
	healthServer := newReadinessHealthServer(checker)

	response, err := healthServer.Check(t.Context(), &healthv1.HealthCheckRequest{
		Service: "netbox.dcim.v1.DCIMService",
	})
	require.Nil(t, response)
	require.Equal(t, codes.NotFound, status.Code(err))
	require.Equal(t, 0, checker.Calls())
}

func TestRuntimeGRPCHealthWatchIsUnimplemented(t *testing.T) {
	checker := &serverScriptedReadiness{results: []error{nil}}
	healthServer := newReadinessHealthServer(checker)

	err := healthServer.Watch(&healthv1.HealthCheckRequest{}, nil)
	require.Equal(t, codes.Unimplemented, status.Code(err))
	require.Equal(t, 0, checker.Calls())
}

func TestRuntimeServerConstructorsRejectMissingReadinessChecker(t *testing.T) {
	require.PanicsWithValue(t, "HTTP server requires a readiness checker", func() {
		NewHTTPServer("127.0.0.1:0", nil)
	})
	require.PanicsWithValue(t, "gRPC server requires a readiness checker", func() {
		NewGRPCServer("127.0.0.1:0", nil)
	})
	require.PanicsWithValue(t, "gRPC health requires a readiness checker", func() {
		newReadinessHealthServer(nil)
	})
}

type serverScriptedReadiness struct {
	mu      sync.Mutex
	results []error
	calls   int
}

func (checker *serverScriptedReadiness) Check(context.Context) error {
	checker.mu.Lock()
	defer checker.mu.Unlock()
	if checker.calls >= len(checker.results) {
		return readiness.ErrUnavailable
	}
	result := checker.results[checker.calls]
	checker.calls++
	return result
}

func (checker *serverScriptedReadiness) Calls() int {
	checker.mu.Lock()
	defer checker.mu.Unlock()
	return checker.calls
}

type serverCheckerFunc func(context.Context) error

func (check serverCheckerFunc) Check(ctx context.Context) error { return check(ctx) }

func serverReadyChecker() readiness.Checker {
	return serverCheckerFunc(func(context.Context) error { return nil })
}
