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

type PrefixService interface {
	ListPrefixes(context.Context, identity.Principal, applicationipam.ListPrefixesQuery) (applicationipam.PrefixPage, error)
	GetPrefix(context.Context, identity.Principal, applicationipam.GetPrefixQuery) (*domainipam.Prefix, error)
	CreatePrefix(context.Context, identity.Principal, applicationipam.CreatePrefixCommand) (*domainipam.Prefix, error)
	ReplacePrefix(context.Context, identity.Principal, applicationipam.ReplacePrefixCommand) (*domainipam.Prefix, error)
	UpdatePrefix(context.Context, identity.Principal, applicationipam.UpdatePrefixCommand) (*domainipam.Prefix, error)
	DeletePrefix(context.Context, identity.Principal, applicationipam.DeletePrefixCommand) error
}

type PrefixRESTHandler struct{ service PrefixService }

func NewPrefixRESTHandler(service PrefixService) *PrefixRESTHandler {
	if service == nil {
		panic("REST Prefix handler requires a typed Prefix service")
	}
	return &PrefixRESTHandler{service: service}
}

func (handler *PrefixRESTHandler) Register(r gin.IRoutes, middlewares ...gin.HandlerFunc) {
	const base = "/api/ipam/prefixes/"
	r.GET(base, append(middlewares, handler.list())...)
	r.POST(base, append(middlewares, handler.create())...)
	r.GET(base+":id/", append(middlewares, handler.get())...)
	r.PUT(base+":id/", append(middlewares, handler.replace())...)
	r.PATCH(base+":id/", append(middlewares, handler.update())...)
	r.DELETE(base+":id/", append(middlewares, handler.delete())...)
}

func (handler *PrefixRESTHandler) list() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := Principal(c)
		if !ok {
			writeError(c, shared.Unauthenticated())
			return
		}
		query, err := parsePrefixList(c.Request.URL.Query())
		if err != nil {
			writeError(c, err)
			return
		}
		page, err := handler.service.ListPrefixes(c.Request.Context(), principal, query)
		if err != nil {
			writeError(c, err)
			return
		}
		results := make([]prefixDTO, 0, len(page.Results))
		for _, prefix := range page.Results {
			results = append(results, newPrefixDTO(prefix))
		}
		c.JSON(http.StatusOK, prefixListDTO{
			Count: page.Count, Next: prefixPageURL(c, query, page.Count, true),
			Previous: prefixPageURL(c, query, page.Count, false), Results: results,
		})
	}
}

func (handler *PrefixRESTHandler) get() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := prefixRequestIdentity(c)
		if !ok {
			return
		}
		prefix, err := handler.service.GetPrefix(
			c.Request.Context(), principal, applicationipam.GetPrefixQuery{ID: id},
		)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newPrefixDTO(prefix))
	}
}

func (handler *PrefixRESTHandler) create() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := Principal(c)
		if !ok {
			writeError(c, shared.Unauthenticated())
			return
		}
		input, err := decodePrefixInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		prefix, err := handler.service.CreatePrefix(c.Request.Context(), principal, input.createCommand())
		if err != nil {
			writeError(c, err)
			return
		}
		c.Header("Location", prefixURL(prefix.ID()))
		c.JSON(http.StatusCreated, newPrefixDTO(prefix))
	}
}

func (handler *PrefixRESTHandler) replace() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := prefixRequestIdentity(c)
		if !ok {
			return
		}
		input, err := decodePrefixInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		prefix, err := handler.service.ReplacePrefix(c.Request.Context(), principal, input.replaceCommand(id))
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newPrefixDTO(prefix))
	}
}

func (handler *PrefixRESTHandler) update() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := prefixRequestIdentity(c)
		if !ok {
			return
		}
		input, err := decodePrefixInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		prefix, err := handler.service.UpdatePrefix(c.Request.Context(), principal, input.updateCommand(id))
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newPrefixDTO(prefix))
	}
}

func (handler *PrefixRESTHandler) delete() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := prefixRequestIdentity(c)
		if !ok {
			return
		}
		if err := handler.service.DeletePrefix(
			c.Request.Context(), principal, applicationipam.DeletePrefixCommand{ID: id},
		); err != nil {
			writeError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

type prefixListDTO struct {
	Count    uint64      `json:"count"`
	Next     *string     `json:"next"`
	Previous *string     `json:"previous"`
	Results  []prefixDTO `json:"results"`
}

type prefixChoiceStringDTO struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type prefixChoiceUint32DTO struct {
	Value uint32 `json:"value"`
	Label string `json:"label"`
}

type prefixVRFDTO struct {
	ID      int64  `json:"id"`
	URL     string `json:"url"`
	Display string `json:"display"`
}

type prefixDTO struct {
	ID           int64                 `json:"id"`
	URL          string                `json:"url"`
	Display      string                `json:"display"`
	Family       prefixChoiceUint32DTO `json:"family"`
	Prefix       string                `json:"prefix"`
	VRF          *prefixVRFDTO         `json:"vrf"`
	Status       prefixChoiceStringDTO `json:"status"`
	IsPool       bool                  `json:"is_pool"`
	MarkUtilized bool                  `json:"mark_utilized"`
	Description  string                `json:"description"`
	Comments     string                `json:"comments"`
	Created      time.Time             `json:"created"`
	LastUpdated  time.Time             `json:"last_updated"`
	Children     uint64                `json:"children"`
	Depth        uint32                `json:"_depth"`
}

func newPrefixDTO(prefix *domainipam.Prefix) prefixDTO {
	if prefix == nil {
		return prefixDTO{}
	}
	return prefixDTO{
		ID: prefix.ID().Int64(), URL: prefixURL(prefix.ID()), Display: prefix.Display(),
		Family: prefixFamilyChoice(prefix.Family()), Prefix: prefix.Network().String(),
		VRF: prefixVRFProjection(prefix.VRF()), Status: prefixStatusChoice(prefix.Status()),
		IsPool: prefix.IsPool(), MarkUtilized: prefix.MarkUtilized(),
		Description: prefix.Description(), Comments: prefix.Comments(),
		Created: prefix.Created().Time, LastUpdated: prefix.LastUpdated().Time,
		Children: prefix.Children(), Depth: prefix.Depth(),
	}
}

func prefixFamilyChoice(family uint32) prefixChoiceUint32DTO {
	label := "IPv6"
	if family == 4 {
		label = "IPv4"
	}
	return prefixChoiceUint32DTO{Value: family, Label: label}
}

func prefixStatusChoice(status domainipam.PrefixStatus) prefixChoiceStringDTO {
	labels := map[domainipam.PrefixStatus]string{
		domainipam.PrefixStatusContainer: "Container", domainipam.PrefixStatusActive: "Active",
		domainipam.PrefixStatusReserved: "Reserved", domainipam.PrefixStatusDeprecated: "Deprecated",
	}
	return prefixChoiceStringDTO{Value: status.String(), Label: labels[status]}
}

func prefixVRFProjection(nullable domainipam.NullableVRFReference) *prefixVRFDTO {
	reference, present := nullable.Get()
	if !present {
		return nil
	}
	return &prefixVRFDTO{
		ID: reference.ID().Int64(), URL: vrfURL(reference.ID()), Display: reference.Display(),
	}
}

func prefixURL(id shared.ID) string { return "/api/ipam/prefixes/" + id.String() + "/" }

func prefixRequestIdentity(c *gin.Context) (identity.Principal, shared.ID, bool) {
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

type prefixInputJSON struct {
	Prefix       json.RawMessage `json:"prefix"`
	VRF          json.RawMessage `json:"vrf"`
	Status       json.RawMessage `json:"status"`
	IsPool       json.RawMessage `json:"is_pool"`
	MarkUtilized json.RawMessage `json:"mark_utilized"`
	Description  json.RawMessage `json:"description"`
	Comments     json.RawMessage `json:"comments"`
}

type decodedPrefixInput struct {
	prefix       applicationipam.Field[string]
	vrf          applicationipam.Field[int64]
	status       applicationipam.Field[string]
	isPool       applicationipam.Field[bool]
	markUtilized applicationipam.Field[bool]
	description  applicationipam.Field[string]
	comments     applicationipam.Field[string]
}

func decodePrefixInput(c *gin.Context) (decodedPrefixInput, error) {
	raw, err := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20))
	if err != nil {
		return decodedPrefixInput{}, shared.Invalid("non_field_errors", "Expected a JSON object.")
	}
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return decodedPrefixInput{}, shared.Invalid("non_field_errors", "Expected a JSON object.")
	}
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.DisallowUnknownFields()
	var input prefixInputJSON
	if err := decoder.Decode(&input); err != nil {
		if field, unknown := unknownJSONField(err); unknown {
			return decodedPrefixInput{}, shared.Invalid(
				field, "This field is not supported by the active capability profile.")

		}
		return decodedPrefixInput{}, shared.Invalid("non_field_errors", "Expected a JSON object.")
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return decodedPrefixInput{}, shared.Invalid(
			"non_field_errors", "Request body must contain one JSON object.")

	}
	prefix, err := decodeVRFStringField("prefix", input.Prefix)
	if err != nil {
		return decodedPrefixInput{}, err
	}
	vrf, err := decodePrefixInt64Field("vrf", input.VRF)
	if err != nil {
		return decodedPrefixInput{}, err
	}
	status, err := decodeVRFStringField("status", input.Status)
	if err != nil {
		return decodedPrefixInput{}, err
	}
	isPool, err := decodeVRFBoolField("is_pool", input.IsPool)
	if err != nil {
		return decodedPrefixInput{}, err
	}
	markUtilized, err := decodeVRFBoolField("mark_utilized", input.MarkUtilized)
	if err != nil {
		return decodedPrefixInput{}, err
	}
	description, err := decodeVRFStringField("description", input.Description)
	if err != nil {
		return decodedPrefixInput{}, err
	}
	comments, err := decodeVRFStringField("comments", input.Comments)
	if err != nil {
		return decodedPrefixInput{}, err
	}
	return decodedPrefixInput{
		prefix: prefix, vrf: vrf, status: status, isPool: isPool,
		markUtilized: markUtilized, description: description, comments: comments,
	}, nil
}

func decodePrefixInt64Field(name string, raw json.RawMessage) (applicationipam.Field[int64], error) {
	if len(raw) == 0 {
		return applicationipam.OmittedField[int64](), nil
	}
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return applicationipam.NullField[int64](), nil
	}
	var value int64
	if err := json.Unmarshal(raw, &value); err != nil {
		return applicationipam.Field[int64]{}, shared.Invalid(name, "A valid object ID is required.")
	}
	return applicationipam.FieldValue(value), nil
}

func (input decodedPrefixInput) createCommand() applicationipam.CreatePrefixCommand {
	return applicationipam.CreatePrefixCommand{
		Prefix: input.prefix, VRF: input.vrf, Status: input.status,
		IsPool: input.isPool, MarkUtilized: input.markUtilized,
		Description: input.description, Comments: input.comments,
	}
}

func (input decodedPrefixInput) replaceCommand(id shared.ID) applicationipam.ReplacePrefixCommand {
	return applicationipam.ReplacePrefixCommand{
		ID: id, CreatePrefixCommand: input.createCommand(),
	}
}

func (input decodedPrefixInput) updateCommand(id shared.ID) applicationipam.UpdatePrefixCommand {
	return applicationipam.UpdatePrefixCommand{
		ID: id, Prefix: input.prefix, VRF: input.vrf, Status: input.status,
		IsPool: input.isPool, MarkUtilized: input.markUtilized,
		Description: input.description, Comments: input.comments,
	}
}

func parsePrefixList(values url.Values) (applicationipam.ListPrefixesQuery, error) {
	allowed := map[string]struct{}{
		"limit": {}, "offset": {}, "q": {}, "ordering": {}, "id": {},
		"vrf_id": {}, "vrf_rd": {}, "prefix": {}, "family": {}, "status": {},
		"within": {}, "within_include": {}, "contains": {},
	}
	for key := range values {
		if _, ok := allowed[key]; !ok {
			return applicationipam.ListPrefixesQuery{}, shared.Invalid(key, "Unsupported filter.")
		}
	}
	query := applicationipam.ListPrefixesQuery{}
	if raw := values.Get("limit"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value < 0 || value > math.MaxUint32 {
			return query, shared.Invalid("limit", "A valid integer is required.")
		}
		query.Limit, query.LimitPresent = uint32(value), true
	}
	if raw := values.Get("offset"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value < 0 || value > math.MaxUint32 {
			return query, shared.Invalid("offset", "A valid integer is required.")
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
	for _, raw := range values["ordering"] {
		query.Ordering = append(query.Ordering, strings.Split(raw, ",")...)
	}
	query.VRFRDs = append([]string(nil), values["vrf_rd"]...)
	query.Prefixes = append([]string(nil), values["prefix"]...)
	query.Statuses = append([]string(nil), values["status"]...)
	if raw := values.Get("family"); raw != "" {
		value, parseErr := strconv.ParseInt(raw, 10, 64)
		if parseErr != nil {
			return query, shared.Invalid("family", "A valid integer is required.")
		}
		query.Family = &value
	}
	query.Within = optionalQueryString(values, "within")
	query.WithinInclude = optionalQueryString(values, "within_include")
	query.Contains = optionalQueryString(values, "contains")
	return query, nil
}

func parseRepeatedSignedIntegers(values []string) ([]int64, error) {
	var parsed []int64
	for _, raw := range values {
		for _, part := range strings.Split(raw, ",") {
			value, err := strconv.ParseInt(part, 10, 64)
			if err != nil {
				return nil, err
			}
			parsed = append(parsed, value)
		}
	}
	return parsed, nil
}

func optionalQueryString(values url.Values, key string) *string {
	raw := values.Get(key)
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	return &raw
}

func prefixPageURL(
	c *gin.Context,
	query applicationipam.ListPrefixesQuery,
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

var _ PrefixService = (*applicationipam.PrefixService)(nil)
