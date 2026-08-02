package workflow

import (
	"bytes"
	"context"
	"encoding/json"
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

type RackService interface {
	ListRacks(context.Context, identity.Principal, applicationdcim.ListRacksQuery) (applicationdcim.RackPage, error)
	GetRack(context.Context, identity.Principal, applicationdcim.GetRackQuery) (*domaindcim.Rack, error)
	CreateRack(context.Context, identity.Principal, applicationdcim.CreateRackCommand) (*domaindcim.Rack, error)
	ReplaceRack(context.Context, identity.Principal, applicationdcim.ReplaceRackCommand) (*domaindcim.Rack, error)
	UpdateRack(context.Context, identity.Principal, applicationdcim.UpdateRackCommand) (*domaindcim.Rack, error)
	DeleteRack(context.Context, identity.Principal, applicationdcim.DeleteRackCommand) error
}

var _ RackService = (*applicationdcim.RackService)(nil)

type RackRESTHandler struct{ service RackService }

func NewRackRESTHandler(service RackService) *RackRESTHandler {
	if service == nil {
		panic("REST Rack handler requires a typed Rack service")
	}
	return &RackRESTHandler{service: service}
}

func (handler *RackRESTHandler) Register(r gin.IRoutes, middlewares ...gin.HandlerFunc) {
	const base = "/api/dcim/racks/"
	r.GET(base, append(middlewares, handler.list())...)
	r.POST(base, append(middlewares, handler.create())...)
	r.GET(base+":id/", append(middlewares, handler.get())...)
	r.PUT(base+":id/", append(middlewares, handler.replace())...)
	r.PATCH(base+":id/", append(middlewares, handler.update())...)
	r.DELETE(base+":id/", append(middlewares, handler.delete())...)
}

type rackDTO struct {
	ID           int64                       `json:"id"`
	URL          string                      `json:"url"`
	Display      string                      `json:"display"`
	Site         rackTypeObjectReferenceDTO  `json:"site"`
	Name         string                      `json:"name"`
	FacilityID   *string                     `json:"facility_id"`
	RackType     *rackTypeObjectReferenceDTO `json:"rack_type"`
	Status       rackTypeStringChoiceDTO     `json:"status"`
	Role         *rackTypeObjectReferenceDTO `json:"role"`
	Serial       string                      `json:"serial"`
	AssetTag     *string                     `json:"asset_tag"`
	FormFactor   *rackTypeStringChoiceDTO    `json:"form_factor"`
	Width        rackTypeIntegerChoiceDTO    `json:"width"`
	UHeight      uint32                      `json:"u_height"`
	StartingUnit uint32                      `json:"starting_unit"`
	DescUnits    bool                        `json:"desc_units"`
	Airflow      *rackTypeStringChoiceDTO    `json:"airflow"`
	Description  string                      `json:"description"`
	Comments     string                      `json:"comments"`
	Created      time.Time                   `json:"created"`
	LastUpdated  time.Time                   `json:"last_updated"`
	DeviceCount  *uint64                     `json:"device_count,omitempty"`
}

type rackListDTO struct {
	Count    uint64    `json:"count"`
	Next     *string   `json:"next"`
	Previous *string   `json:"previous"`
	Results  []rackDTO `json:"results"`
}

func newRackDTO(rack *domaindcim.Rack, includeRelatedCounts bool) rackDTO {
	site := rack.Site()
	dto := rackDTO{
		ID: rack.ID().Int64(), URL: rackURL(rack.ID()), Display: rack.Display(),
		Site: rackTypeObjectReferenceDTO{
			ID: site.ID().Int64(), URL: siteURL(site.ID()), Display: site.Display(),
		},
		Name: rack.Name(), FacilityID: nullableRackDTOString(rack.FacilityID()),
		Status: rackTypeStringChoiceDTO{
			Value: rack.Status().String(), Label: commonChoiceLabels[rack.Status().String()],
		},
		Serial: rack.Serial(), AssetTag: nullableRackDTOString(rack.AssetTag()),
		Width: rackTypeIntegerChoiceDTO{
			Value: rack.Width().Uint32(),
			Label: strconv.FormatUint(uint64(rack.Width().Uint32()), 10) + " inches",
		},
		UHeight: rack.UHeight(), StartingUnit: rack.StartingUnit(), DescUnits: rack.DescUnits(),
		Description: rack.Description(), Comments: rack.Comments(),
		Created: rack.Created().Time, LastUpdated: rack.LastUpdated().Time,
	}
	if rackType, present := rack.RackType().Get(); present {
		dto.RackType = &rackTypeObjectReferenceDTO{
			ID: rackType.ID().Int64(), URL: rackTypeURL(rackType.ID()), Display: rackType.Display(),
		}
	}
	if role, present := rack.Role().Get(); present {
		dto.Role = &rackTypeObjectReferenceDTO{
			ID: role.ID().Int64(), URL: rackRoleURL(role.ID()), Display: role.Display(),
		}
	}
	if formFactor, present := rack.FormFactor().Get(); present && formFactor.String() != "" {
		dto.FormFactor = &rackTypeStringChoiceDTO{
			Value: formFactor.String(), Label: rackTypeFormFactorLabel(formFactor),
		}
	}
	if airflow, present := rack.Airflow().Get(); present && airflow.String() != "" {
		dto.Airflow = &rackTypeStringChoiceDTO{
			Value: airflow.String(), Label: airflowLabels[airflow.String()],
		}
	}
	if includeRelatedCounts {
		count := rack.DeviceCount()
		dto.DeviceCount = &count
	}
	return dto
}

func nullableRackDTOString(value domaindcim.RackNullable[string]) *string {
	text, present := value.Get()
	if !present {
		return nil
	}
	return &text
}

func rackURL(id shared.ID) string {
	return "/api/dcim/racks/" + id.String() + "/"
}

func (handler *RackRESTHandler) list() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := Principal(c)
		if !ok {
			writeError(c, shared.Unauthenticated())
			return
		}
		query, err := parseRackList(c.Request.URL.Query())
		if err != nil {
			writeError(c, err)
			return
		}
		page, err := handler.service.ListRacks(c.Request.Context(), principal, query)
		if err != nil {
			writeError(c, err)
			return
		}
		results := make([]rackDTO, 0, len(page.Results))
		for _, rack := range page.Results {
			results = append(results, newRackDTO(rack, true))
		}
		c.JSON(http.StatusOK, rackListDTO{
			Count:    page.Count,
			Next:     organizationPageURL(c, query.Offset, query.EffectiveLimit(), page.Count, true),
			Previous: organizationPageURL(c, query.Offset, query.EffectiveLimit(), page.Count, false),
			Results:  results,
		})
	}
}

func (handler *RackRESTHandler) get() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		rack, err := handler.service.GetRack(
			c.Request.Context(), principal, applicationdcim.GetRackQuery{ID: id},
		)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newRackDTO(rack, true))
	}
}

func (handler *RackRESTHandler) create() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := Principal(c)
		if !ok {
			writeError(c, shared.Unauthenticated())
			return
		}
		input, err := decodeRackInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		rack, err := handler.service.CreateRack(c.Request.Context(), principal, input.createCommand())
		if err != nil {
			writeError(c, err)
			return
		}
		c.Header("Location", rackURL(rack.ID()))
		c.JSON(http.StatusCreated, newRackDTO(rack, false))
	}
}

func (handler *RackRESTHandler) replace() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		input, err := decodeRackInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		rack, err := handler.service.ReplaceRack(c.Request.Context(), principal, input.replaceCommand(id))
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newRackDTO(rack, true))
	}
}

func (handler *RackRESTHandler) update() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		input, err := decodeRackInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		rack, err := handler.service.UpdateRack(c.Request.Context(), principal, input.updateCommand(id))
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newRackDTO(rack, true))
	}
}

func (handler *RackRESTHandler) delete() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		if err := handler.service.DeleteRack(
			c.Request.Context(), principal, applicationdcim.DeleteRackCommand{ID: id},
		); err != nil {
			writeError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

type rackInputJSON struct {
	Site         json.RawMessage `json:"site"`
	Name         json.RawMessage `json:"name"`
	FacilityID   json.RawMessage `json:"facility_id"`
	RackType     json.RawMessage `json:"rack_type"`
	Status       json.RawMessage `json:"status"`
	Role         json.RawMessage `json:"role"`
	Serial       json.RawMessage `json:"serial"`
	AssetTag     json.RawMessage `json:"asset_tag"`
	FormFactor   json.RawMessage `json:"form_factor"`
	Width        json.RawMessage `json:"width"`
	UHeight      json.RawMessage `json:"u_height"`
	StartingUnit json.RawMessage `json:"starting_unit"`
	DescUnits    json.RawMessage `json:"desc_units"`
	Airflow      json.RawMessage `json:"airflow"`
	Description  json.RawMessage `json:"description"`
	Comments     json.RawMessage `json:"comments"`
}

type decodedRackInput struct {
	site         applicationdcim.Field[shared.ID]
	name         applicationdcim.Field[string]
	facilityID   applicationdcim.Field[string]
	rackType     applicationdcim.Field[shared.ID]
	status       applicationdcim.Field[string]
	role         applicationdcim.Field[shared.ID]
	serial       applicationdcim.Field[string]
	assetTag     applicationdcim.Field[string]
	formFactor   applicationdcim.Field[string]
	width        applicationdcim.Field[uint32]
	uHeight      applicationdcim.Field[uint32]
	startingUnit applicationdcim.Field[uint32]
	descUnits    applicationdcim.Field[bool]
	airflow      applicationdcim.Field[string]
	description  applicationdcim.Field[string]
	comments     applicationdcim.Field[string]
}

func decodeRackInput(c *gin.Context) (decodedRackInput, error) {
	input, err := decodeTypedObject[rackInputJSON](c)
	if err != nil {
		return decodedRackInput{}, err
	}
	site, err := decodeRackTypeIDField("site", input.Site)
	if err != nil {
		return decodedRackInput{}, err
	}
	name, err := decodeSiteStringField("name", input.Name)
	if err != nil {
		return decodedRackInput{}, err
	}
	facilityID, err := decodeSiteStringField("facility_id", input.FacilityID)
	if err != nil {
		return decodedRackInput{}, err
	}
	rackType, err := decodeRackTypeIDField("rack_type", input.RackType)
	if err != nil {
		return decodedRackInput{}, err
	}
	status, err := decodeSiteStringField("status", input.Status)
	if err != nil {
		return decodedRackInput{}, err
	}
	role, err := decodeRackTypeIDField("role", input.Role)
	if err != nil {
		return decodedRackInput{}, err
	}
	serial, err := decodeSiteStringField("serial", input.Serial)
	if err != nil {
		return decodedRackInput{}, err
	}
	assetTag, err := decodeSiteStringField("asset_tag", input.AssetTag)
	if err != nil {
		return decodedRackInput{}, err
	}
	formFactor, err := decodeSiteStringField("form_factor", input.FormFactor)
	if err != nil {
		return decodedRackInput{}, err
	}
	width, err := decodeRackTypeUint32Field("width", input.Width)
	if err != nil {
		return decodedRackInput{}, err
	}
	uHeight, err := decodeRackTypeUint32Field("u_height", input.UHeight)
	if err != nil {
		return decodedRackInput{}, err
	}
	startingUnit, err := decodeRackTypeUint32Field("starting_unit", input.StartingUnit)
	if err != nil {
		return decodedRackInput{}, err
	}
	descUnits, err := decodeRackTypeBoolField("desc_units", input.DescUnits)
	if err != nil {
		return decodedRackInput{}, err
	}
	airflow, err := decodeRackAirflowField(input.Airflow)
	if err != nil {
		return decodedRackInput{}, err
	}
	description, err := decodeSiteStringField("description", input.Description)
	if err != nil {
		return decodedRackInput{}, err
	}
	comments, err := decodeSiteStringField("comments", input.Comments)
	if err != nil {
		return decodedRackInput{}, err
	}
	return decodedRackInput{
		site: site, name: name, facilityID: facilityID, rackType: rackType,
		status: status, role: role, serial: serial, assetTag: assetTag,
		formFactor: formFactor, width: width, uHeight: uHeight, startingUnit: startingUnit,
		descUnits: descUnits, airflow: airflow, description: description, comments: comments,
	}, nil
}

func decodeRackAirflowField(raw json.RawMessage) (applicationdcim.Field[string], error) {
	if len(raw) > 0 && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		// The pinned ChoiceField has allow_blank but not allow_null, so REST
		// coerces an explicit null to the present blank choice.
		return applicationdcim.FieldValue(""), nil
	}
	return decodeSiteStringField("airflow", raw)
}

func (input decodedRackInput) createCommand() applicationdcim.CreateRackCommand {
	return applicationdcim.CreateRackCommand{
		Site: input.site, Name: input.name, FacilityID: input.facilityID,
		RackType: input.rackType, Status: input.status, Role: input.role,
		Serial: input.serial, AssetTag: input.assetTag, FormFactor: input.formFactor,
		Width: input.width, UHeight: input.uHeight, StartingUnit: input.startingUnit,
		DescUnits: input.descUnits, Airflow: input.airflow,
		Description: input.description, Comments: input.comments,
	}
}

func (input decodedRackInput) replaceCommand(id shared.ID) applicationdcim.ReplaceRackCommand {
	return applicationdcim.ReplaceRackCommand{ID: id, CreateRackCommand: input.createCommand()}
}

func (input decodedRackInput) updateCommand(id shared.ID) applicationdcim.UpdateRackCommand {
	return applicationdcim.UpdateRackCommand{
		ID: id, Site: input.site, Name: input.name, FacilityID: input.facilityID,
		RackType: input.rackType, Status: input.status, Role: input.role,
		Serial: input.serial, AssetTag: input.assetTag, FormFactor: input.formFactor,
		Width: input.width, UHeight: input.uHeight, StartingUnit: input.startingUnit,
		DescUnits: input.descUnits, Airflow: input.airflow,
		Description: input.description, Comments: input.comments,
	}
}

func parseRackList(values url.Values) (applicationdcim.ListRacksQuery, error) {
	allowed := map[string]struct{}{
		"limit": {}, "offset": {}, "q": {}, "ordering": {}, "id": {},
		"site_id": {}, "site_slug": {}, "name": {}, "status": {},
		"role_id": {}, "role_slug": {}, "rack_type_id": {}, "rack_type_slug": {},
	}
	for key := range values {
		if _, ok := allowed[key]; !ok {
			return applicationdcim.ListRacksQuery{}, shared.Invalid(key, "Unsupported filter.")
		}
	}
	query := applicationdcim.ListRacksQuery{}
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
	if query.IDs, err = parseRackTypeSignedFilters(values["id"]); err != nil {
		return query, shared.Invalid("id", "A valid integer is required.")
	}
	if query.SiteIDs, err = parseRackTypeSignedFilters(values["site_id"]); err != nil {
		return query, shared.Invalid("site_id", "A valid integer is required.")
	}
	if query.RoleIDs, err = parseRackTypeSignedFilters(values["role_id"]); err != nil {
		return query, shared.Invalid("role_id", "A valid integer is required.")
	}
	if query.RackTypeIDs, err = parseRackTypeSignedFilters(values["rack_type_id"]); err != nil {
		return query, shared.Invalid("rack_type_id", "A valid integer is required.")
	}
	for _, raw := range values["ordering"] {
		query.Ordering = append(query.Ordering, strings.Split(raw, ",")...)
	}
	query.SiteSlugs = append([]string(nil), values["site_slug"]...)
	query.Names = append([]string(nil), values["name"]...)
	query.Statuses = append([]string(nil), values["status"]...)
	query.RoleSlugs = append([]string(nil), values["role_slug"]...)
	query.RackTypeSlugs = append([]string(nil), values["rack_type_slug"]...)
	return query, nil
}
