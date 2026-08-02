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

type DeviceTypeService interface {
	ListDeviceTypes(
		context.Context,
		identity.Principal,
		applicationdcim.ListDeviceTypesQuery,
	) (applicationdcim.DeviceTypePage, error)
	GetDeviceType(
		context.Context,
		identity.Principal,
		applicationdcim.GetDeviceTypeQuery,
	) (*domaindcim.DeviceType, error)
	CreateDeviceType(
		context.Context,
		identity.Principal,
		applicationdcim.CreateDeviceTypeCommand,
	) (*domaindcim.DeviceType, error)
	ReplaceDeviceType(
		context.Context,
		identity.Principal,
		applicationdcim.ReplaceDeviceTypeCommand,
	) (*domaindcim.DeviceType, error)
	UpdateDeviceType(
		context.Context,
		identity.Principal,
		applicationdcim.UpdateDeviceTypeCommand,
	) (*domaindcim.DeviceType, error)
	DeleteDeviceType(
		context.Context,
		identity.Principal,
		applicationdcim.DeleteDeviceTypeCommand,
	) error
}

var _ DeviceTypeService = (*applicationdcim.DeviceTypeService)(nil)

type DeviceTypeRESTHandler struct{ service DeviceTypeService }

func NewDeviceTypeRESTHandler(service DeviceTypeService) *DeviceTypeRESTHandler {
	if service == nil {
		panic("REST DeviceType handler requires a typed DeviceType service")
	}
	return &DeviceTypeRESTHandler{service: service}
}

func (handler *DeviceTypeRESTHandler) Register(
	r gin.IRoutes,
	middlewares ...gin.HandlerFunc,
) {
	const base = "/api/dcim/device-types/"
	r.GET(base, append(middlewares, handler.list())...)
	r.POST(base, append(middlewares, handler.create())...)
	r.GET(base+":id/", append(middlewares, handler.get())...)
	r.PUT(base+":id/", append(middlewares, handler.replace())...)
	r.PATCH(base+":id/", append(middlewares, handler.update())...)
	r.DELETE(base+":id/", append(middlewares, handler.delete())...)
}

type deviceTypeDTO struct {
	ID                     int64                      `json:"id"`
	URL                    string                     `json:"url"`
	Display                string                     `json:"display"`
	Manufacturer           rackTypeObjectReferenceDTO `json:"manufacturer"`
	Model                  string                     `json:"model"`
	Slug                   string                     `json:"slug"`
	PartNumber             string                     `json:"part_number"`
	UHeight                float64                    `json:"u_height"`
	ExcludeFromUtilization bool                       `json:"exclude_from_utilization"`
	IsFullDepth            bool                       `json:"is_full_depth"`
	Airflow                *rackTypeStringChoiceDTO   `json:"airflow"`
	Description            string                     `json:"description"`
	Comments               string                     `json:"comments"`
	Created                time.Time                  `json:"created"`
	LastUpdated            time.Time                  `json:"last_updated"`
	DeviceCount            *uint64                    `json:"device_count,omitempty"`
	InterfaceTemplateCount uint64                     `json:"interface_template_count"`
}

type deviceTypeListDTO struct {
	Count    uint64          `json:"count"`
	Next     *string         `json:"next"`
	Previous *string         `json:"previous"`
	Results  []deviceTypeDTO `json:"results"`
}

func newDeviceTypeDTO(
	deviceType *domaindcim.DeviceType,
	includeDeviceCount bool,
) deviceTypeDTO {
	manufacturer := deviceType.Manufacturer()
	dto := deviceTypeDTO{
		ID: deviceType.ID().Int64(), URL: deviceTypeURL(deviceType.ID()),
		Display: deviceType.Display(),
		Manufacturer: rackTypeObjectReferenceDTO{
			ID: manufacturer.ID().Int64(), URL: manufacturerURL(manufacturer.ID()),
			Display: manufacturer.Display(),
		},
		Model: deviceType.Model(), Slug: deviceType.Slug().String(),
		PartNumber: deviceType.PartNumber(), UHeight: deviceType.UHeight().Float64(),
		ExcludeFromUtilization: deviceType.ExcludeFromUtilization(),
		IsFullDepth:            deviceType.IsFullDepth(),
		Airflow:                deviceTypeAirflowDTO(deviceType.Airflow()),
		Description:            deviceType.Description(), Comments: deviceType.Comments(),
		Created: deviceType.Created().Time, LastUpdated: deviceType.LastUpdated().Time,
		InterfaceTemplateCount: deviceType.InterfaceTemplateCount(),
	}
	if includeDeviceCount {
		count := deviceType.DeviceCount()
		dto.DeviceCount = &count
	}
	return dto
}

func deviceTypeAirflowDTO(
	nullable domaindcim.NullableDeviceAirflow,
) *rackTypeStringChoiceDTO {
	airflow, present := nullable.Get()
	if !present || airflow == "" {
		return nil
	}
	return &rackTypeStringChoiceDTO{
		Value: airflow.String(), Label: airflowLabels[airflow.String()],
	}
}

func deviceTypeURL(id shared.ID) string {
	return "/api/dcim/device-types/" + id.String() + "/"
}

func (handler *DeviceTypeRESTHandler) list() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := Principal(c)
		if !ok {
			writeError(c, shared.Unauthenticated())
			return
		}
		query, err := parseDeviceTypeList(c.Request.URL.Query())
		if err != nil {
			writeError(c, err)
			return
		}
		page, err := handler.service.ListDeviceTypes(
			c.Request.Context(), principal, query,
		)
		if err != nil {
			writeError(c, err)
			return
		}
		results := make([]deviceTypeDTO, 0, len(page.Results))
		for _, deviceType := range page.Results {
			results = append(results, newDeviceTypeDTO(deviceType, true))
		}
		c.JSON(http.StatusOK, deviceTypeListDTO{
			Count: page.Count,
			Next: organizationPageURL(
				c, query.Offset, query.EffectiveLimit(), page.Count, true,
			),
			Previous: organizationPageURL(
				c, query.Offset, query.EffectiveLimit(), page.Count, false,
			),
			Results: results,
		})
	}
}

func (handler *DeviceTypeRESTHandler) get() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		deviceType, err := handler.service.GetDeviceType(
			c.Request.Context(), principal, applicationdcim.GetDeviceTypeQuery{ID: id},
		)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newDeviceTypeDTO(deviceType, true))
	}
}

func (handler *DeviceTypeRESTHandler) create() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := Principal(c)
		if !ok {
			writeError(c, shared.Unauthenticated())
			return
		}
		input, err := decodeDeviceTypeInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		deviceType, err := handler.service.CreateDeviceType(
			c.Request.Context(), principal, input.createCommand(),
		)
		if err != nil {
			writeError(c, err)
			return
		}
		c.Header("Location", deviceTypeURL(deviceType.ID()))
		c.JSON(http.StatusCreated, newDeviceTypeDTO(deviceType, false))
	}
}

func (handler *DeviceTypeRESTHandler) replace() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		input, err := decodeDeviceTypeInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		deviceType, err := handler.service.ReplaceDeviceType(
			c.Request.Context(), principal, input.replaceCommand(id),
		)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newDeviceTypeDTO(deviceType, true))
	}
}

func (handler *DeviceTypeRESTHandler) update() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		input, err := decodeDeviceTypeInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		deviceType, err := handler.service.UpdateDeviceType(
			c.Request.Context(), principal, input.updateCommand(id),
		)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newDeviceTypeDTO(deviceType, true))
	}
}

func (handler *DeviceTypeRESTHandler) delete() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		if err := handler.service.DeleteDeviceType(
			c.Request.Context(), principal,
			applicationdcim.DeleteDeviceTypeCommand{ID: id},
		); err != nil {
			writeError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

type deviceTypeInputJSON struct {
	Manufacturer           json.RawMessage `json:"manufacturer"`
	Model                  json.RawMessage `json:"model"`
	Slug                   json.RawMessage `json:"slug"`
	PartNumber             json.RawMessage `json:"part_number"`
	UHeight                json.RawMessage `json:"u_height"`
	ExcludeFromUtilization json.RawMessage `json:"exclude_from_utilization"`
	IsFullDepth            json.RawMessage `json:"is_full_depth"`
	Airflow                json.RawMessage `json:"airflow"`
	Description            json.RawMessage `json:"description"`
	Comments               json.RawMessage `json:"comments"`
}

type decodedDeviceTypeInput struct {
	manufacturer           applicationdcim.Field[shared.ID]
	model                  applicationdcim.Field[string]
	slug                   applicationdcim.Field[string]
	partNumber             applicationdcim.Field[string]
	uHeight                applicationdcim.Field[string]
	excludeFromUtilization applicationdcim.Field[bool]
	isFullDepth            applicationdcim.Field[bool]
	airflow                applicationdcim.Field[string]
	description            applicationdcim.Field[string]
	comments               applicationdcim.Field[string]
}

func decodeDeviceTypeInput(c *gin.Context) (decodedDeviceTypeInput, error) {
	input, err := decodeTypedObject[deviceTypeInputJSON](c)
	if err != nil {
		return decodedDeviceTypeInput{}, err
	}
	manufacturer, err := decodeRackTypeIDField("manufacturer", input.Manufacturer)
	if err != nil {
		return decodedDeviceTypeInput{}, err
	}
	model, err := decodeSiteStringField("model", input.Model)
	if err != nil {
		return decodedDeviceTypeInput{}, err
	}
	slug, err := decodeSiteStringField("slug", input.Slug)
	if err != nil {
		return decodedDeviceTypeInput{}, err
	}
	partNumber, err := decodeSiteStringField("part_number", input.PartNumber)
	if err != nil {
		return decodedDeviceTypeInput{}, err
	}
	uHeight, err := decodeDeviceTypeHeightField(input.UHeight)
	if err != nil {
		return decodedDeviceTypeInput{}, err
	}
	excludeFromUtilization, err := decodeRackTypeBoolField(
		"exclude_from_utilization", input.ExcludeFromUtilization,
	)
	if err != nil {
		return decodedDeviceTypeInput{}, err
	}
	isFullDepth, err := decodeRackTypeBoolField("is_full_depth", input.IsFullDepth)
	if err != nil {
		return decodedDeviceTypeInput{}, err
	}
	airflow, err := decodeSiteStringField("airflow", input.Airflow)
	if err != nil {
		return decodedDeviceTypeInput{}, err
	}
	description, err := decodeSiteStringField("description", input.Description)
	if err != nil {
		return decodedDeviceTypeInput{}, err
	}
	comments, err := decodeSiteStringField("comments", input.Comments)
	if err != nil {
		return decodedDeviceTypeInput{}, err
	}
	return decodedDeviceTypeInput{
		manufacturer: manufacturer, model: model, slug: slug,
		partNumber: partNumber, uHeight: uHeight,
		excludeFromUtilization: excludeFromUtilization,
		isFullDepth:            isFullDepth, airflow: airflow,
		description: description, comments: comments,
	}, nil
}

func decodeDeviceTypeHeightField(
	raw json.RawMessage,
) (applicationdcim.Field[string], error) {
	if len(raw) == 0 {
		return applicationdcim.OmittedField[string](), nil
	}
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return applicationdcim.NullField[string](), nil
	}
	var stringValue string
	if err := json.Unmarshal(raw, &stringValue); err == nil {
		return applicationdcim.FieldValue(stringValue), nil
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return applicationdcim.Field[string]{},
			shared.Invalid("u_height", "Must be a non-negative number.")
	}
	number, ok := value.(json.Number)
	if !ok {
		return applicationdcim.Field[string]{},
			shared.Invalid("u_height", "Must be a non-negative number.")
	}
	return applicationdcim.FieldValue(number.String()), nil
}

func (input decodedDeviceTypeInput) createCommand() applicationdcim.CreateDeviceTypeCommand {
	return applicationdcim.CreateDeviceTypeCommand{
		Manufacturer: input.manufacturer, Model: input.model, Slug: input.slug,
		PartNumber: input.partNumber, UHeight: input.uHeight,
		ExcludeFromUtilization: input.excludeFromUtilization,
		IsFullDepth:            input.isFullDepth, Airflow: input.airflow,
		Description: input.description, Comments: input.comments,
	}
}

func (input decodedDeviceTypeInput) replaceCommand(
	id shared.ID,
) applicationdcim.ReplaceDeviceTypeCommand {
	return applicationdcim.ReplaceDeviceTypeCommand{
		ID: id, CreateDeviceTypeCommand: input.createCommand(),
	}
}

func (input decodedDeviceTypeInput) updateCommand(
	id shared.ID,
) applicationdcim.UpdateDeviceTypeCommand {
	return applicationdcim.UpdateDeviceTypeCommand{
		ID: id, Manufacturer: input.manufacturer, Model: input.model,
		Slug: input.slug, PartNumber: input.partNumber, UHeight: input.uHeight,
		ExcludeFromUtilization: input.excludeFromUtilization,
		IsFullDepth:            input.isFullDepth, Airflow: input.airflow,
		Description: input.description, Comments: input.comments,
	}
}

func parseDeviceTypeList(
	values url.Values,
) (applicationdcim.ListDeviceTypesQuery, error) {
	allowed := map[string]struct{}{
		"limit": {}, "offset": {}, "q": {}, "ordering": {}, "id": {},
		"manufacturer_id": {}, "manufacturer_slug": {}, "model": {}, "slug": {},
	}
	for key := range values {
		if _, ok := allowed[key]; !ok {
			return applicationdcim.ListDeviceTypesQuery{},
				shared.Invalid(key, "Unsupported filter.")
		}
	}
	query := applicationdcim.ListDeviceTypesQuery{}
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
