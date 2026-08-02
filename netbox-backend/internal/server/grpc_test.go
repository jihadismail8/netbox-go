package server

import (
	"testing"

	"google.golang.org/grpc"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"netbox-go/internal/platform/composition"
)

func TestRuntimeGRPCServerRegistersOnlyCanonicalAndOperationalServices(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	server := grpc.NewServer()
	registerRuntimeServices(server, composition.NewCore(db))
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
