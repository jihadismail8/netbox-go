package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

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
	if !utf8.Valid(raw) {
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
	address, err := decodeIPAddressAddressField(input.Address)
	if err != nil {
		return decodedIPAddressInput{}, err
	}
	vrf, err := decodePrefixInt64Field("vrf", input.VRF)
	if err != nil {
		return decodedIPAddressInput{}, err
	}
	status, err := decodeIPAddressChoiceField("status", input.Status)
	if err != nil {
		return decodedIPAddressInput{}, err
	}
	role, err := decodeIPAddressChoiceField("role", input.Role)
	if err != nil {
		return decodedIPAddressInput{}, err
	}
	dnsName, err := decodeIPAddressCharField("dns_name", input.DNSName, false)
	if err != nil {
		return decodedIPAddressInput{}, err
	}
	description, err := decodeIPAddressCharField("description", input.Description, true)
	if err != nil {
		return decodedIPAddressInput{}, err
	}
	comments, err := decodeIPAddressCharField("comments", input.Comments, true)
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

func decodeIPAddressAddressField(
	raw json.RawMessage,
) (applicationipam.Field[string], error) {
	if len(raw) == 0 {
		return applicationipam.OmittedField[string](), nil
	}
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return applicationipam.NullField[string](), nil
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return applicationipam.FieldValue(text), nil
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return applicationipam.Field[string]{}, shared.Invalid(
			"address", "Must be a string.",
		)
	}
	typeName := ""
	switch typed := value.(type) {
	case bool:
		typeName = "bool"
	case json.Number:
		if strings.ContainsAny(typed.String(), ".eE") {
			typeName = "float"
		} else {
			typeName = "int"
		}
	case map[string]any:
		typeName = "dict"
	case []any:
		typeName = "list"
	}
	if typeName == "" {
		return applicationipam.Field[string]{}, shared.Invalid(
			"address", "Must be a string.",
		)
	}
	return applicationipam.Field[string]{}, shared.Invalid(
		"address",
		fmt.Sprintf("unexpected type <class '%s'> for addr arg", typeName),
	)
}

func decodeIPAddressChoiceField(
	name string,
	raw json.RawMessage,
) (applicationipam.Field[string], error) {
	if len(raw) == 0 {
		return applicationipam.OmittedField[string](), nil
	}
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return applicationipam.NullField[string](), nil
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return applicationipam.FieldValue(text), nil
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return applicationipam.Field[string]{}, shared.Invalid(name, "Must be a string.")
	}
	switch typed := value.(type) {
	case map[string]any, []any:
		return applicationipam.Field[string]{}, shared.Invalid(
			name,
			`Value must be passed directly (e.g. "foo": 123); do not use a dictionary or list.`,
		)
	case bool:
		return applicationipam.FieldValue(strconv.FormatBool(typed)), nil
	case json.Number:
		return applicationipam.FieldValue(pythonJSONNumberString(typed)), nil
	default:
		return applicationipam.Field[string]{}, shared.Invalid(name, "Must be a string.")
	}
}

func pythonJSONNumberString(value json.Number) string {
	raw := value.String()
	if !strings.ContainsAny(raw, ".eE") {
		if integer, ok := new(big.Int).SetString(raw, 10); ok {
			return integer.String()
		}
		return raw
	}
	number, err := strconv.ParseFloat(raw, 64)
	if err != nil && !errors.Is(err, strconv.ErrRange) {
		return raw
	}
	if math.IsInf(number, 1) {
		return "inf"
	}
	if math.IsInf(number, -1) {
		return "-inf"
	}
	scientific := strconv.FormatFloat(number, 'e', -1, 64)
	exponentStart := strings.LastIndexByte(scientific, 'e') + 1
	exponent, exponentErr := strconv.Atoi(scientific[exponentStart:])
	if exponentErr != nil || exponent < -4 || exponent >= 16 {
		return scientific
	}
	rendered := strconv.FormatFloat(number, 'f', -1, 64)
	if !strings.ContainsRune(rendered, '.') {
		return rendered + ".0"
	}
	return rendered
}

func decodeIPAddressCharField(
	name string,
	raw json.RawMessage,
	rejectSurrogate bool,
) (applicationipam.Field[string], error) {
	if len(raw) == 0 {
		return applicationipam.OmittedField[string](), nil
	}
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return applicationipam.NullField[string](), nil
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		if codePoint, invalid := firstUnpairedJSONSurrogate(raw); invalid {
			if name == "dns_name" || rejectSurrogate {
				return applicationipam.Field[string]{}, shared.NewValidationError(
					ipAddressSurrogateViolations(name, text, codePoint)...,
				)
			}
		}
		return applicationipam.FieldValue(text), nil
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err == nil {
		if number, ok := value.(json.Number); ok {
			return applicationipam.FieldValue(pythonJSONNumberString(number)), nil
		}
	}
	return applicationipam.Field[string]{}, shared.Invalid(name, "Not a valid string.")
}

func ipAddressSurrogateViolations(
	name string,
	value string,
	codePoint uint16,
) []shared.FieldViolation {
	trimmed := strings.TrimSpace(value)
	violations := make([]shared.FieldViolation, 0, 4)
	if name == "dns_name" {
		violations = append(violations, shared.FieldViolation{
			Field: name, Reason: "invalid",
			Description: "Only alphanumeric characters, asterisks, hyphens, periods, and underscores are allowed in DNS names",
		})
	}
	maximum := 0
	switch name {
	case "dns_name":
		maximum = domainipam.IPAddressDNSNameMaxLength
	case "description":
		maximum = domainipam.IPAddressDescriptionMaxLength
	}
	if maximum > 0 && utf8.RuneCountInString(trimmed) > maximum {
		violations = append(violations, shared.FieldViolation{
			Field: name, Reason: "max_length",
			Description: fmt.Sprintf(
				"Ensure this field has no more than %d characters.", maximum,
			),
		})
	}
	if strings.ContainsRune(trimmed, '\x00') {
		violations = append(violations, shared.FieldViolation{
			Field: name, Reason: "invalid",
			Description: "Null characters are not allowed.",
		})
	}
	return append(violations, shared.FieldViolation{
		Field: name, Reason: "invalid",
		Description: fmt.Sprintf(
			"Surrogate characters are not allowed: U+%04X.", codePoint,
		),
	})
}

func firstUnpairedJSONSurrogate(raw json.RawMessage) (uint16, bool) {
	for index := 1; index+5 < len(raw); index++ {
		if raw[index] != '\\' || raw[index+1] != 'u' {
			continue
		}
		backslashes := 1
		for previous := index - 1; previous > 0 && raw[previous] == '\\'; previous-- {
			backslashes++
		}
		if backslashes%2 == 0 {
			continue
		}
		parsed, err := strconv.ParseUint(string(raw[index+2:index+6]), 16, 16)
		if err != nil {
			return 0, false
		}
		codePoint := uint16(parsed)
		if codePoint >= 0xD800 && codePoint <= 0xDBFF {
			if index+11 < len(raw) && raw[index+6] == '\\' && raw[index+7] == 'u' {
				low, lowErr := strconv.ParseUint(string(raw[index+8:index+12]), 16, 16)
				if lowErr == nil && low >= 0xDC00 && low <= 0xDFFF {
					index += 11
					continue
				}
			}
			return codePoint, true
		}
		if codePoint >= 0xDC00 && codePoint <= 0xDFFF {
			return codePoint, true
		}
		index += 5
	}
	return 0, false
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
