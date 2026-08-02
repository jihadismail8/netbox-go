package workflow

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	changelogpostgres "netbox-go/internal/adapters/postgres/changelog"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	ipamrow "netbox-go/internal/adapters/postgres/ipam/row"
	applicationdcim "netbox-go/internal/application/dcim"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
	"netbox-go/internal/platform/composition"
)

func TestCreateOmitsQuerysetCountsWhileRetrieveAndUpdateIncludeThem(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:rest_annotation_shape?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(profileRowModels()...); err != nil {
		t.Fatal(err)
	}
	core := composition.NewCore(db)
	principal := identity.Principal{ID: 1, Username: "shape-test", IsSuperuser: true}
	router := gin.New()
	newTypedCoreHandler(core).Register(router, func(c *gin.Context) {
		SetPrincipal(c, principal)
		c.Next()
	})

	created := workflowRequest(t, router, http.MethodPost, "/api/dcim/sites/", `{"name":"Shape Site","slug":"shape-site"}`, http.StatusCreated)
	for _, field := range []string{"device_count", "prefix_count", "rack_count"} {
		if _, exists := created[field]; exists {
			t.Errorf("POST response unexpectedly contains queryset annotation %q", field)
		}
	}
	id := int64(created["id"].(float64))

	retrieved := workflowRequest(t, router, http.MethodGet, "/api/dcim/sites/"+formatID(id)+"/", "", http.StatusOK)
	for _, field := range []string{"device_count", "prefix_count", "rack_count"} {
		if value, exists := retrieved[field]; !exists || value != float64(0) {
			t.Errorf("GET annotation %q = %#v (present %v), want 0", field, value, exists)
		}
	}
	updated := workflowRequest(t, router, http.MethodPatch, "/api/dcim/sites/"+formatID(id)+"/", `{"description":"updated"}`, http.StatusOK)
	for _, field := range []string{"device_count", "prefix_count", "rack_count"} {
		if _, exists := updated[field]; !exists {
			t.Errorf("PATCH response omitted queryset annotation %q", field)
		}
	}

	manufacturer := workflowRequest(t, router, http.MethodPost, "/api/dcim/manufacturers/", `{"name":"Shape Vendor","slug":"shape-vendor"}`, http.StatusCreated)
	manufacturerID := int64(manufacturer["id"].(float64))
	deviceType := workflowRequest(t, router, http.MethodPost, "/api/dcim/device-types/", `{"manufacturer":`+formatID(manufacturerID)+`,"model":"Shape Router","slug":"shape-router"}`, http.StatusCreated)
	deviceTypeID := int64(deviceType["id"].(float64))
	workflowRequest(t, router, http.MethodPost, "/api/dcim/interface-templates/", `{"device_type":`+formatID(deviceTypeID)+`,"name":"eth0","type":"1000base-t"}`, http.StatusCreated)
	role := workflowRequest(t, router, http.MethodPost, "/api/dcim/device-roles/", `{"name":"Shape Role","slug":"shape-role"}`, http.StatusCreated)
	roleID := int64(role["id"].(float64))
	device := workflowRequest(t, router, http.MethodPost, "/api/dcim/devices/", `{"device_type":`+formatID(deviceTypeID)+`,"role":`+formatID(roleID)+`,"site":`+formatID(id)+`,"name":"shape-device"}`, http.StatusCreated)
	if value := device["interface_count"]; value != float64(0) {
		t.Errorf("Device POST interface_count = %#v, want stale baseline value 0", value)
	}
	deviceID := int64(device["id"].(float64))
	device = workflowRequest(t, router, http.MethodGet, "/api/dcim/devices/"+formatID(deviceID)+"/", "", http.StatusOK)
	if value := device["interface_count"]; value != float64(1) {
		t.Errorf("Device GET interface_count = %#v, want durable value 1", value)
	}
}

func TestPreviousPageURLUsesDRFDefaultOffsetShape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:rest_page_url_shape?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(profileRowModels()...); err != nil {
		t.Fatal(err)
	}
	core := composition.NewCore(db)
	principal := identity.Principal{ID: 1, Username: "page-test", IsSuperuser: true}
	router := gin.New()
	newTypedCoreHandler(core).Register(router, func(c *gin.Context) {
		SetPrincipal(c, principal)
		c.Next()
	})
	for index := 1; index <= 3; index++ {
		workflowRequest(
			t,
			router,
			http.MethodPost,
			"/api/dcim/sites/",
			`{"name":"Page Site `+strconv.Itoa(index)+`","slug":"page-site-`+strconv.Itoa(index)+`"}`,
			http.StatusCreated,
		)
	}
	page := workflowRequest(t, router, http.MethodGet, "/api/dcim/sites/?limit=2&offset=2", "", http.StatusOK)
	if got, want := page["previous"], "/api/dcim/sites/?limit=2"; got != want {
		t.Fatalf("previous page URL = %#v, want %#v", got, want)
	}
}

func TestReadResponsesUseNetBoxChoiceEnvelopes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:rest_choice_envelopes?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(profileRowModels()...); err != nil {
		t.Fatal(err)
	}
	core := composition.NewCore(db)
	principal := identity.Principal{ID: 1, Username: "choice-test", IsSuperuser: true}
	router := gin.New()
	newTypedCoreHandler(core).Register(router, func(c *gin.Context) {
		SetPrincipal(c, principal)
		c.Next()
	})

	site := workflowRequest(t, router, http.MethodPost, "/api/dcim/sites/", `{"name":"Choice Site","slug":"choice-site","status":"planned"}`, http.StatusCreated)
	assertChoice(t, site, "status", "planned", "Planned")
	siteID := int64(site["id"].(float64))

	manufacturer := workflowRequest(t, router, http.MethodPost, "/api/dcim/manufacturers/", `{"name":"Choice Vendor","slug":"choice-vendor"}`, http.StatusCreated)
	manufacturerID := int64(manufacturer["id"].(float64))
	rackType := workflowRequest(t, router, http.MethodPost, "/api/dcim/rack-types/", `{"manufacturer":`+formatID(manufacturerID)+`,"model":"Choice Rack","slug":"choice-rack","form_factor":"4-post-cabinet","width":21}`, http.StatusCreated)
	assertChoice(t, rackType, "form_factor", "4-post-cabinet", "4-post cabinet")
	assertChoice(t, rackType, "width", float64(21), "21 inches")

	rack := workflowRequest(t, router, http.MethodPost, "/api/dcim/racks/", `{"site":`+formatID(siteID)+`,"name":"Choice Rack","airflow":null}`, http.StatusCreated)
	assertChoice(t, rack, "status", "active", "Active")
	assertChoice(t, rack, "width", float64(19), "19 inches")
	if rack["airflow"] != nil {
		t.Fatalf("airflow = %#v, want null choice representation", rack["airflow"])
	}

	deviceType := workflowRequest(t, router, http.MethodPost, "/api/dcim/device-types/", `{"manufacturer":`+formatID(manufacturerID)+`,"model":"Choice Router","slug":"choice-router","airflow":"front-to-rear"}`, http.StatusCreated)
	assertChoice(t, deviceType, "airflow", "front-to-rear", "Front to rear")
	deviceTypeID := int64(deviceType["id"].(float64))
	interfaceTemplate := workflowRequest(t, router, http.MethodPost, "/api/dcim/interface-templates/", `{"device_type":`+formatID(deviceTypeID)+`,"name":"eth0","type":"1000base-t"}`, http.StatusCreated)
	assertChoice(t, interfaceTemplate, "type", "1000base-t", "1000BASE-T (1GE)")

	prefix := workflowRequest(t, router, http.MethodPost, "/api/ipam/prefixes/", `{"prefix":"192.0.2.0/24","status":"reserved"}`, http.StatusCreated)
	assertChoice(t, prefix, "family", float64(4), "IPv4")
	assertChoice(t, prefix, "status", "reserved", "Reserved")

	ipAddress := workflowRequest(t, router, http.MethodPost, "/api/ipam/ip-addresses/", `{"address":"192.0.2.1/24","status":"deprecated","role":"loopback"}`, http.StatusCreated)
	assertChoice(t, ipAddress, "family", float64(4), "IPv4")
	assertChoice(t, ipAddress, "status", "deprecated", "Deprecated")
	assertChoice(t, ipAddress, "role", "loopback", "Loopback")
}

func TestTypedSiteRESTPreservesPresenceAndRejectsUnknownFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:typed_site_rest_presence?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(profileRowModels()...); err != nil {
		t.Fatal(err)
	}
	core := composition.NewCore(db)
	router := gin.New()
	newTypedCoreHandler(core).Register(router, func(c *gin.Context) {
		SetPrincipal(c, identity.Principal{ID: 1, Username: "presence", IsSuperuser: true})
		c.Next()
	})

	missing := workflowRequest(
		t,
		router,
		http.MethodPost,
		"/api/dcim/sites/",
		`{"name":"Missing Slug"}`,
		http.StatusBadRequest,
	)
	if got := missing["slug"]; !reflect.DeepEqual(got, []any{"This field is required."}) {
		t.Fatalf("missing slug response = %#v", missing)
	}

	unknown := workflowRequest(
		t,
		router,
		http.MethodPost,
		"/api/dcim/sites/",
		`{"name":"Unknown","slug":"unknown","tenant":null}`,
		http.StatusBadRequest,
	)
	if got := unknown["tenant"]; !reflect.DeepEqual(got, []any{"This field is not supported by the active capability profile."}) {
		t.Fatalf("unknown field response = %#v", unknown)
	}

	created := workflowRequest(
		t,
		router,
		http.MethodPost,
		"/api/dcim/sites/",
		`{"name":"Presence","slug":"presence"}`,
		http.StatusCreated,
	)
	id := int64(created["id"].(float64))
	nullPatch := workflowRequest(
		t,
		router,
		http.MethodPatch,
		"/api/dcim/sites/"+formatID(id)+"/",
		`{"facility":null}`,
		http.StatusBadRequest,
	)
	if got := nullPatch["facility"]; !reflect.DeepEqual(got, []any{"This field may not be null."}) {
		t.Fatalf("explicit null response = %#v", nullPatch)
	}
}

func TestTypedSiteRESTUsesPinnedPaginationAndRepeatedSignedFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, router := newTypedSiteRESTFixture(t, "typed_site_rest_list_contract")

	now := shared.SystemClock{}.Now().Time
	rows := make([]dcimrow.SiteRow, 51)
	for index := range rows {
		name := "Site-" + strconv.Itoa(index+1)
		slug := "site-" + strconv.Itoa(index+1)
		status := "active"
		if index == 0 {
			name, slug = "Alpha", "alpha"
		}
		if index == 1 {
			name, slug, status = "Beta", "beta", "planned"
		}
		rows[index] = dcimrow.SiteRow{
			RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now},
			Name:        name,
			Slug:        slug,
			Status:      status,
		}
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatal(err)
	}

	omitted := workflowRequest(t, router, http.MethodGet, "/api/dcim/sites/?ordering=id", "", http.StatusOK)
	if got := omitted["count"]; got != float64(51) {
		t.Fatalf("omitted-limit count = %#v, want 51", got)
	}
	if got := len(listResults(t, omitted)); got != int(applicationdcim.DefaultSitePageLimit) {
		t.Fatalf("omitted-limit result count = %d, want %d", got, applicationdcim.DefaultSitePageLimit)
	}

	unpaginated := workflowRequest(t, router, http.MethodGet, "/api/dcim/sites/?ordering=id&limit=0", "", http.StatusOK)
	if got := len(listResults(t, unpaginated)); got != 51 {
		t.Fatalf("limit=0 result count = %d, want 51", got)
	}
	if unpaginated["next"] != nil || unpaginated["previous"] != nil {
		t.Fatalf("limit=0 links = next %#v previous %#v, want nil", unpaginated["next"], unpaginated["previous"])
	}

	filteredPath := "/api/dcim/sites/?ordering=id" +
		"&id=-1&id=%2B" + strconv.FormatInt(rows[0].ID, 10) +
		"&id=" + strconv.FormatInt(rows[1].ID, 10) +
		"&name=Alpha&name=Beta" +
		"&slug=alpha&slug=beta" +
		"&status=active&status=planned"
	filtered := workflowRequest(t, router, http.MethodGet, filteredPath, "", http.StatusOK)
	if got := filtered["count"]; got != float64(2) {
		t.Fatalf("repeated-filter count = %#v, want 2", got)
	}
	results := listResults(t, filtered)
	if got := []any{results[0].(map[string]any)["name"], results[1].(map[string]any)["name"]}; !reflect.DeepEqual(got, []any{"Alpha", "Beta"}) {
		t.Fatalf("repeated-filter names = %#v, want Alpha and Beta", got)
	}

	negative := workflowRequest(t, router, http.MethodGet, "/api/dcim/sites/?id=-1", "", http.StatusOK)
	if got := negative["count"]; got != float64(0) {
		t.Fatalf("negative-ID count = %#v, want 0", got)
	}
}

func TestTypedSiteRESTMapsUniqueConstraintsToExactNetBoxFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, router := newTypedSiteRESTFixture(t, "typed_site_rest_unique_contract")

	workflowRequest(
		t,
		router,
		http.MethodPost,
		"/api/dcim/sites/",
		`{"name":"Moscow","slug":"moscow"}`,
		http.StatusCreated,
	)
	duplicateName := workflowRequest(
		t,
		router,
		http.MethodPost,
		"/api/dcim/sites/",
		`{"name":"Moscow","slug":"moscow-two"}`,
		http.StatusBadRequest,
	)
	if want := map[string]any{"name": []any{"site with this name already exists."}}; !reflect.DeepEqual(duplicateName, want) {
		t.Fatalf("duplicate-name response = %#v, want %#v", duplicateName, want)
	}

	duplicateSlug := workflowRequest(
		t,
		router,
		http.MethodPost,
		"/api/dcim/sites/",
		`{"name":"Moscow Two","slug":"moscow"}`,
		http.StatusBadRequest,
	)
	if want := map[string]any{"slug": []any{"site with this slug already exists."}}; !reflect.DeepEqual(duplicateSlug, want) {
		t.Fatalf("duplicate-slug response = %#v, want %#v", duplicateSlug, want)
	}
}

func newTypedSiteRESTFixture(t *testing.T, databaseName string) (*gorm.DB, http.Handler) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+databaseName+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(profileRowModels()...); err != nil {
		t.Fatal(err)
	}
	core := composition.NewCore(db)
	router := gin.New()
	newTypedCoreHandler(core).Register(router, func(c *gin.Context) {
		SetPrincipal(c, identity.Principal{ID: 1, Username: "site-contract", IsSuperuser: true})
		c.Next()
	})
	return db, router
}

func listResults(t *testing.T, page map[string]any) []any {
	t.Helper()
	results, ok := page["results"].([]any)
	if !ok {
		t.Fatalf("results = %#v, want array", page["results"])
	}
	return results
}

func assertChoice(t *testing.T, response map[string]any, field string, wantValue any, wantLabel string) {
	t.Helper()
	choice, ok := response[field].(map[string]any)
	if !ok {
		t.Fatalf("%s = %#v, want choice object", field, response[field])
	}
	if got := choice["value"]; got != wantValue {
		t.Errorf("%s.value = %#v, want %#v", field, got, wantValue)
	}
	if got := choice["label"]; got != wantLabel {
		t.Errorf("%s.label = %#v, want %#v", field, got, wantLabel)
	}
	if len(choice) != 2 {
		t.Errorf("%s = %#v, want exactly value and label", field, choice)
	}
}

func workflowRequest(t *testing.T, router http.Handler, method, path, body string, wantStatus int) map[string]any {
	t.Helper()
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != wantStatus {
		t.Fatalf("%s %s = %d, want %d: %s", method, path, response.Code, wantStatus, response.Body.String())
	}
	var decoded map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}
	return decoded
}

func formatID(id int64) string {
	return strconv.FormatInt(id, 10)
}

func newTypedCoreHandler(core composition.Core) *Handler {
	return newCompleteTypedHandler(
		core.Sites,
		WithOrganizationServices(core.Manufacturers, core.RackRoles),
		WithRackTypeService(core.RackTypes),
		WithRackService(core.Racks),
		WithDeviceRoleService(core.DeviceRoles),
		WithDeviceTypeService(core.DeviceTypes),
		WithInterfaceTemplateService(core.InterfaceTemplates),
		WithDeviceService(core.Devices),
		WithInterfaceService(core.Interfaces),
		WithVRFService(core.VRFs),
		WithPrefixService(core.Prefixes),
		WithIPAddressService(core.IPAddresses),
	)
}

func profileRowModels() []any {
	models := make([]any, 0, 14)
	models = append(models, dcimrow.Models()...)
	models = append(models, ipamrow.Models()...)
	return append(models, &changelogpostgres.ChangeRow{})
}
