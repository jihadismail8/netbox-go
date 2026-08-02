package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
	"math/big"
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

// RackTypeService is the typed RackType application boundary shared by REST
// and composition. No transitional workflow map crosses this interface.
type RackTypeService interface {
	ListRackTypes(context.Context, identity.Principal, applicationdcim.ListRackTypesQuery) (applicationdcim.RackTypePage, error)
	GetRackType(context.Context, identity.Principal, applicationdcim.GetRackTypeQuery) (*domaindcim.RackType, error)
	CreateRackType(context.Context, identity.Principal, applicationdcim.CreateRackTypeCommand) (*domaindcim.RackType, error)
	ReplaceRackType(context.Context, identity.Principal, applicationdcim.ReplaceRackTypeCommand) (*domaindcim.RackType, error)
	UpdateRackType(context.Context, identity.Principal, applicationdcim.UpdateRackTypeCommand) (*domaindcim.RackType, error)
	DeleteRackType(context.Context, identity.Principal, applicationdcim.DeleteRackTypeCommand) error
}

var _ RackTypeService = (*applicationdcim.RackTypeService)(nil)

// RackTypeRESTHandler is separately registrable so composition can cut the
// six RackType routes over atomically without changing unrelated resources.
type RackTypeRESTHandler struct{ service RackTypeService }

func NewRackTypeRESTHandler(service RackTypeService) *RackTypeRESTHandler {
	if service == nil {
		panic("REST RackType handler requires a typed RackType service")
	}
	return &RackTypeRESTHandler{service: service}
}

func (handler *RackTypeRESTHandler) Register(r gin.IRoutes, middlewares ...gin.HandlerFunc) {
	const base = "/api/dcim/rack-types/"
	r.GET(base, append(middlewares, handler.list())...)
	r.POST(base, append(middlewares, handler.create())...)
	r.GET(base+":id/", append(middlewares, handler.get())...)
	r.PUT(base+":id/", append(middlewares, handler.replace())...)
	r.PATCH(base+":id/", append(middlewares, handler.update())...)
	r.DELETE(base+":id/", append(middlewares, handler.delete())...)
}

type rackTypeObjectReferenceDTO struct {
	ID      int64  `json:"id"`
	URL     string `json:"url"`
	Display string `json:"display"`
}

type rackTypeStringChoiceDTO struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type rackTypeIntegerChoiceDTO struct {
	Value uint32 `json:"value"`
	Label string `json:"label"`
}

type rackTypeDTO struct {
	ID           int64                      `json:"id"`
	URL          string                     `json:"url"`
	Display      string                     `json:"display"`
	Manufacturer rackTypeObjectReferenceDTO `json:"manufacturer"`
	Model        string                     `json:"model"`
	Slug         string                     `json:"slug"`
	FormFactor   rackTypeStringChoiceDTO    `json:"form_factor"`
	Width        rackTypeIntegerChoiceDTO   `json:"width"`
	UHeight      uint32                     `json:"u_height"`
	StartingUnit uint32                     `json:"starting_unit"`
	DescUnits    bool                       `json:"desc_units"`
	Description  string                     `json:"description"`
	Comments     string                     `json:"comments"`
	Created      time.Time                  `json:"created"`
	LastUpdated  time.Time                  `json:"last_updated"`
}

type rackTypeListDTO struct {
	Count    uint64        `json:"count"`
	Next     *string       `json:"next"`
	Previous *string       `json:"previous"`
	Results  []rackTypeDTO `json:"results"`
}

func newRackTypeDTO(rackType *domaindcim.RackType) rackTypeDTO {
	manufacturer := rackType.Manufacturer()
	return rackTypeDTO{
		ID: rackType.ID().Int64(), URL: rackTypeURL(rackType.ID()), Display: rackType.Display(),
		Manufacturer: rackTypeObjectReferenceDTO{
			ID: manufacturer.ID().Int64(), URL: manufacturerURL(manufacturer.ID()),
			Display: manufacturer.Display(),
		},
		Model: rackType.Model(), Slug: rackType.Slug().String(),
		FormFactor: rackTypeStringChoiceDTO{
			Value: rackType.FormFactor().String(), Label: rackTypeFormFactorLabel(rackType.FormFactor()),
		},
		Width: rackTypeIntegerChoiceDTO{
			Value: rackType.Width().Uint32(), Label: strconv.FormatUint(uint64(rackType.Width().Uint32()), 10) + " inches",
		},
		UHeight: rackType.UHeight(), StartingUnit: rackType.StartingUnit(), DescUnits: rackType.DescUnits(),
		Description: rackType.Description(), Comments: rackType.Comments(),
		Created: rackType.Created().Time, LastUpdated: rackType.LastUpdated().Time,
	}
}

func rackTypeFormFactorLabel(value domaindcim.RackFormFactor) string {
	switch value {
	case domaindcim.RackFormFactorTwoPostFrame:
		return "2-post frame"
	case domaindcim.RackFormFactorFourPostFrame:
		return "4-post frame"
	case domaindcim.RackFormFactorFourPostCabinet:
		return "4-post cabinet"
	case domaindcim.RackFormFactorWallFrame:
		return "Wall-mounted frame"
	case domaindcim.RackFormFactorWallFrameVertical:
		return "Wall-mounted frame (vertical)"
	case domaindcim.RackFormFactorWallCabinet:
		return "Wall-mounted cabinet"
	case domaindcim.RackFormFactorWallCabinetVertical:
		return "Wall-mounted cabinet (vertical)"
	default:
		return ""
	}
}

func rackTypeURL(id shared.ID) string {
	return "/api/dcim/rack-types/" + id.String() + "/"
}

func (handler *RackTypeRESTHandler) list() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := Principal(c)
		if !ok {
			writeError(c, shared.Unauthenticated())
			return
		}
		query, err := parseRackTypeList(c.Request.URL.Query())
		if err != nil {
			writeError(c, err)
			return
		}
		page, err := handler.service.ListRackTypes(c.Request.Context(), principal, query)
		if err != nil {
			writeError(c, err)
			return
		}
		results := make([]rackTypeDTO, 0, len(page.Results))
		for _, rackType := range page.Results {
			results = append(results, newRackTypeDTO(rackType))
		}
		c.JSON(http.StatusOK, rackTypeListDTO{
			Count:    page.Count,
			Next:     organizationPageURL(c, query.Offset, query.EffectiveLimit(), page.Count, true),
			Previous: organizationPageURL(c, query.Offset, query.EffectiveLimit(), page.Count, false),
			Results:  results,
		})
	}
}

func (handler *RackTypeRESTHandler) get() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		rackType, err := handler.service.GetRackType(
			c.Request.Context(), principal, applicationdcim.GetRackTypeQuery{ID: id},
		)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newRackTypeDTO(rackType))
	}
}

func (handler *RackTypeRESTHandler) create() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := Principal(c)
		if !ok {
			writeError(c, shared.Unauthenticated())
			return
		}
		input, err := decodeRackTypeInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		rackType, err := handler.service.CreateRackType(c.Request.Context(), principal, input.createCommand())
		if err != nil {
			writeError(c, err)
			return
		}
		c.Header("Location", rackTypeURL(rackType.ID()))
		c.JSON(http.StatusCreated, newRackTypeDTO(rackType))
	}
}

func (handler *RackTypeRESTHandler) replace() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		input, err := decodeRackTypeInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		rackType, err := handler.service.ReplaceRackType(c.Request.Context(), principal, input.replaceCommand(id))
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newRackTypeDTO(rackType))
	}
}

func (handler *RackTypeRESTHandler) update() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		input, err := decodeRackTypeInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		rackType, err := handler.service.UpdateRackType(c.Request.Context(), principal, input.updateCommand(id))
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newRackTypeDTO(rackType))
	}
}

func (handler *RackTypeRESTHandler) delete() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		if err := handler.service.DeleteRackType(
			c.Request.Context(), principal, applicationdcim.DeleteRackTypeCommand{ID: id},
		); err != nil {
			writeError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

type rackTypeInputJSON struct {
	Manufacturer json.RawMessage `json:"manufacturer"`
	Model        json.RawMessage `json:"model"`
	Slug         json.RawMessage `json:"slug"`
	FormFactor   json.RawMessage `json:"form_factor"`
	Width        json.RawMessage `json:"width"`
	UHeight      json.RawMessage `json:"u_height"`
	StartingUnit json.RawMessage `json:"starting_unit"`
	DescUnits    json.RawMessage `json:"desc_units"`
	Description  json.RawMessage `json:"description"`
	Comments     json.RawMessage `json:"comments"`
}

type decodedRackTypeInput struct {
	manufacturer applicationdcim.Field[shared.ID]
	model        applicationdcim.Field[string]
	slug         applicationdcim.Field[string]
	formFactor   applicationdcim.Field[string]
	width        applicationdcim.Field[uint32]
	uHeight      applicationdcim.Field[uint32]
	startingUnit applicationdcim.Field[uint32]
	descUnits    applicationdcim.Field[bool]
	description  applicationdcim.Field[string]
	comments     applicationdcim.Field[string]
}

func decodeRackTypeInput(c *gin.Context) (decodedRackTypeInput, error) {
	input, err := decodeTypedObject[rackTypeInputJSON](c)
	if err != nil {
		return decodedRackTypeInput{}, err
	}
	manufacturer, err := decodeRackTypeIDField("manufacturer", input.Manufacturer)
	if err != nil {
		return decodedRackTypeInput{}, err
	}
	model, err := decodeSiteStringField("model", input.Model)
	if err != nil {
		return decodedRackTypeInput{}, err
	}
	slug, err := decodeSiteStringField("slug", input.Slug)
	if err != nil {
		return decodedRackTypeInput{}, err
	}
	formFactor, err := decodeSiteStringField("form_factor", input.FormFactor)
	if err != nil {
		return decodedRackTypeInput{}, err
	}
	width, err := decodeRackTypeUint32Field("width", input.Width)
	if err != nil {
		return decodedRackTypeInput{}, err
	}
	uHeight, err := decodeRackTypeUint32Field("u_height", input.UHeight)
	if err != nil {
		return decodedRackTypeInput{}, err
	}
	startingUnit, err := decodeRackTypeUint32Field("starting_unit", input.StartingUnit)
	if err != nil {
		return decodedRackTypeInput{}, err
	}
	descUnits, err := decodeRackTypeBoolField("desc_units", input.DescUnits)
	if err != nil {
		return decodedRackTypeInput{}, err
	}
	description, err := decodeSiteStringField("description", input.Description)
	if err != nil {
		return decodedRackTypeInput{}, err
	}
	comments, err := decodeSiteStringField("comments", input.Comments)
	if err != nil {
		return decodedRackTypeInput{}, err
	}
	return decodedRackTypeInput{
		manufacturer: manufacturer, model: model, slug: slug, formFactor: formFactor,
		width: width, uHeight: uHeight, startingUnit: startingUnit, descUnits: descUnits,
		description: description, comments: comments,
	}, nil
}

func (input decodedRackTypeInput) createCommand() applicationdcim.CreateRackTypeCommand {
	return applicationdcim.CreateRackTypeCommand{
		Manufacturer: input.manufacturer, Model: input.model, Slug: input.slug,
		FormFactor: input.formFactor, Width: input.width, UHeight: input.uHeight,
		StartingUnit: input.startingUnit, DescUnits: input.descUnits,
		Description: input.description, Comments: input.comments,
	}
}

func (input decodedRackTypeInput) replaceCommand(id shared.ID) applicationdcim.ReplaceRackTypeCommand {
	return applicationdcim.ReplaceRackTypeCommand{ID: id, CreateRackTypeCommand: input.createCommand()}
}

func (input decodedRackTypeInput) updateCommand(id shared.ID) applicationdcim.UpdateRackTypeCommand {
	return applicationdcim.UpdateRackTypeCommand{
		ID: id, Manufacturer: input.manufacturer, Model: input.model, Slug: input.slug,
		FormFactor: input.formFactor, Width: input.width, UHeight: input.uHeight,
		StartingUnit: input.startingUnit, DescUnits: input.descUnits,
		Description: input.description, Comments: input.comments,
	}
}

func decodeRackTypeIDField(name string, raw json.RawMessage) (applicationdcim.Field[shared.ID], error) {
	if len(raw) == 0 {
		return applicationdcim.OmittedField[shared.ID](), nil
	}
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return applicationdcim.NullField[shared.ID](), nil
	}
	value, valid := exactJSONInteger(raw)
	if !valid || !value.IsInt64() {
		return applicationdcim.Field[shared.ID]{}, shared.Invalid(name, "A valid integer is required.")
	}
	return applicationdcim.FieldValue(shared.ID(value.Int64())), nil
}

func decodeRackTypeUint32Field(name string, raw json.RawMessage) (applicationdcim.Field[uint32], error) {
	if len(raw) == 0 {
		return applicationdcim.OmittedField[uint32](), nil
	}
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return applicationdcim.NullField[uint32](), nil
	}
	value, valid := exactJSONInteger(raw)
	if !valid || value.Sign() < 0 || value.BitLen() > 32 {
		return applicationdcim.Field[uint32]{}, shared.Invalid(name, "A valid integer is required.")
	}
	return applicationdcim.FieldValue(uint32(value.Uint64())), nil
}

func exactJSONInteger(raw json.RawMessage) (*big.Int, bool) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" {
		return nil, false
	}
	rational, valid := new(big.Rat).SetString(trimmed)
	if !valid || !rational.IsInt() {
		return nil, false
	}
	return new(big.Int).Set(rational.Num()), true
}

func decodeRackTypeBoolField(name string, raw json.RawMessage) (applicationdcim.Field[bool], error) {
	if len(raw) == 0 {
		return applicationdcim.OmittedField[bool](), nil
	}
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return applicationdcim.NullField[bool](), nil
	}
	var value bool
	if err := json.Unmarshal(raw, &value); err != nil {
		return applicationdcim.Field[bool]{}, shared.Invalid(name, "Must be a valid boolean.")
	}
	return applicationdcim.FieldValue(value), nil
}

func parseRackTypeList(values url.Values) (applicationdcim.ListRackTypesQuery, error) {
	allowed := map[string]struct{}{
		"limit": {}, "offset": {}, "q": {}, "ordering": {}, "id": {},
		"manufacturer_id": {}, "manufacturer_slug": {}, "model": {}, "slug": {},
	}
	for key := range values {
		if _, ok := allowed[key]; !ok {
			return applicationdcim.ListRackTypesQuery{}, shared.Invalid(key, "Unsupported filter.")
		}
	}
	query := applicationdcim.ListRackTypesQuery{}
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
	ids, err := parseRackTypeSignedFilters(values["id"])
	if err != nil {
		return query, shared.Invalid("id", "A valid integer is required.")
	}
	query.IDs = ids
	manufacturerIDs, err := parseRackTypeSignedFilters(values["manufacturer_id"])
	if err != nil {
		return query, shared.Invalid("manufacturer_id", "A valid integer is required.")
	}
	query.ManufacturerIDs = manufacturerIDs
	for _, raw := range values["ordering"] {
		query.Ordering = append(query.Ordering, strings.Split(raw, ",")...)
	}
	query.ManufacturerSlugs = append([]string(nil), values["manufacturer_slug"]...)
	query.Models = append([]string(nil), values["model"]...)
	query.Slugs = append([]string(nil), values["slug"]...)
	return query, nil
}

func parseRackTypeSignedFilters(values []string) ([]int64, error) {
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
