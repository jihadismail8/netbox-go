// Package server is a package that holds the http or grpc service.
package server

import (
	"context"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"github.com/go-dev-frame/sponge/pkg/app"
	"github.com/go-dev-frame/sponge/pkg/grpc/gtls"
	"github.com/go-dev-frame/sponge/pkg/grpc/interceptor"
	"github.com/go-dev-frame/sponge/pkg/grpc/metrics"
	"github.com/go-dev-frame/sponge/pkg/logger"
	"github.com/go-dev-frame/sponge/pkg/servicerd/registry"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	identityv1 "netbox-go/gen/go/netbox/identity/v1"
	ipamv1 "netbox-go/gen/go/netbox/ipam/v1"
	identitygrpc "netbox-go/internal/adapters/grpc/identity"
	workflowgrpc "netbox-go/internal/adapters/grpc/workflow"
	identityapp "netbox-go/internal/application/identity"
	"netbox-go/internal/config"
	"netbox-go/internal/database"
	"netbox-go/internal/ecode"
	"netbox-go/internal/platform/composition"
	"netbox-go/internal/platform/readiness"
)

var _ app.IServer = (*grpcServer)(nil)

type grpcServer struct {
	addr   string
	server *grpc.Server
	listen net.Listener

	iRegistry registry.Registry
	instance  *registry.ServiceInstance
	identity  *identityapp.Service
}

// Start grpc service
func (s *grpcServer) Start() error {
	// registration Services
	if s.iRegistry != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.iRegistry.Register(ctx, s.instance); err != nil {
			return err
		}
	}

	listen := metrics.NewCustomListener(s.listen, metrics.WithConnectionsLogger(logger.Get()), metrics.WithConnectionsGauge())
	if err := s.server.Serve(listen); err != nil { // block
		return err
	}

	return nil
}

// Stop grpc service
func (s *grpcServer) Stop() error {
	if s.iRegistry != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = s.iRegistry.Deregister(ctx, s.instance)
	}

	s.server.GracefulStop()

	return nil
}

// String comment
func (s *grpcServer) String() string {
	return "grpc service address " + s.addr
}

// secure option
func (s *grpcServer) secureServerOption() grpc.ServerOption {
	switch config.Get().Grpc.ServerSecure.Type {
	case "one-way": // server side certification
		credentials, err := gtls.GetServerTLSCredentials(
			config.Get().Grpc.ServerSecure.CertFile,
			config.Get().Grpc.ServerSecure.KeyFile,
		)
		if err != nil {
			panic(err)
		}
		logger.Info("grpc security type: sever-side certification")
		return grpc.Creds(credentials)

	case "two-way": // both client and server side certification
		credentials, err := gtls.GetServerTLSCredentialsByCA(
			config.Get().Grpc.ServerSecure.CaFile,
			config.Get().Grpc.ServerSecure.CertFile,
			config.Get().Grpc.ServerSecure.KeyFile,
		)
		if err != nil {
			panic(err)
		}
		logger.Info("grpc security type: both client-side and server-side certification")
		return grpc.Creds(credentials)
	}

	logger.Info("grpc security type: insecure")
	return nil
}

// setting up unary server interceptors
func (s *grpcServer) unaryServerOptions() grpc.ServerOption {
	unaryServerInterceptors := []grpc.UnaryServerInterceptor{
		interceptor.UnaryServerRecovery(),
		interceptor.UnaryServerRequestID(),
		identitygrpc.UnaryAuthenticator(s.identity),
	}

	// logger interceptor, to print simple messages, replace interceptor.UnaryServerLog with interceptor.UnaryServerSimpleLog
	unaryServerInterceptors = append(unaryServerInterceptors, interceptor.UnaryServerLog(
		logger.Get(),
		interceptor.WithReplaceGRPCLogger(),
	))

	// jwt token interceptor
	//unaryServerInterceptors = append(unaryServerInterceptors, interceptor.UnaryServerJwtAuth(
	// // choose a verification method as needed
	//interceptor.WithStandardVerify(standardVerifyFn), // standard verify (default), you can set standardVerifyFn to nil if you don't need it
	//interceptor.WithCustomVerify(customVerifyFn), // custom verify
	// // specify the grpc API to ignore token verification(full path)
	//interceptor.WithAuthIgnoreMethods("/api.user.v1.User/Register", "/api.user.v1.User/Login"),
	//))

	// metrics interceptor
	if config.Get().App.EnableMetrics {
		unaryServerInterceptors = append(unaryServerInterceptors, interceptor.UnaryServerMetrics())
	}

	// limit interceptor
	if config.Get().App.EnableLimit {
		unaryServerInterceptors = append(unaryServerInterceptors, interceptor.UnaryServerRateLimit(
		//interceptor.WithWindow(time.Second*5), // default 10s
		//interceptor.WithBucket(200),           // default 100
		//interceptor.WithCPUThreshold(900),     // default 800
		))
	}

	// circuit breaker interceptor
	if config.Get().App.EnableCircuitBreaker {
		unaryServerInterceptors = append(unaryServerInterceptors, interceptor.UnaryServerCircuitBreaker(
			//interceptor.WithBreakerOption(
			//circuitbreaker.WithSuccess(75),           // default 60
			//circuitbreaker.WithRequest(100),          // default 100
			//circuitbreaker.WithBucket(10),            // default 10
			//circuitbreaker.WithWindow(time.Second*3), // default 3s
			//),
			//interceptor.WithUnaryServerDegradeHandler(handler), // Add degradation processing
			interceptor.WithValidCode( // Add timeout error codes can also trigger circuit breaking
				ecode.StatusInternalServerError.Code(),
				ecode.StatusServiceUnavailable.Code(),
			),
		))
	}

	// trace interceptor
	if config.Get().App.EnableTrace {
		unaryServerInterceptors = append(unaryServerInterceptors, interceptor.UnaryServerTracing())
	}

	return grpc.ChainUnaryInterceptor(unaryServerInterceptors...)
}

// setting up stream server interceptors
func (s *grpcServer) streamServerOptions() grpc.ServerOption {
	streamServerInterceptors := []grpc.StreamServerInterceptor{
		interceptor.StreamServerRecovery(),
		//interceptor.StreamServerRequestID(),
	}

	// logger interceptor, to print simple messages, replace interceptor.StreamServerLog with interceptor.StreamServerSimpleLog
	streamServerInterceptors = append(streamServerInterceptors, interceptor.StreamServerLog(
		logger.Get(),
		interceptor.WithReplaceGRPCLogger(),
	))

	// jwt token interceptor
	//streamServerInterceptors = append(streamServerInterceptors, interceptor.StreamServerJwtAuth(
	// // choose a verification method as needed
	//interceptor.WithStandardVerify(standardVerifyFn), // standard verify (default), you can set standardVerifyFn to nil if you don't need it
	//interceptor.WithCustomVerify(customVerifyFn), // custom verify
	// // specify the grpc API to ignore token verification(full path)
	//	interceptor.WithAuthIgnoreMethods("/api.user.v1.User/Register", "/api.user.v1.User/Login"),
	//))

	// metrics interceptor
	if config.Get().App.EnableMetrics {
		streamServerInterceptors = append(streamServerInterceptors, interceptor.StreamServerMetrics())
	}

	// limit interceptor
	if config.Get().App.EnableLimit {
		streamServerInterceptors = append(streamServerInterceptors, interceptor.StreamServerRateLimit())
	}

	// circuit breaker interceptor
	if config.Get().App.EnableCircuitBreaker {
		streamServerInterceptors = append(streamServerInterceptors, interceptor.StreamServerCircuitBreaker(
			// set rpc code for circuit breaker, default already includes codes.Internal and codes.Unavailable
			interceptor.WithValidCode(ecode.StatusInternalServerError.Code()),
			interceptor.WithValidCode(ecode.StatusServiceUnavailable.Code()),
		))
	}

	// trace interceptor
	if config.Get().App.EnableTrace {
		streamServerInterceptors = append(streamServerInterceptors, interceptor.StreamServerTracing())
	}

	return grpc.ChainStreamInterceptor(streamServerInterceptors...)
}

func (s *grpcServer) setOptions() []grpc.ServerOption {
	var options []grpc.ServerOption

	secureOption := s.secureServerOption()
	if secureOption != nil {
		options = append(options, secureOption)
	}

	options = append(options, s.unaryServerOptions())
	options = append(options, s.streamServerOptions())

	return options
}

// NewGRPCServer creates a new grpc server
func NewGRPCServer(addr string, readinessChecker readiness.Checker, opts ...GrpcOption) app.IServer {
	if readinessChecker == nil {
		panic("gRPC server requires a readiness checker")
	}
	var err error
	o := defaultGrpcOptions()
	o.apply(opts...)
	s := &grpcServer{
		addr:      addr,
		iRegistry: o.iRegistry,
		instance:  o.instance,
	}
	core := composition.NewCore(database.GetDB())
	s.identity = core.Identity
	s.listen, err = net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	s.server = grpc.NewServer(s.setOptions()...)
	registerRuntimeServices(s.server, core, readinessChecker)
	return s
}

// registerRuntimeServices publishes the bounded application contract together
// with the standard operational services required by gRPC clients. Reflection
// is intentionally registered only after the legacy generated services have
// been excluded, so discovery cannot advertise unsupported capabilities.
func registerRuntimeServices(server *grpc.Server, core composition.Core, readinessChecker readiness.Checker) {
	registerCanonicalServices(server, core)
	healthv1.RegisterHealthServer(server, newReadinessHealthServer(readinessChecker))
	reflection.Register(server)
}

type readinessHealthServer struct {
	healthv1.UnimplementedHealthServer
	checker readiness.Checker
}

func newReadinessHealthServer(checker readiness.Checker) *readinessHealthServer {
	if checker == nil {
		panic("gRPC health requires a readiness checker")
	}
	return &readinessHealthServer{checker: checker}
}

func (server *readinessHealthServer) Check(
	ctx context.Context,
	request *healthv1.HealthCheckRequest,
) (*healthv1.HealthCheckResponse, error) {
	if request.GetService() != "" {
		return nil, status.Error(codes.NotFound, "unknown service")
	}
	servingStatus := healthv1.HealthCheckResponse_SERVING
	if err := server.checker.Check(ctx); err != nil {
		servingStatus = healthv1.HealthCheckResponse_NOT_SERVING
	}
	return &healthv1.HealthCheckResponse{Status: servingStatus}, nil
}

func (*readinessHealthServer) Watch(
	*healthv1.HealthCheckRequest,
	grpc.ServerStreamingServer[healthv1.HealthCheckResponse],
) error {
	return status.Error(codes.Unimplemented, "health watch is not supported")
}

func registerCanonicalServices(server *grpc.Server, core composition.Core) {
	identityv1.RegisterIdentityServiceServer(server, identitygrpc.NewServer(core.Identity))
	dcimv1.RegisterDCIMServiceServer(
		server,
		workflowgrpc.NewTypedDCIMServer(
			core.Sites,
			core.Manufacturers,
			core.RackRoles,
			core.RackTypes,
			core.Racks,
			core.DeviceRoles,
			core.DeviceTypes,
			core.InterfaceTemplates,
			core.Devices,
			core.Interfaces,
		),
	)
	ipamv1.RegisterIPAMServiceServer(
		server,
		workflowgrpc.NewIPAMServer(
			core.VRFs,
			core.Prefixes,
			core.IPAddresses,
		),
	)
}
