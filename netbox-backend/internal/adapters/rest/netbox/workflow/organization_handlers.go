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

	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

type ManufacturerService interface {
	ListManufacturers(context.Context, identity.Principal, applicationdcim.ListManufacturersQuery) (applicationdcim.ManufacturerPage, error)
	GetManufacturer(context.Context, identity.Principal, applicationdcim.GetManufacturerQuery) (*domaindcim.Manufacturer, error)
	CreateManufacturer(context.Context, identity.Principal, applicationdcim.CreateManufacturerCommand) (*domaindcim.Manufacturer, error)
	ReplaceManufacturer(context.Context, identity.Principal, applicationdcim.ReplaceManufacturerCommand) (*domaindcim.Manufacturer, error)
	UpdateManufacturer(context.Context, identity.Principal, applicationdcim.UpdateManufacturerCommand) (*domaindcim.Manufacturer, error)
	DeleteManufacturer(context.Context, identity.Principal, applicationdcim.DeleteManufacturerCommand) error
}

type RackRoleService interface {
	ListRackRoles(context.Context, identity.Principal, applicationdcim.ListRackRolesQuery) (applicationdcim.RackRolePage, error)
	GetRackRole(context.Context, identity.Principal, applicationdcim.GetRackRoleQuery) (*domaindcim.RackRole, error)
	CreateRackRole(context.Context, identity.Principal, applicationdcim.CreateRackRoleCommand) (*domaindcim.RackRole, error)
	ReplaceRackRole(context.Context, identity.Principal, applicationdcim.ReplaceRackRoleCommand) (*domaindcim.RackRole, error)
	UpdateRackRole(context.Context, identity.Principal, applicationdcim.UpdateRackRoleCommand) (*domaindcim.RackRole, error)
	DeleteRackRole(context.Context, identity.Principal, applicationdcim.DeleteRackRoleCommand) error
}

var (
	_ ManufacturerService = (*applicationdcim.ManufacturerService)(nil)
	_ RackRoleService     = (*applicationdcim.RackRoleService)(nil)
)

// OrganizationHandler owns the typed Manufacturer and RackRole HTTP surface.
// It is deliberately independent from Handler so composition can cut routes
// over without making these resources depend on the dynamic workflow service.
type OrganizationHandler struct {
	manufacturers ManufacturerService
	rackRoles     RackRoleService
}

func NewOrganizationHandler(
	manufacturers ManufacturerService,
	rackRoles RackRoleService,
) *OrganizationHandler {
	if manufacturers == nil || rackRoles == nil {
		panic("REST organization handler requires typed Manufacturer and RackRole services")
	}
	return &OrganizationHandler{manufacturers: manufacturers, rackRoles: rackRoles}
}

func (handler *OrganizationHandler) Register(r gin.IRoutes, middlewares ...gin.HandlerFunc) {
	handler.registerManufacturers(r, middlewares...)
	handler.registerRackRoles(r, middlewares...)
}

func (handler *OrganizationHandler) registerManufacturers(r gin.IRoutes, middlewares ...gin.HandlerFunc) {
	const base = "/api/dcim/manufacturers/"
	r.GET(base, append(middlewares, handler.listManufacturers())...)
	r.POST(base, append(middlewares, handler.createManufacturer())...)
	r.GET(base+":id/", append(middlewares, handler.getManufacturer())...)
	r.PUT(base+":id/", append(middlewares, handler.replaceManufacturer())...)
	r.PATCH(base+":id/", append(middlewares, handler.updateManufacturer())...)
	r.DELETE(base+":id/", append(middlewares, handler.deleteManufacturer())...)
}

func (handler *OrganizationHandler) registerRackRoles(r gin.IRoutes, middlewares ...gin.HandlerFunc) {
	const base = "/api/dcim/rack-roles/"
	r.GET(base, append(middlewares, handler.listRackRoles())...)
	r.POST(base, append(middlewares, handler.createRackRole())...)
	r.GET(base+":id/", append(middlewares, handler.getRackRole())...)
	r.PUT(base+":id/", append(middlewares, handler.replaceRackRole())...)
	r.PATCH(base+":id/", append(middlewares, handler.updateRackRole())...)
	r.DELETE(base+":id/", append(middlewares, handler.deleteRackRole())...)
}

type manufacturerDTO struct {
	ID              int64     `json:"id"`
	URL             string    `json:"url"`
	Display         string    `json:"display"`
	Name            string    `json:"name"`
	Slug            string    `json:"slug"`
	Description     string    `json:"description"`
	Created         time.Time `json:"created"`
	LastUpdated     time.Time `json:"last_updated"`
	DeviceTypeCount *uint64   `json:"devicetype_count,omitempty"`
}

type manufacturerListDTO struct {
	Count    uint64            `json:"count"`
	Next     *string           `json:"next"`
	Previous *string           `json:"previous"`
	Results  []manufacturerDTO `json:"results"`
}

func newManufacturerDTO(manufacturer *domaindcim.Manufacturer, includeRelatedCounts bool) manufacturerDTO {
	dto := manufacturerDTO{
		ID: manufacturer.ID().Int64(), URL: manufacturerURL(manufacturer.ID()),
		Display: manufacturer.Display(), Name: manufacturer.Name(), Slug: manufacturer.Slug().String(),
		Description: manufacturer.Description(), Created: manufacturer.Created().Time,
		LastUpdated: manufacturer.LastUpdated().Time,
	}
	if includeRelatedCounts {
		count := manufacturer.DeviceTypeCount()
		dto.DeviceTypeCount = &count
	}
	return dto
}

func manufacturerURL(id shared.ID) string {
	return "/api/dcim/manufacturers/" + id.String() + "/"
}

func (handler *OrganizationHandler) listManufacturers() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := Principal(c)
		if !ok {
			writeError(c, shared.Unauthenticated())
			return
		}
		query, err := parseManufacturerList(c.Request.URL.Query())
		if err != nil {
			writeError(c, err)
			return
		}
		page, err := handler.manufacturers.ListManufacturers(c.Request.Context(), principal, query)
		if err != nil {
			writeError(c, err)
			return
		}
		results := make([]manufacturerDTO, 0, len(page.Results))
		for _, manufacturer := range page.Results {
			results = append(results, newManufacturerDTO(manufacturer, true))
		}
		c.JSON(http.StatusOK, manufacturerListDTO{
			Count:    page.Count,
			Next:     organizationPageURL(c, query.Offset, query.EffectiveLimit(), page.Count, true),
			Previous: organizationPageURL(c, query.Offset, query.EffectiveLimit(), page.Count, false),
			Results:  results,
		})
	}
}

func (handler *OrganizationHandler) getManufacturer() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		manufacturer, err := handler.manufacturers.GetManufacturer(
			c.Request.Context(), principal, applicationdcim.GetManufacturerQuery{ID: id},
		)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newManufacturerDTO(manufacturer, true))
	}
}

func (handler *OrganizationHandler) createManufacturer() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := Principal(c)
		if !ok {
			writeError(c, shared.Unauthenticated())
			return
		}
		input, err := decodeManufacturerInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		manufacturer, err := handler.manufacturers.CreateManufacturer(
			c.Request.Context(), principal, input.createCommand(),
		)
		if err != nil {
			writeError(c, err)
			return
		}
		c.Header("Location", manufacturerURL(manufacturer.ID()))
		c.JSON(http.StatusCreated, newManufacturerDTO(manufacturer, false))
	}
}

func (handler *OrganizationHandler) replaceManufacturer() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		input, err := decodeManufacturerInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		manufacturer, err := handler.manufacturers.ReplaceManufacturer(
			c.Request.Context(), principal, input.replaceCommand(id),
		)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newManufacturerDTO(manufacturer, true))
	}
}

func (handler *OrganizationHandler) updateManufacturer() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		input, err := decodeManufacturerInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		manufacturer, err := handler.manufacturers.UpdateManufacturer(
			c.Request.Context(), principal, input.updateCommand(id),
		)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newManufacturerDTO(manufacturer, true))
	}
}

func (handler *OrganizationHandler) deleteManufacturer() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		if err := handler.manufacturers.DeleteManufacturer(
			c.Request.Context(), principal, applicationdcim.DeleteManufacturerCommand{ID: id},
		); err != nil {
			writeError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

type rackRoleDTO struct {
	ID          int64     `json:"id"`
	URL         string    `json:"url"`
	Display     string    `json:"display"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Color       string    `json:"color"`
	Description string    `json:"description"`
	Created     time.Time `json:"created"`
	LastUpdated time.Time `json:"last_updated"`
	RackCount   *uint64   `json:"rack_count,omitempty"`
}

type rackRoleListDTO struct {
	Count    uint64        `json:"count"`
	Next     *string       `json:"next"`
	Previous *string       `json:"previous"`
	Results  []rackRoleDTO `json:"results"`
}

func newRackRoleDTO(role *domaindcim.RackRole, includeRelatedCounts bool) rackRoleDTO {
	dto := rackRoleDTO{
		ID: role.ID().Int64(), URL: rackRoleURL(role.ID()), Display: role.Display(),
		Name: role.Name(), Slug: role.Slug().String(), Color: role.Color().String(),
		Description: role.Description(), Created: role.Created().Time, LastUpdated: role.LastUpdated().Time,
	}
	if includeRelatedCounts {
		count := role.RackCount()
		dto.RackCount = &count
	}
	return dto
}

func rackRoleURL(id shared.ID) string {
	return "/api/dcim/rack-roles/" + id.String() + "/"
}

func (handler *OrganizationHandler) listRackRoles() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := Principal(c)
		if !ok {
			writeError(c, shared.Unauthenticated())
			return
		}
		query, err := parseRackRoleList(c.Request.URL.Query())
		if err != nil {
			writeError(c, err)
			return
		}
		page, err := handler.rackRoles.ListRackRoles(c.Request.Context(), principal, query)
		if err != nil {
			writeError(c, err)
			return
		}
		results := make([]rackRoleDTO, 0, len(page.Results))
		for _, role := range page.Results {
			results = append(results, newRackRoleDTO(role, true))
		}
		c.JSON(http.StatusOK, rackRoleListDTO{
			Count:    page.Count,
			Next:     organizationPageURL(c, query.Offset, query.EffectiveLimit(), page.Count, true),
			Previous: organizationPageURL(c, query.Offset, query.EffectiveLimit(), page.Count, false),
			Results:  results,
		})
	}
}

func (handler *OrganizationHandler) getRackRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		role, err := handler.rackRoles.GetRackRole(
			c.Request.Context(), principal, applicationdcim.GetRackRoleQuery{ID: id},
		)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newRackRoleDTO(role, true))
	}
}

func (handler *OrganizationHandler) createRackRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := Principal(c)
		if !ok {
			writeError(c, shared.Unauthenticated())
			return
		}
		input, err := decodeRackRoleInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		role, err := handler.rackRoles.CreateRackRole(c.Request.Context(), principal, input.createCommand())
		if err != nil {
			writeError(c, err)
			return
		}
		c.Header("Location", rackRoleURL(role.ID()))
		c.JSON(http.StatusCreated, newRackRoleDTO(role, false))
	}
}

func (handler *OrganizationHandler) replaceRackRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		input, err := decodeRackRoleInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		role, err := handler.rackRoles.ReplaceRackRole(c.Request.Context(), principal, input.replaceCommand(id))
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newRackRoleDTO(role, true))
	}
}

func (handler *OrganizationHandler) updateRackRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		input, err := decodeRackRoleInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		role, err := handler.rackRoles.UpdateRackRole(c.Request.Context(), principal, input.updateCommand(id))
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newRackRoleDTO(role, true))
	}
}

func (handler *OrganizationHandler) deleteRackRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		if err := handler.rackRoles.DeleteRackRole(
			c.Request.Context(), principal, applicationdcim.DeleteRackRoleCommand{ID: id},
		); err != nil {
			writeError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func organizationRequestIdentity(c *gin.Context) (identity.Principal, shared.ID, bool) {
	principal, ok := Principal(c)
	if !ok {
		writeError(c, shared.Unauthenticated())
		return identity.Principal{}, 0, false
	}
	id, err := parseID(c.Param("id"))
	if err != nil {
		writeError(c, err)
		return identity.Principal{}, 0, false
	}
	return principal, shared.ID(id), true
}

type manufacturerInputJSON struct {
	Name        json.RawMessage `json:"name"`
	Slug        json.RawMessage `json:"slug"`
	Description json.RawMessage `json:"description"`
}

type decodedManufacturerInput struct {
	name        applicationdcim.Field[string]
	slug        applicationdcim.Field[string]
	description applicationdcim.Field[string]
}

func decodeManufacturerInput(c *gin.Context) (decodedManufacturerInput, error) {
	input, err := decodeTypedObject[manufacturerInputJSON](c)
	if err != nil {
		return decodedManufacturerInput{}, err
	}
	name, err := decodeSiteStringField("name", input.Name)
	if err != nil {
		return decodedManufacturerInput{}, err
	}
	slug, err := decodeSiteStringField("slug", input.Slug)
	if err != nil {
		return decodedManufacturerInput{}, err
	}
	description, err := decodeSiteStringField("description", input.Description)
	if err != nil {
		return decodedManufacturerInput{}, err
	}
	return decodedManufacturerInput{name: name, slug: slug, description: description}, nil
}

func (input decodedManufacturerInput) createCommand() applicationdcim.CreateManufacturerCommand {
	return applicationdcim.CreateManufacturerCommand{
		Name: input.name, Slug: input.slug, Description: input.description,
	}
}

func (input decodedManufacturerInput) replaceCommand(id shared.ID) applicationdcim.ReplaceManufacturerCommand {
	return applicationdcim.ReplaceManufacturerCommand{
		ID: id, Name: input.name, Slug: input.slug, Description: input.description,
	}
}

func (input decodedManufacturerInput) updateCommand(id shared.ID) applicationdcim.UpdateManufacturerCommand {
	return applicationdcim.UpdateManufacturerCommand{
		ID: id, Name: input.name, Slug: input.slug, Description: input.description,
	}
}

type rackRoleInputJSON struct {
	Name        json.RawMessage `json:"name"`
	Slug        json.RawMessage `json:"slug"`
	Color       json.RawMessage `json:"color"`
	Description json.RawMessage `json:"description"`
}

type decodedRackRoleInput struct {
	name        applicationdcim.Field[string]
	slug        applicationdcim.Field[string]
	color       applicationdcim.Field[string]
	description applicationdcim.Field[string]
}

func decodeRackRoleInput(c *gin.Context) (decodedRackRoleInput, error) {
	input, err := decodeTypedObject[rackRoleInputJSON](c)
	if err != nil {
		return decodedRackRoleInput{}, err
	}
	name, err := decodeSiteStringField("name", input.Name)
	if err != nil {
		return decodedRackRoleInput{}, err
	}
	slug, err := decodeSiteStringField("slug", input.Slug)
	if err != nil {
		return decodedRackRoleInput{}, err
	}
	color, err := decodeSiteStringField("color", input.Color)
	if err != nil {
		return decodedRackRoleInput{}, err
	}
	description, err := decodeSiteStringField("description", input.Description)
	if err != nil {
		return decodedRackRoleInput{}, err
	}
	return decodedRackRoleInput{name: name, slug: slug, color: color, description: description}, nil
}

func (input decodedRackRoleInput) createCommand() applicationdcim.CreateRackRoleCommand {
	return applicationdcim.CreateRackRoleCommand{
		Name: input.name, Slug: input.slug, Color: input.color, Description: input.description,
	}
}

func (input decodedRackRoleInput) replaceCommand(id shared.ID) applicationdcim.ReplaceRackRoleCommand {
	return applicationdcim.ReplaceRackRoleCommand{
		ID: id, Name: input.name, Slug: input.slug, Color: input.color, Description: input.description,
	}
}

func (input decodedRackRoleInput) updateCommand(id shared.ID) applicationdcim.UpdateRackRoleCommand {
	return applicationdcim.UpdateRackRoleCommand{
		ID: id, Name: input.name, Slug: input.slug, Color: input.color, Description: input.description,
	}
}

func decodeTypedObject[T any](c *gin.Context) (T, error) {
	var zero T
	raw, err := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20))
	if err != nil {
		return zero, shared.Invalid("non_field_errors", "Expected a JSON object.")
	}
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return zero, shared.Invalid("non_field_errors", "Expected a JSON object.")
	}
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.DisallowUnknownFields()
	var input T
	if err := decoder.Decode(&input); err != nil {
		if field, unknown := unknownJSONField(err); unknown {
			return zero, shared.Invalid(
				field, "This field is not supported by the active capability profile.")

		}
		return zero, shared.Invalid("non_field_errors", "Expected a JSON object.")
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return zero, shared.Invalid("non_field_errors", "Request body must contain one JSON object.")
	}
	return input, nil
}

type organizationListParameters struct {
	limit        uint32
	limitPresent bool
	offset       uint32
	query        string
	ids          []int64
	ordering     []string
	names        []string
	slugs        []string
}

func parseManufacturerList(values url.Values) (applicationdcim.ListManufacturersQuery, error) {
	parameters, err := parseOrganizationList(values)
	if err != nil {
		return applicationdcim.ListManufacturersQuery{}, err
	}
	return applicationdcim.ListManufacturersQuery{
		Limit: parameters.limit, LimitPresent: parameters.limitPresent,
		Offset: parameters.offset, Query: parameters.query, IDs: parameters.ids,
		Ordering: parameters.ordering, Names: parameters.names, Slugs: parameters.slugs,
	}, nil
}

func parseRackRoleList(values url.Values) (applicationdcim.ListRackRolesQuery, error) {
	parameters, err := parseOrganizationList(values)
	if err != nil {
		return applicationdcim.ListRackRolesQuery{}, err
	}
	return applicationdcim.ListRackRolesQuery{
		Limit: parameters.limit, LimitPresent: parameters.limitPresent,
		Offset: parameters.offset, Query: parameters.query, IDs: parameters.ids,
		Ordering: parameters.ordering, Names: parameters.names, Slugs: parameters.slugs,
	}, nil
}

func parseOrganizationList(values url.Values) (organizationListParameters, error) {
	allowed := map[string]struct{}{
		"limit": {}, "offset": {}, "q": {}, "ordering": {}, "id": {}, "name": {}, "slug": {},
	}
	for key := range values {
		if _, ok := allowed[key]; !ok {
			return organizationListParameters{}, shared.Invalid(key, "Unsupported filter.")
		}
	}
	parameters := organizationListParameters{}
	if raw := values.Get("limit"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value < 0 || value > math.MaxUint32 {
			return parameters, shared.Invalid("limit", "A valid integer is required.")
		}
		parameters.limit = uint32(value)
		parameters.limitPresent = true
	}
	if raw := values.Get("offset"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value < 0 || value > math.MaxUint32 {
			return parameters, shared.Invalid("offset", "A valid integer is required.")
		}
		parameters.offset = uint32(value)
	}
	parameters.query = values.Get("q")
	for _, raw := range values["id"] {
		for _, part := range strings.Split(raw, ",") {
			value, err := strconv.ParseInt(part, 10, 64)
			if err != nil {
				return parameters, shared.Invalid("id", "A valid integer is required.")
			}
			parameters.ids = append(parameters.ids, value)
		}
	}
	for _, raw := range values["ordering"] {
		parameters.ordering = append(parameters.ordering, strings.Split(raw, ",")...)
	}
	parameters.names = append([]string(nil), values["name"]...)
	parameters.slugs = append([]string(nil), values["slug"]...)
	return parameters, nil
}

func organizationPageURL(c *gin.Context, offset uint32, limit uint32, count uint64, next bool) *string {
	pageOffset := uint64(offset)
	pageLimit := uint64(limit)
	if next {
		pageOffset += pageLimit
		if pageOffset >= count {
			return nil
		}
	} else {
		if pageOffset == 0 {
			return nil
		}
		if pageOffset <= pageLimit {
			pageOffset = 0
		} else {
			pageOffset -= pageLimit
		}
	}
	values := c.Request.URL.Query()
	values.Set("limit", strconv.FormatUint(pageLimit, 10))
	if pageOffset == 0 {
		values.Del("offset")
	} else {
		values.Set("offset", strconv.FormatUint(pageOffset, 10))
	}
	pageURL := c.Request.URL.Path + "?" + values.Encode()
	return &pageURL
}
