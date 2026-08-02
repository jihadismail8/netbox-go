package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	applicationipam "netbox-go/internal/application/ipam"
	"netbox-go/internal/domain/identity"
	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

type IPAddressService interface {
	ListIPAddresses(
		context.Context,
		identity.Principal,
		applicationipam.ListIPAddressesQuery,
	) (applicationipam.IPAddressPage, error)
	GetIPAddress(
		context.Context,
		identity.Principal,
		applicationipam.GetIPAddressQuery,
	) (*domainipam.IPAddress, error)
	CreateIPAddress(
		context.Context,
		identity.Principal,
		applicationipam.CreateIPAddressCommand,
	) (*domainipam.IPAddress, error)
	ReplaceIPAddress(
		context.Context,
		identity.Principal,
		applicationipam.ReplaceIPAddressCommand,
	) (*domainipam.IPAddress, error)
	UpdateIPAddress(
		context.Context,
		identity.Principal,
		applicationipam.UpdateIPAddressCommand,
	) (*domainipam.IPAddress, error)
	DeleteIPAddress(
		context.Context,
		identity.Principal,
		applicationipam.DeleteIPAddressCommand,
	) error
}

type IPAddressRESTHandler struct{ service IPAddressService }

func NewIPAddressRESTHandler(service IPAddressService) *IPAddressRESTHandler {
	if service == nil {
		panic("REST IPAddress handler requires a typed IPAddress service")
	}
	return &IPAddressRESTHandler{service: service}
}

func (handler *IPAddressRESTHandler) Register(
	r gin.IRoutes,
	middlewares ...gin.HandlerFunc,
) {
	const base = "/api/ipam/ip-addresses/"
	r.GET(base, append(middlewares, handler.list())...)
	r.POST(base, append(middlewares, handler.create())...)
	r.GET(base+":id/", append(middlewares, handler.get())...)
	r.PUT(base+":id/", append(middlewares, handler.replace())...)
	r.PATCH(base+":id/", append(middlewares, handler.update())...)
	r.DELETE(base+":id/", append(middlewares, handler.delete())...)
}

func (handler *IPAddressRESTHandler) list() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := Principal(c)
		if !ok {
			writeError(c, shared.Unauthenticated())
			return
		}
		query, err := parseIPAddressList(c.Request.URL.Query())
		if err != nil {
			writeError(c, err)
			return
		}
		page, err := handler.service.ListIPAddresses(
			c.Request.Context(), principal, query,
		)
		if err != nil {
			writeError(c, err)
			return
		}
		results := make([]ipAddressDTO, 0, len(page.Results))
		for _, address := range page.Results {
			results = append(results, newIPAddressDTO(address))
		}
		c.JSON(http.StatusOK, ipAddressListDTO{
			Count:    page.Count,
			Next:     ipAddressPageURL(c, query, page.Count, true),
			Previous: ipAddressPageURL(c, query, page.Count, false),
			Results:  results,
		})
	}
}

func (handler *IPAddressRESTHandler) get() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := ipAddressRequestIdentity(c)
		if !ok {
			return
		}
		address, err := handler.service.GetIPAddress(
			c.Request.Context(),
			principal,
			applicationipam.GetIPAddressQuery{ID: id},
		)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newIPAddressDTO(address))
	}
}

func (handler *IPAddressRESTHandler) create() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := Principal(c)
		if !ok {
			writeError(c, shared.Unauthenticated())
			return
		}
		input, err := decodeIPAddressInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		address, err := handler.service.CreateIPAddress(
			c.Request.Context(), principal, input.createCommand(),
		)
		if err != nil {
			writeError(c, err)
			return
		}
		c.Header("Location", ipAddressURL(address.ID()))
		c.JSON(http.StatusCreated, newIPAddressDTO(address))
	}
}

func (handler *IPAddressRESTHandler) replace() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := ipAddressRequestIdentity(c)
		if !ok {
			return
		}
		input, err := decodeIPAddressInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		address, err := handler.service.ReplaceIPAddress(
			c.Request.Context(), principal, input.replaceCommand(id),
		)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newIPAddressDTO(address))
	}
}

func (handler *IPAddressRESTHandler) update() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := ipAddressRequestIdentity(c)
		if !ok {
			return
		}
		input, err := decodeIPAddressInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		address, err := handler.service.UpdateIPAddress(
			c.Request.Context(), principal, input.updateCommand(id),
		)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newIPAddressDTO(address))
	}
}

func (handler *IPAddressRESTHandler) delete() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := ipAddressRequestIdentity(c)
		if !ok {
			return
		}
		if err := handler.service.DeleteIPAddress(
			c.Request.Context(),
			principal,
			applicationipam.DeleteIPAddressCommand{ID: id},
		); err != nil {
			writeError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

type ipAddressListDTO struct {
	Count    uint64         `json:"count"`
	Next     *string        `json:"next"`
	Previous *string        `json:"previous"`
	Results  []ipAddressDTO `json:"results"`
}

type ipAddressInterfaceDTO struct {
	ID      int64  `json:"id"`
	URL     string `json:"url"`
	Display string `json:"display"`
}

type ipAddressDTO struct {
	ID                 int64                  `json:"id"`
	URL                string                 `json:"url"`
	Display            string                 `json:"display"`
	Family             prefixChoiceUint32DTO  `json:"family"`
	Address            string                 `json:"address"`
	VRF                *prefixVRFDTO          `json:"vrf"`
	Status             prefixChoiceStringDTO  `json:"status"`
	Role               *prefixChoiceStringDTO `json:"role"`
	DNSName            string                 `json:"dns_name"`
	Description        string                 `json:"description"`
	Comments           string                 `json:"comments"`
	AssignedObjectType *string                `json:"assigned_object_type"`
	AssignedObjectID   *int64                 `json:"assigned_object_id"`
	AssignedObject     *ipAddressInterfaceDTO `json:"assigned_object"`
	Created            time.Time              `json:"created"`
	LastUpdated        time.Time              `json:"last_updated"`
}

func newIPAddressDTO(address *domainipam.IPAddress) ipAddressDTO {
	if address == nil {
		return ipAddressDTO{}
	}
	dto := ipAddressDTO{
		ID: address.ID().Int64(), URL: ipAddressURL(address.ID()),
		Display: address.Display(), Family: prefixFamilyChoice(address.Family()),
		Address: address.Address().String(),
		VRF:     prefixVRFProjection(address.VRF()),
		Status:  ipAddressStatusChoice(address.Status()),
		Role:    ipAddressRoleChoice(address.Role()),
		DNSName: address.DNSName(), Description: address.Description(),
		Comments: address.Comments(), Created: address.Created().Time,
		LastUpdated: address.LastUpdated().Time,
	}
	if assignment, present := address.Assignment().Get(); present {
		objectType := domainipam.IPAddressAssignmentType
		id := assignment.ID().Int64()
		dto.AssignedObjectType, dto.AssignedObjectID = &objectType, &id
		dto.AssignedObject = &ipAddressInterfaceDTO{
			ID: id, URL: "/api/dcim/interfaces/" + assignment.ID().String() + "/",
			Display: assignment.Display(),
		}
	}
	return dto
}

func ipAddressStatusChoice(
	status domainipam.IPAddressStatus,
) prefixChoiceStringDTO {
	labels := map[domainipam.IPAddressStatus]string{
		domainipam.IPAddressStatusActive:     "Active",
		domainipam.IPAddressStatusReserved:   "Reserved",
		domainipam.IPAddressStatusDeprecated: "Deprecated",
		domainipam.IPAddressStatusDHCP:       "DHCP",
		domainipam.IPAddressStatusSLAAC:      "SLAAC",
	}
	return prefixChoiceStringDTO{Value: status.String(), Label: labels[status]}
}

func ipAddressRoleChoice(
	nullable domainipam.NullableIPAddressRole,
) *prefixChoiceStringDTO {
	role, present := nullable.Get()
	if !present || role.String() == "" {
		return nil
	}
	labels := map[domainipam.IPAddressRole]string{
		domainipam.IPAddressRoleLoopback:  "Loopback",
		domainipam.IPAddressRoleSecondary: "Secondary",
		domainipam.IPAddressRoleAnycast:   "Anycast",
		domainipam.IPAddressRoleVIP:       "VIP",
		domainipam.IPAddressRoleVRRP:      "VRRP",
		domainipam.IPAddressRoleHSRP:      "HSRP",
		domainipam.IPAddressRoleGLBP:      "GLBP",
		domainipam.IPAddressRoleCARP:      "CARP",
	}
	return &prefixChoiceStringDTO{
		Value: role.String(), Label: labels[role],
	}
}

func ipAddressURL(id shared.ID) string {
	return "/api/ipam/ip-addresses/" + id.String() + "/"
}

func ipAddressRequestIdentity(
	c *gin.Context,
) (identity.Principal, shared.ID, bool) {
	principal, ok := Principal(c)
	if !ok {
		writeError(c, shared.Unauthenticated())
		return identity.Principal{}, 0, false
	}
	primitive, err := parseID(c.Param("id"))
	if err != nil {
		writeError(c, err)
		return identity.Principal{}, 0, false
	}
	return principal, shared.ID(primitive), true
}

type ipAddressInputJSON struct {
	Address            json.RawMessage `json:"address"`
	VRF                json.RawMessage `json:"vrf"`
	Status             json.RawMessage `json:"status"`
	Role               json.RawMessage `json:"role"`
	DNSName            json.RawMessage `json:"dns_name"`
	Description        json.RawMessage `json:"description"`
	Comments           json.RawMessage `json:"comments"`
	AssignedObjectType json.RawMessage `json:"assigned_object_type"`
	AssignedObjectID   json.RawMessage `json:"assigned_object_id"`
}

type decodedIPAddressInput struct {
	address            applicationipam.Field[string]
	vrf                applicationipam.Field[int64]
	status             applicationipam.Field[string]
	role               applicationipam.Field[string]
	dnsName            applicationipam.Field[string]
	description        applicationipam.Field[string]
	comments           applicationipam.Field[string]
	assignedObjectType applicationipam.Field[string]
	assignedObjectID   applicationipam.Field[int64]
}

func decodeIPAddressInput(c *gin.Context) (decodedIPAddressInput, error) {
	raw, err := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20))
	if err != nil {
		return decodedIPAddressInput{},
			shared.Invalid("non_field_errors", "Expected a JSON object.")
	}
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return decodedIPAddressInput{},
			shared.Invalid("non_field_errors", "Expected a JSON object.")
	}
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.DisallowUnknownFields()
	var input ipAddressInputJSON
	if err := decoder.Decode(&input); err != nil {
		if field, unknown := unknownJSONField(err); unknown {
			return decodedIPAddressInput{}, shared.Invalid(
				field,
				"This field is not supported by the active capability profile.")

		}
		return decodedIPAddressInput{},
			shared.Invalid("non_field_errors", "Expected a JSON object.")
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return decodedIPAddressInput{}, shared.Invalid(
			"non_field_errors", "Request body must contain one JSON object.")

	}
	address, err := decodeVRFStringField("address", input.Address)
	if err != nil {
		return decodedIPAddressInput{}, err
	}
	vrf, err := decodePrefixInt64Field("vrf", input.VRF)
	if err != nil {
		return decodedIPAddressInput{}, err
	}
	status, err := decodeVRFStringField("status", input.Status)
	if err != nil {
		return decodedIPAddressInput{}, err
	}
	role, err := decodeVRFStringField("role", input.Role)
	if err != nil {
		return decodedIPAddressInput{}, err
	}
	dnsName, err := decodeVRFStringField("dns_name", input.DNSName)
	if err != nil {
		return decodedIPAddressInput{}, err
	}
	description, err := decodeVRFStringField("description", input.Description)
	if err != nil {
		return decodedIPAddressInput{}, err
	}
	comments, err := decodeVRFStringField("comments", input.Comments)
	if err != nil {
		return decodedIPAddressInput{}, err
	}
	assignedObjectType, err := decodeVRFStringField(
		"assigned_object_type", input.AssignedObjectType,
	)
	if err != nil {
		return decodedIPAddressInput{}, err
	}
	assignedObjectID, err := decodePrefixInt64Field(
		"assigned_object_id", input.AssignedObjectID,
	)
	if err != nil {
		return decodedIPAddressInput{}, err
	}
	return decodedIPAddressInput{
		address: address, vrf: vrf, status: status, role: role,
		dnsName: dnsName, description: description, comments: comments,
		assignedObjectType: assignedObjectType,
		assignedObjectID:   assignedObjectID,
	}, nil
}

func (input decodedIPAddressInput) createCommand() applicationipam.CreateIPAddressCommand {
	return applicationipam.CreateIPAddressCommand{
		Address: input.address, VRF: input.vrf, Status: input.status,
		Role: input.role, DNSName: input.dnsName,
		Description: input.description, Comments: input.comments,
		AssignedObjectType: input.assignedObjectType,
		AssignedObjectID:   input.assignedObjectID,
	}
}

func (input decodedIPAddressInput) replaceCommand(
	id shared.ID,
) applicationipam.ReplaceIPAddressCommand {
	return applicationipam.ReplaceIPAddressCommand{
		ID: id, CreateIPAddressCommand: input.createCommand(),
	}
}

func (input decodedIPAddressInput) updateCommand(
	id shared.ID,
) applicationipam.UpdateIPAddressCommand {
	return applicationipam.UpdateIPAddressCommand{
		ID: id, Address: input.address, VRF: input.vrf,
		Status: input.status, Role: input.role, DNSName: input.dnsName,
		Description: input.description, Comments: input.comments,
		AssignedObjectType: input.assignedObjectType,
		AssignedObjectID:   input.assignedObjectID,
	}
}

func parseIPAddressList(
	values url.Values,
) (applicationipam.ListIPAddressesQuery, error) {
	allowed := map[string]struct{}{
		"limit": {}, "offset": {}, "q": {}, "ordering": {}, "id": {},
		"vrf_id": {}, "vrf_rd": {}, "address": {}, "family": {},
		"parent": {}, "status": {}, "assigned": {}, "interface_id": {},
		"device_id": {},
	}
	for key := range values {
		if _, ok := allowed[key]; !ok {
			return applicationipam.ListIPAddressesQuery{},
				shared.Invalid(key, "Unsupported filter.")
		}
	}
	query := applicationipam.ListIPAddressesQuery{}
	if raw := values.Get("limit"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value < 0 || value > math.MaxUint32 {
			return query, shared.Invalid(
				"limit", "A valid integer is required.")

		}
		query.Limit, query.LimitPresent = uint32(value), true
	}
	if raw := values.Get("offset"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value < 0 || value > math.MaxUint32 {
			return query, shared.Invalid(
				"offset", "A valid integer is required.")

		}
		query.Offset = uint32(value)
	}
	query.Query = values.Get("q")
	var err error
	query.IDs, err = parseRepeatedSignedIntegers(values["id"])
	if err != nil {
		return query, shared.Invalid("id", "A valid integer is required.")
	}
	query.VRFIDs, err = parseRepeatedSignedIntegers(values["vrf_id"])
	if err != nil {
		return query, shared.Invalid("vrf_id", "A valid integer is required.")
	}
	query.InterfaceIDs, err = parseRepeatedSignedIntegers(
		values["interface_id"],
	)
	if err != nil {
		return query,
			shared.Invalid("interface_id", "A valid integer is required.")
	}
	query.DeviceIDs, err = parseRepeatedSignedIntegers(values["device_id"])
	if err != nil {
		return query,
			shared.Invalid("device_id", "A valid integer is required.")
	}
	for _, raw := range values["ordering"] {
		query.Ordering = append(query.Ordering, strings.Split(raw, ",")...)
	}
	query.VRFRDs = append([]string(nil), values["vrf_rd"]...)
	query.Addresses = append([]string(nil), values["address"]...)
	query.Statuses = append([]string(nil), values["status"]...)
	if raw := values.Get("family"); raw != "" {
		value, parseErr := strconv.ParseInt(raw, 10, 64)
		if parseErr != nil {
			return query,
				shared.Invalid("family", "A valid integer is required.")
		}
		query.Family = &value
	}
	query.Parent = optionalQueryString(values, "parent")
	if raw := values.Get("assigned"); raw != "" {
		value, parseErr := strconv.ParseBool(raw)
		if parseErr != nil {
			return query,
				shared.Invalid("assigned", "A valid boolean is required.")
		}
		query.Assigned = &value
	}
	return query, nil
}

func ipAddressPageURL(
	c *gin.Context,
	query applicationipam.ListIPAddressesQuery,
	count uint64,
	next bool,
) *string {
	offset, limit := uint64(query.Offset), uint64(query.EffectiveLimit())
	if next {
		offset += limit
		if offset >= count {
			return nil
		}
	} else {
		if offset == 0 {
			return nil
		}
		if offset <= limit {
			offset = 0
		} else {
			offset -= limit
		}
	}
	values := c.Request.URL.Query()
	values.Set("limit", strconv.FormatUint(limit, 10))
	if offset == 0 {
		values.Del("offset")
	} else {
		values.Set("offset", strconv.FormatUint(offset, 10))
	}
	page := c.Request.URL.Path + "?" + values.Encode()
	return &page
}

var _ IPAddressService = (*applicationipam.IPAddressService)(nil)
