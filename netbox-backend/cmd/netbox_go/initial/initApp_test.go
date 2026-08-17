package initial

import (
	"fmt"
	"strings"
	"testing"
)

func TestCORSConfigurationLoadsBeforeDatabaseInitialization(t *testing.T) {
	const secretLikeValue = "https://startup-user:CORS-STARTUP-SENTINEL@example.com"
	t.Setenv("NETBOX_CORS_ALLOWED_ORIGINS", secretLikeValue)

	databaseInitializationCalled := false
	var recovered any
	func() {
		defer func() {
			recovered = recover()
		}()
		initApp(func() {
			databaseInitializationCalled = true
		})
	}()

	if recovered == nil {
		t.Fatal("initApp() did not reject invalid CORS configuration")
	}
	if databaseInitializationCalled {
		t.Fatal("database initializer ran before invalid CORS configuration was rejected")
	}

	message := fmt.Sprint(recovered)
	if !strings.Contains(message, "NETBOX_CORS_ALLOWED_ORIGINS[0]: user information is not allowed") {
		t.Fatalf("panic = %q, want safe CORS configuration reason", message)
	}
	for _, forbidden := range []string{
		secretLikeValue,
		"startup-user",
		"CORS-STARTUP-SENTINEL",
		"SENTINEL",
		"example.com",
	} {
		if strings.Contains(message, forbidden) {
			t.Fatalf("panic %q disclosed forbidden input fragment %q", message, forbidden)
		}
	}
}
