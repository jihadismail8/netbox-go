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

// VRFService is the typed application boundary used by both public
// transports. Transport code never projects a VRF through an untyped map.
type VRFService interface {
	ListVRFs(context.Context, identity.Principal, applicationipam.ListVRFsQuery) (applicationipam.VRFPage, error)
	GetVRF(context.Context, identity.Principal, applicationipam.GetVRFQuery) (*domainipam.VRF, error)
	CreateVRF(context.Context, identity.Principal, applicationipam.CreateVRFCommand) (*domainipam.VRF, error)
	ReplaceVRF(context.Context, identity.Principal, applicationipam.ReplaceVRFCommand) (*domainipam.VRF, error)
	UpdateVRF(context.Context, identity.Principal, applicationipam.UpdateVRFCommand) (*domainipam.VRF, error)
	DeleteVRF(context.Context, identity.Principal, applicationipam.DeleteVRFCommand) error
}

// VRFRESTHandler is separately registrable so composition can cut the VRF
// routes over atomically while the remaining profile resources still use the
// transitional workflow handler.
type VRFRESTHandler struct {
	service VRFService
}

func NewVRFRESTHandler(service VRFService) *VRFRESTHandler {
	if service == nil {
		panic("REST VRF handler requires a typed VRF service")
	}
	return &VRFRESTHandler{service: service}
}

func (handler *VRFRESTHandler) Register(r gin.IRoutes, middlewares ...gin.HandlerFunc) {
	const base = "/api/ipam/vrfs/"
	r.GET(base, append(middlewares, handler.list())...)
	r.POST(base, append(middlewares, handler.create())...)
	r.GET(base+":id/", append(middlewares, handler.get())...)
	r.PUT(base+":id/", append(middlewares, handler.replace())...)
	r.PATCH(base+":id/", append(middlewares, handler.update())...)
	r.DELETE(base+":id/", append(middlewares, handler.delete())...)
}

func (handler *VRFRESTHandler) list() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := Principal(c)
		if !ok {
			writeError(c, shared.Unauthenticated())
			return
		}
		query, err := parseVRFList(c.Request.URL.Query())
		if err != nil {
			writeError(c, err)
			return
		}
		page, err := handler.service.ListVRFs(c.Request.Context(), principal, query)
		if err != nil {
			writeError(c, err)
			return
		}
		results := make([]vrfDTO, 0, len(page.Results))
		for _, vrf := range page.Results {
			results = append(results, newVRFDTO(vrf, true))
		}
		c.JSON(http.StatusOK, vrfListDTO{
			Count: page.Count, Next: vrfPageURL(c, query, page.Count, true),
			Previous: vrfPageURL(c, query, page.Count, false), Results: results,
		})
	}
}

func (handler *VRFRESTHandler) get() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := vrfRequestIdentity(c)
		if !ok {
			return
		}
		vrf, err := handler.service.GetVRF(c.Request.Context(), principal, applicationipam.GetVRFQuery{ID: id})
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newVRFDTO(vrf, true))
	}
}

func (handler *VRFRESTHandler) create() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := Principal(c)
		if !ok {
			writeError(c, shared.Unauthenticated())
			return
		}
		input, err := decodeVRFInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		vrf, err := handler.service.CreateVRF(c.Request.Context(), principal, input.createCommand())
		if err != nil {
			writeError(c, err)
			return
		}
		c.Header("Location", vrfURL(vrf.ID()))
		c.JSON(http.StatusCreated, newVRFDTO(vrf, false))
	}
}

func (handler *VRFRESTHandler) replace() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := vrfRequestIdentity(c)
		if !ok {
			return
		}
		input, err := decodeVRFInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		vrf, err := handler.service.ReplaceVRF(c.Request.Context(), principal, input.replaceCommand(id))
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newVRFDTO(vrf, true))
	}
}

func (handler *VRFRESTHandler) update() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := vrfRequestIdentity(c)
		if !ok {
			return
		}
		input, err := decodeVRFInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		vrf, err := handler.service.UpdateVRF(c.Request.Context(), principal, input.updateCommand(id))
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newVRFDTO(vrf, true))
	}
}

func (handler *VRFRESTHandler) delete() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := vrfRequestIdentity(c)
		if !ok {
			return
		}
		if err := handler.service.DeleteVRF(
			c.Request.Context(), principal, applicationipam.DeleteVRFCommand{ID: id},
		); err != nil {
			writeError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

type vrfListDTO struct {
	Count    uint64   `json:"count"`
	Next     *string  `json:"next"`
	Previous *string  `json:"previous"`
	Results  []vrfDTO `json:"results"`
}

type vrfDTO struct {
	ID             int64     `json:"id"`
	URL            string    `json:"url"`
	Display        string    `json:"display"`
	Name           string    `json:"name"`
	RD             *string   `json:"rd"`
	EnforceUnique  bool      `json:"enforce_unique"`
	Description    string    `json:"description"`
	Comments       string    `json:"comments"`
	Created        time.Time `json:"created"`
	LastUpdated    time.Time `json:"last_updated"`
	IPAddressCount *uint64   `json:"ipaddress_count,omitempty"`
	PrefixCount    *uint64   `json:"prefix_count,omitempty"`
}

func newVRFDTO(vrf *domainipam.VRF, includeRelatedCounts bool) vrfDTO {
	if vrf == nil {
		return vrfDTO{}
	}
	dto := vrfDTO{
		ID: vrf.ID().Int64(), URL: vrfURL(vrf.ID()), Display: vrf.Display(),
		Name: vrf.Name(), RD: vrfRDPointer(vrf.RD()), EnforceUnique: vrf.EnforceUnique(),
		Description: vrf.Description(), Comments: vrf.Comments(),
		Created: vrf.Created().Time, LastUpdated: vrf.LastUpdated().Time,
	}
	if includeRelatedCounts {
		ipAddressCount := vrf.IPAddressCount()
		prefixCount := vrf.PrefixCount()
		dto.IPAddressCount = &ipAddressCount
		dto.PrefixCount = &prefixCount
	}
	return dto
}

func vrfRDPointer(nullable domainipam.NullableRouteDistinguisher) *string {
	rd, present := nullable.Get()
	if !present {
		return nil
	}
	value := rd.String()
	return &value
}

func vrfURL(id shared.ID) string { return "/api/ipam/vrfs/" + id.String() + "/" }

func vrfRequestIdentity(c *gin.Context) (identity.Principal, shared.ID, bool) {
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

type vrfInputJSON struct {
	Name          json.RawMessage `json:"name"`
	RD            json.RawMessage `json:"rd"`
	EnforceUnique json.RawMessage `json:"enforce_unique"`
	Description   json.RawMessage `json:"description"`
	Comments      json.RawMessage `json:"comments"`
}

type decodedVRFInput struct {
	name          applicationipam.Field[string]
	rd            applicationipam.Field[string]
	enforceUnique applicationipam.Field[bool]
	description   applicationipam.Field[string]
	comments      applicationipam.Field[string]
}

func decodeVRFInput(c *gin.Context) (decodedVRFInput, error) {
	raw, err := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20))
	if err != nil {
		return decodedVRFInput{}, shared.Invalid("non_field_errors", "Expected a JSON object.")
	}
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return decodedVRFInput{}, shared.Invalid("non_field_errors", "Expected a JSON object.")
	}
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.DisallowUnknownFields()
	var input vrfInputJSON
	if err := decoder.Decode(&input); err != nil {
		if field, unknown := unknownJSONField(err); unknown {
			return decodedVRFInput{}, shared.Invalid(
				field, "This field is not supported by the active capability profile.")

		}
		return decodedVRFInput{}, shared.Invalid("non_field_errors", "Expected a JSON object.")
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return decodedVRFInput{}, shared.Invalid(
			"non_field_errors", "Request body must contain one JSON object.")

	}
	name, err := decodeVRFStringField("name", input.Name)
	if err != nil {
		return decodedVRFInput{}, err
	}
	rd, err := decodeVRFStringField("rd", input.RD)
	if err != nil {
		return decodedVRFInput{}, err
	}
	enforceUnique, err := decodeVRFBoolField("enforce_unique", input.EnforceUnique)
	if err != nil {
		return decodedVRFInput{}, err
	}
	description, err := decodeVRFStringField("description", input.Description)
	if err != nil {
		return decodedVRFInput{}, err
	}
	comments, err := decodeVRFStringField("comments", input.Comments)
	if err != nil {
		return decodedVRFInput{}, err
	}
	return decodedVRFInput{
		name: name, rd: rd, enforceUnique: enforceUnique,
		description: description, comments: comments,
	}, nil
}

func decodeVRFStringField(name string, raw json.RawMessage) (applicationipam.Field[string], error) {
	if len(raw) == 0 {
		return applicationipam.OmittedField[string](), nil
	}
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return applicationipam.NullField[string](), nil
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return applicationipam.Field[string]{}, shared.Invalid(name, "Must be a string.")
	}
	return applicationipam.FieldValue(value), nil
}

func decodeVRFBoolField(name string, raw json.RawMessage) (applicationipam.Field[bool], error) {
	if len(raw) == 0 {
		return applicationipam.OmittedField[bool](), nil
	}
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return applicationipam.NullField[bool](), nil
	}
	var value bool
	if err := json.Unmarshal(raw, &value); err != nil {
		return applicationipam.Field[bool]{}, shared.Invalid(name, "Must be a valid boolean.")
	}
	return applicationipam.FieldValue(value), nil
}

func (input decodedVRFInput) createCommand() applicationipam.CreateVRFCommand {
	return applicationipam.CreateVRFCommand{
		Name: input.name, RD: input.rd, EnforceUnique: input.enforceUnique,
		Description: input.description, Comments: input.comments,
	}
}

func (input decodedVRFInput) replaceCommand(id shared.ID) applicationipam.ReplaceVRFCommand {
	return applicationipam.ReplaceVRFCommand{
		ID: id, Name: input.name, RD: input.rd, EnforceUnique: input.enforceUnique,
		Description: input.description, Comments: input.comments,
	}
}

func (input decodedVRFInput) updateCommand(id shared.ID) applicationipam.UpdateVRFCommand {
	return applicationipam.UpdateVRFCommand{
		ID: id, Name: input.name, RD: input.rd, EnforceUnique: input.enforceUnique,
		Description: input.description, Comments: input.comments,
	}
}

func parseVRFList(values url.Values) (applicationipam.ListVRFsQuery, error) {
	allowed := map[string]struct{}{
		"limit": {}, "offset": {}, "q": {}, "ordering": {}, "id": {},
		"name": {}, "rd": {}, "enforce_unique": {},
	}
	for key := range values {
		if _, ok := allowed[key]; !ok {
			return applicationipam.ListVRFsQuery{}, shared.Invalid(key, "Unsupported filter.")
		}
	}

	query := applicationipam.ListVRFsQuery{}
	if raw := values.Get("limit"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value < 0 || value > math.MaxUint32 {
			return query, shared.Invalid("limit", "A valid integer is required.")
		}
		query.Limit = uint32(value)
		query.LimitPresent = true
	}
	if raw := values.Get("offset"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value < 0 || value > math.MaxUint32 {
			return query, shared.Invalid("offset", "A valid integer is required.")
		}
		query.Offset = uint32(value)
	}
	query.Query = values.Get("q")
	for _, raw := range values["id"] {
		for _, part := range strings.Split(raw, ",") {
			value, err := strconv.ParseInt(part, 10, 64)
			if err != nil {
				return query, shared.Invalid("id", "A valid integer is required.")
			}
			query.IDs = append(query.IDs, value)
		}
	}
	for _, raw := range values["ordering"] {
		query.Ordering = append(query.Ordering, strings.Split(raw, ",")...)
	}
	query.Names = append([]string(nil), values["name"]...)
	query.RDs = append([]string(nil), values["rd"]...)
	if raw, present := values["enforce_unique"]; present && len(raw) > 0 && raw[len(raw)-1] != "" {
		value, err := strconv.ParseBool(raw[len(raw)-1])
		if err != nil {
			return query, shared.Invalid("enforce_unique", "A valid boolean is required.")
		}
		query.EnforceUnique = &value
	}
	return query, nil
}

func vrfPageURL(
	c *gin.Context,
	query applicationipam.ListVRFsQuery,
	count uint64,
	next bool,
) *string {
	offset := uint64(query.Offset)
	limit := uint64(query.EffectiveLimit())
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

var _ VRFService = (*applicationipam.VRFService)(nil)
