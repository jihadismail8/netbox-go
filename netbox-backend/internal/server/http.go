package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/go-dev-frame/sponge/pkg/app"
	"github.com/go-dev-frame/sponge/pkg/httpsrv"
	"github.com/go-dev-frame/sponge/pkg/servicerd/registry"

	runtimehttp "netbox-go/internal/adapters/rest/netbox/router"
	workflowhttp "netbox-go/internal/adapters/rest/netbox/workflow"
	"netbox-go/internal/config"
	"netbox-go/internal/database"
	"netbox-go/internal/platform/composition"
	"netbox-go/internal/platform/readiness"
)

var _ app.IServer = (*httpServer)(nil)

type httpServer struct {
	addr   string
	server *httpsrv.Server

	instance  *registry.ServiceInstance
	iRegistry registry.Registry
}

// Start http service
func (s *httpServer) Start() error {
	if s.iRegistry != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.iRegistry.Register(ctx, s.instance); err != nil {
			return err
		}
	}

	if err := s.server.Run(); err != nil {
		return fmt.Errorf("run %s service error: %v", s.server.Scheme(), err)
	}
	return nil
}

// Stop http service
func (s *httpServer) Stop() error {
	if s.iRegistry != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = s.iRegistry.Deregister(ctx, s.instance)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}

// String comment
func (s *httpServer) String() string {
	return s.server.Scheme() + " service address is " + s.addr
}

func newServer(server *http.Server, tls config.TLS) *httpsrv.Server {
	var c *httpsrv.Server
	switch httpsrv.Mode(tls.EnableMode) {
	case httpsrv.ModeTLSSelfSigned:
		c = httpsrv.New(server, httpsrv.NewTLSSelfSignedConfig())
	case httpsrv.ModeTLSEncrypt:
		c = httpsrv.New(server,
			httpsrv.NewTLSEAutoEncryptConfig(
				tls.Domain,
				tls.Email,
				// enable http redirect to https, port 80 to 443, default is false
				//httpsrv.WithTLSEncryptEnableRedirect(),
			),
		)
	case httpsrv.ModeTLSExternal:
		c = httpsrv.New(server, httpsrv.NewTLSExternalConfig(tls.CertFile, tls.KeyFile))
	default:
		c = httpsrv.New(server) // default is http, no tls
	}
	return c
}

// NewHTTPServer creates a new http server
func NewHTTPServer(addr string, readinessChecker readiness.Checker, opts ...HTTPOption) app.IServer {
	if readinessChecker == nil {
		panic("HTTP server requires a readiness checker")
	}
	o := defaultHTTPOptions()
	o.apply(opts...)

	if o.isProd {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	core := composition.NewCore(database.GetDB())
	secureCookies := o.isProd || o.tls.EnableMode != ""
	router := runtimehttp.New(
		core.Identity,
		core.Sites,
		secureCookies,
		o.corsAllowedOrigins,
		readinessChecker,
		workflowhttp.WithOrganizationServices(core.Manufacturers, core.RackRoles),
		workflowhttp.WithRackTypeService(core.RackTypes),
		workflowhttp.WithRackService(core.Racks),
		workflowhttp.WithDeviceRoleService(core.DeviceRoles),
		workflowhttp.WithDeviceTypeService(core.DeviceTypes),
		workflowhttp.WithInterfaceTemplateService(core.InterfaceTemplates),
		workflowhttp.WithDeviceService(core.Devices),
		workflowhttp.WithInterfaceService(core.Interfaces),
		workflowhttp.WithVRFService(core.VRFs),
		workflowhttp.WithPrefixService(core.Prefixes),
		workflowhttp.WithIPAddressService(core.IPAddresses),
	)
	server := &http.Server{
		Addr:    addr,
		Handler: router,
		//ReadTimeout:    time.Second*30,
		//WriteTimeout:   time.Second*60,
		IdleTimeout:    time.Second * 60,
		MaxHeaderBytes: 1 << 20,
	}

	return &httpServer{
		addr:      addr,
		server:    newServer(server, o.tls),
		iRegistry: o.iRegistry,
		instance:  o.instance,
	}
}
