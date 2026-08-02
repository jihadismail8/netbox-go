// Package workflow is the NetBox-compatible REST adapter for the first
// capability profile. It translates HTTP only; all policy lives in the shared
// application service.
package workflow

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

const principalKey = "netbox.principal"

type Handler struct {
	sites              SiteService
	organizations      *OrganizationHandler
	rackTypes          *RackTypeRESTHandler
	racks              *RackRESTHandler
	deviceRoles        *DeviceRoleHandler
	deviceTypes        *DeviceTypeRESTHandler
	interfaceTemplates *InterfaceTemplateRESTHandler
	devices            *DeviceRESTHandler
	interfaces         *InterfaceRESTHandler
	vrfs               *VRFRESTHandler
	prefixes           *PrefixRESTHandler
	ipAddresses        *IPAddressRESTHandler
}

// HandlerOption supplies one typed resource handler to the complete profile
// runtime. NewHandler rejects an incomplete option set.
type HandlerOption func(*Handler)

func WithOrganizationServices(
	manufacturers ManufacturerService,
	rackRoles RackRoleService,
) HandlerOption {
	return func(handler *Handler) {
		handler.organizations = NewOrganizationHandler(manufacturers, rackRoles)
	}
}

func WithVRFService(vrfs VRFService) HandlerOption {
	return func(handler *Handler) {
		handler.vrfs = NewVRFRESTHandler(vrfs)
	}
}

func WithPrefixService(prefixes PrefixService) HandlerOption {
	return func(handler *Handler) {
		handler.prefixes = NewPrefixRESTHandler(prefixes)
	}
}

func WithIPAddressService(ipAddresses IPAddressService) HandlerOption {
	return func(handler *Handler) {
		handler.ipAddresses = NewIPAddressRESTHandler(ipAddresses)
	}
}

func WithRackTypeService(rackTypes RackTypeService) HandlerOption {
	return func(handler *Handler) {
		handler.rackTypes = NewRackTypeRESTHandler(rackTypes)
	}
}

func WithRackService(racks RackService) HandlerOption {
	return func(handler *Handler) {
		handler.racks = NewRackRESTHandler(racks)
	}
}

func WithDeviceRoleService(deviceRoles DeviceRoleService) HandlerOption {
	return func(handler *Handler) {
		handler.deviceRoles = NewDeviceRoleHandler(deviceRoles)
	}
}

func WithDeviceTypeService(deviceTypes DeviceTypeService) HandlerOption {
	return func(handler *Handler) {
		handler.deviceTypes = NewDeviceTypeRESTHandler(deviceTypes)
	}
}

func WithInterfaceTemplateService(
	interfaceTemplates InterfaceTemplateService,
) HandlerOption {
	return func(handler *Handler) {
		handler.interfaceTemplates = NewInterfaceTemplateRESTHandler(interfaceTemplates)
	}
}

func WithDeviceService(devices DeviceService) HandlerOption {
	return func(handler *Handler) {
		handler.devices = NewDeviceRESTHandler(devices)
	}
}

func WithInterfaceService(interfaces InterfaceService) HandlerOption {
	return func(handler *Handler) {
		handler.interfaces = NewInterfaceRESTHandler(interfaces)
	}
}

func NewHandler(sites SiteService, options ...HandlerOption) *Handler {
	handler := &Handler{sites: sites}
	for _, option := range options {
		if option != nil {
			option(handler)
		}
	}
	if missing := handler.missingTypedServices(); len(missing) > 0 {
		panic("REST profile handler requires typed services for: " + strings.Join(missing, ", "))
	}
	return handler
}

func (h *Handler) missingTypedServices() []string {
	missing := make([]string, 0, 13)
	if h.sites == nil {
		missing = append(missing, "Site")
	}
	if h.organizations == nil {
		missing = append(missing, "Manufacturer", "RackRole")
	}
	if h.rackTypes == nil {
		missing = append(missing, "RackType")
	}
	if h.racks == nil {
		missing = append(missing, "Rack")
	}
	if h.deviceRoles == nil {
		missing = append(missing, "DeviceRole")
	}
	if h.deviceTypes == nil {
		missing = append(missing, "DeviceType")
	}
	if h.interfaceTemplates == nil {
		missing = append(missing, "InterfaceTemplate")
	}
	if h.devices == nil {
		missing = append(missing, "Device")
	}
	if h.interfaces == nil {
		missing = append(missing, "Interface")
	}
	if h.vrfs == nil {
		missing = append(missing, "VRF")
	}
	if h.prefixes == nil {
		missing = append(missing, "Prefix")
	}
	if h.ipAddresses == nil {
		missing = append(missing, "IPAddress")
	}
	return missing
}

// SetPrincipal stores an authenticated application principal for downstream
// handlers. Authentication middleware is the only runtime caller.
func SetPrincipal(c *gin.Context, principal identity.Principal) { c.Set(principalKey, principal) }
func Principal(c *gin.Context) (identity.Principal, bool) {
	value, ok := c.Get(principalKey)
	if !ok {
		return identity.Principal{}, false
	}
	principal, ok := value.(identity.Principal)
	return principal, ok && principal.Authenticated()
}

func (h *Handler) Register(r gin.IRoutes, middlewares ...gin.HandlerFunc) {
	h.registerSites(r, middlewares...)
	h.organizations.Register(r, middlewares...)
	h.rackTypes.Register(r, middlewares...)
	h.racks.Register(r, middlewares...)
	h.deviceRoles.Register(r, middlewares...)
	h.deviceTypes.Register(r, middlewares...)
	h.interfaceTemplates.Register(r, middlewares...)
	h.devices.Register(r, middlewares...)
	h.interfaces.Register(r, middlewares...)
	h.vrfs.Register(r, middlewares...)
	h.prefixes.Register(r, middlewares...)
	h.ipAddresses.Register(r, middlewares...)
}

func parseID(raw string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, shared.Invalid("id", "A positive integer is required.")
	}
	return id, nil
}

func writeError(c *gin.Context, err error) {
	var sharedErr *shared.Error
	if !errors.As(err, &sharedErr) {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "An internal error occurred."})
		return
	}
	status := map[shared.ErrorReason]int{
		shared.ErrorReasonValidation:      http.StatusBadRequest,
		shared.ErrorReasonUnauthenticated: http.StatusForbidden,
		shared.ErrorReasonForbidden:       http.StatusForbidden,
		shared.ErrorReasonNotFound:        http.StatusNotFound,
		shared.ErrorReasonConflict:        http.StatusBadRequest,
		shared.ErrorReasonProtected:       http.StatusConflict,
		shared.ErrorReasonRateLimited:     http.StatusTooManyRequests,
		shared.ErrorReasonInternal:        http.StatusInternalServerError,
	}[sharedErr.Reason]
	if status == 0 {
		status = http.StatusInternalServerError
	}
	if len(sharedErr.FieldViolations) > 0 {
		fields := make(map[string][]string)
		for _, violation := range sharedErr.FieldViolations {
			fields[violation.Field] = append(fields[violation.Field], violation.Description)
		}
		c.JSON(status, fields)
		return
	}
	c.JSON(status, gin.H{"detail": sharedErr.Message})
}
