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

type DeviceService interface {
	ListDevices(context.Context, identity.Principal, applicationdcim.ListDevicesQuery) (applicationdcim.DevicePage, error)
	GetDevice(context.Context, identity.Principal, applicationdcim.GetDeviceQuery) (*domaindcim.Device, error)
	CreateDevice(context.Context, identity.Principal, applicationdcim.CreateDeviceCommand) (*domaindcim.Device, error)
	ReplaceDevice(context.Context, identity.Principal, applicationdcim.ReplaceDeviceCommand) (*domaindcim.Device, error)
	UpdateDevice(context.Context, identity.Principal, applicationdcim.UpdateDeviceCommand) (*domaindcim.Device, error)
	DeleteDevice(context.Context, identity.Principal, applicationdcim.DeleteDeviceCommand) error
}

var _ DeviceService = (*applicationdcim.DeviceService)(nil)

type DeviceRESTHandler struct{ service DeviceService }

func NewDeviceRESTHandler(service DeviceService) *DeviceRESTHandler {
	if service == nil {
		panic("REST Device handler requires a typed Device service")
	}
	return &DeviceRESTHandler{service: service}
}

func (handler *DeviceRESTHandler) Register(r gin.IRoutes, middlewares ...gin.HandlerFunc) {
	const base = "/api/dcim/devices/"
	r.GET(base, append(middlewares, handler.list())...)
	r.POST(base, append(middlewares, handler.create())...)
	r.GET(base+":id/", append(middlewares, handler.get())...)
	r.PUT(base+":id/", append(middlewares, handler.replace())...)
	r.PATCH(base+":id/", append(middlewares, handler.update())...)
	r.DELETE(base+":id/", append(middlewares, handler.delete())...)
}

type deviceDTO struct {
	ID             int64                       `json:"id"`
	URL            string                      `json:"url"`
	Display        string                      `json:"display"`
	DeviceType     rackTypeObjectReferenceDTO  `json:"device_type"`
	Role           rackTypeObjectReferenceDTO  `json:"role"`
	Name           *string                     `json:"name"`
	Site           rackTypeObjectReferenceDTO  `json:"site"`
	Rack           *rackTypeObjectReferenceDTO `json:"rack"`
	Position       *float64                    `json:"position"`
	Face           *rackTypeStringChoiceDTO    `json:"face"`
	Status         rackTypeStringChoiceDTO     `json:"status"`
	Serial         string                      `json:"serial"`
	AssetTag       *string                     `json:"asset_tag"`
	Airflow        *rackTypeStringChoiceDTO    `json:"airflow"`
	Description    string                      `json:"description"`
	Comments       string                      `json:"comments"`
	Created        time.Time                   `json:"created"`
	LastUpdated    time.Time                   `json:"last_updated"`
	InterfaceCount uint64                      `json:"interface_count"`
}

type deviceListDTO struct {
	Count    uint64      `json:"count"`
	Next     *string     `json:"next"`
	Previous *string     `json:"previous"`
	Results  []deviceDTO `json:"results"`
}

func newDeviceDTO(device *domaindcim.Device) deviceDTO {
	deviceType, role, site := device.DeviceType(), device.Role(), device.Site()
	dto := deviceDTO{
		ID: device.ID().Int64(), URL: deviceURL(device.ID()), Display: device.Display(),
		DeviceType: rackTypeObjectReferenceDTO{
			ID: deviceType.ID().Int64(), URL: deviceTypeURL(deviceType.ID()),
			Display: deviceType.Display(),
		},
		Role: rackTypeObjectReferenceDTO{
			ID: role.ID.Int64(), URL: deviceRoleURL(role.ID), Display: role.Display,
		},
		Name: nullableDeviceDTOString(device.Name()),
		Site: rackTypeObjectReferenceDTO{
			ID: site.ID().Int64(), URL: siteURL(site.ID()), Display: site.Display(),
		},
		Status: rackTypeStringChoiceDTO{
			Value: device.Status().String(), Label: commonChoiceLabels[device.Status().String()],
		},
		Serial: device.Serial(), AssetTag: nullableDeviceDTOString(device.AssetTag()),
		Description: device.Description(), Comments: device.Comments(),
		Created: device.Created().Time, LastUpdated: device.LastUpdated().Time,
		InterfaceCount: device.InterfaceCount(),
	}
	if rack, present := device.Rack().Get(); present {
		dto.Rack = &rackTypeObjectReferenceDTO{
			ID: rack.ID().Int64(), URL: rackURL(rack.ID()), Display: rack.Display(),
		}
	}
	if position, present := device.Position().Get(); present {
		value := float64(position.HalfUnits()) / 2
		dto.Position = &value
	}
	if face := device.Face().String(); face != "" {
		dto.Face = &rackTypeStringChoiceDTO{
			Value: face, Label: map[string]string{"front": "Front", "rear": "Rear"}[face],
		}
	}
	if airflow, present := device.Airflow().Get(); present && airflow.String() != "" {
		dto.Airflow = &rackTypeStringChoiceDTO{
			Value: airflow.String(), Label: airflowLabels[airflow.String()],
		}
	}
	return dto
}

func nullableDeviceDTOString(value domaindcim.DeviceNullable[string]) *string {
	text, present := value.Get()
	if !present {
		return nil
	}
	return &text
}

func (handler *DeviceRESTHandler) list() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := Principal(c)
		if !ok {
			writeError(c, shared.Unauthenticated())
			return
		}
		query, err := parseDeviceList(c.Request.URL.Query())
		if err != nil {
			writeError(c, err)
			return
		}
		page, err := handler.service.ListDevices(c.Request.Context(), principal, query)
		if err != nil {
			writeError(c, err)
			return
		}
		results := make([]deviceDTO, 0, len(page.Results))
		for _, device := range page.Results {
			results = append(results, newDeviceDTO(device))
		}
		c.JSON(http.StatusOK, deviceListDTO{
			Count: page.Count,
			Next:  organizationPageURL(c, query.Offset, query.EffectiveLimit(), page.Count, true),
			Previous: organizationPageURL(
				c, query.Offset, query.EffectiveLimit(), page.Count, false,
			),
			Results: results,
		})
	}
}

func (handler *DeviceRESTHandler) get() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		device, err := handler.service.GetDevice(
			c.Request.Context(), principal, applicationdcim.GetDeviceQuery{ID: id},
		)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newDeviceDTO(device))
	}
}

func (handler *DeviceRESTHandler) create() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := Principal(c)
		if !ok {
			writeError(c, shared.Unauthenticated())
			return
		}
		input, err := decodeDeviceInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		device, err := handler.service.CreateDevice(
			c.Request.Context(), principal, input.createCommand(),
		)
		if err != nil {
			writeError(c, err)
			return
		}
		c.Header("Location", deviceURL(device.ID()))
		c.JSON(http.StatusCreated, newDeviceDTO(device))
	}
}

func (handler *DeviceRESTHandler) replace() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		input, err := decodeDeviceInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		device, err := handler.service.ReplaceDevice(
			c.Request.Context(), principal, input.replaceCommand(id),
		)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newDeviceDTO(device))
	}
}

func (handler *DeviceRESTHandler) update() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		input, err := decodeDeviceInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		device, err := handler.service.UpdateDevice(
			c.Request.Context(), principal, input.updateCommand(id),
		)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newDeviceDTO(device))
	}
}

func (handler *DeviceRESTHandler) delete() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		if err := handler.service.DeleteDevice(
			c.Request.Context(), principal, applicationdcim.DeleteDeviceCommand{ID: id},
		); err != nil {
			writeError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

type deviceInputJSON struct {
	DeviceType  json.RawMessage `json:"device_type"`
	Role        json.RawMessage `json:"role"`
	Name        json.RawMessage `json:"name"`
	Site        json.RawMessage `json:"site"`
	Rack        json.RawMessage `json:"rack"`
	Position    json.RawMessage `json:"position"`
	Face        json.RawMessage `json:"face"`
	Status      json.RawMessage `json:"status"`
	Serial      json.RawMessage `json:"serial"`
	AssetTag    json.RawMessage `json:"asset_tag"`
	Airflow     json.RawMessage `json:"airflow"`
	Description json.RawMessage `json:"description"`
	Comments    json.RawMessage `json:"comments"`
}

type decodedDeviceInput struct {
	deviceType  applicationdcim.Field[shared.ID]
	role        applicationdcim.Field[shared.ID]
	name        applicationdcim.Field[string]
	site        applicationdcim.Field[shared.ID]
	rack        applicationdcim.Field[shared.ID]
	position    applicationdcim.Field[string]
	face        applicationdcim.Field[string]
	status      applicationdcim.Field[string]
	serial      applicationdcim.Field[string]
	assetTag    applicationdcim.Field[string]
	airflow     applicationdcim.Field[string]
	description applicationdcim.Field[string]
	comments    applicationdcim.Field[string]
}

func decodeDeviceInput(c *gin.Context) (decodedDeviceInput, error) {
	input, err := decodeTypedObject[deviceInputJSON](c)
	if err != nil {
		return decodedDeviceInput{}, err
	}
	deviceType, err := decodeRackTypeIDField("device_type", input.DeviceType)
	if err != nil {
		return decodedDeviceInput{}, err
	}
	role, err := decodeRackTypeIDField("role", input.Role)
	if err != nil {
		return decodedDeviceInput{}, err
	}
	name, err := decodeSiteStringField("name", input.Name)
	if err != nil {
		return decodedDeviceInput{}, err
	}
	site, err := decodeRackTypeIDField("site", input.Site)
	if err != nil {
		return decodedDeviceInput{}, err
	}
	rack, err := decodeRackTypeIDField("rack", input.Rack)
	if err != nil {
		return decodedDeviceInput{}, err
	}
	position, err := decodeDevicePositionField(input.Position)
	if err != nil {
		return decodedDeviceInput{}, err
	}
	face, err := decodeDeviceBlankChoiceField("face", input.Face)
	if err != nil {
		return decodedDeviceInput{}, err
	}
	status, err := decodeSiteStringField("status", input.Status)
	if err != nil {
		return decodedDeviceInput{}, err
	}
	serial, err := decodeSiteStringField("serial", input.Serial)
	if err != nil {
		return decodedDeviceInput{}, err
	}
	assetTag, err := decodeSiteStringField("asset_tag", input.AssetTag)
	if err != nil {
		return decodedDeviceInput{}, err
	}
	airflow, err := decodeDeviceBlankChoiceField("airflow", input.Airflow)
	if err != nil {
		return decodedDeviceInput{}, err
	}
	description, err := decodeSiteStringField("description", input.Description)
	if err != nil {
		return decodedDeviceInput{}, err
	}
	comments, err := decodeSiteStringField("comments", input.Comments)
	if err != nil {
		return decodedDeviceInput{}, err
	}
	return decodedDeviceInput{
		deviceType: deviceType, role: role, name: name, site: site, rack: rack,
		position: position, face: face, status: status, serial: serial,
		assetTag: assetTag, airflow: airflow, description: description, comments: comments,
	}, nil
}

func decodeDeviceBlankChoiceField(
	name string,
	raw json.RawMessage,
) (applicationdcim.Field[string], error) {
	if len(raw) > 0 && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return applicationdcim.FieldValue(""), nil
	}
	return decodeSiteStringField(name, raw)
}

func decodeDevicePositionField(
	raw json.RawMessage,
) (applicationdcim.Field[string], error) {
	if len(raw) == 0 {
		return applicationdcim.OmittedField[string](), nil
	}
	trimmed := bytes.TrimSpace(raw)
	if bytes.Equal(trimmed, []byte("null")) {
		return applicationdcim.NullField[string](), nil
	}
	var text string
	if len(trimmed) > 0 && trimmed[0] == '"' {
		if err := json.Unmarshal(trimmed, &text); err != nil {
			return applicationdcim.Field[string]{}, shared.Invalid(
				"position", "A valid number is required.")

		}
	} else {
		var number json.Number
		if err := json.Unmarshal(trimmed, &number); err != nil {
			return applicationdcim.Field[string]{}, shared.Invalid(
				"position", "A valid number is required.")

		}
		text = number.String()
	}
	if _, err := strconv.ParseFloat(strings.TrimSpace(text), 64); err != nil {
		return applicationdcim.Field[string]{}, shared.Invalid(
			"position", "A valid number is required.")

	}
	return applicationdcim.FieldValue(text), nil
}

func (input decodedDeviceInput) createCommand() applicationdcim.CreateDeviceCommand {
	return applicationdcim.CreateDeviceCommand{
		DeviceType: input.deviceType, Role: input.role, Name: input.name,
		Site: input.site, Rack: input.rack, Position: input.position,
		Face: input.face, Status: input.status, Serial: input.serial,
		AssetTag: input.assetTag, Airflow: input.airflow,
		Description: input.description, Comments: input.comments,
	}
}

func (input decodedDeviceInput) replaceCommand(id shared.ID) applicationdcim.ReplaceDeviceCommand {
	return applicationdcim.ReplaceDeviceCommand{ID: id, CreateDeviceCommand: input.createCommand()}
}

func (input decodedDeviceInput) updateCommand(id shared.ID) applicationdcim.UpdateDeviceCommand {
	return applicationdcim.UpdateDeviceCommand{
		ID: id, DeviceType: input.deviceType, Role: input.role, Name: input.name,
		Site: input.site, Rack: input.rack, Position: input.position,
		Face: input.face, Status: input.status, Serial: input.serial,
		AssetTag: input.assetTag, Airflow: input.airflow,
		Description: input.description, Comments: input.comments,
	}
}

func parseDeviceList(values url.Values) (applicationdcim.ListDevicesQuery, error) {
	allowed := map[string]struct{}{
		"limit": {}, "offset": {}, "q": {}, "ordering": {}, "id": {},
		"site_id": {}, "site_slug": {}, "rack_id": {},
		"device_type_id": {}, "device_type_slug": {},
		"role_id": {}, "role_slug": {}, "name": {}, "status": {},
	}
	for key := range values {
		if _, ok := allowed[key]; !ok {
			return applicationdcim.ListDevicesQuery{}, shared.Invalid(key, "Unsupported filter.")
		}
	}
	query := applicationdcim.ListDevicesQuery{}
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
	if query.RackIDs, err = parseRackTypeSignedFilters(values["rack_id"]); err != nil {
		return query, shared.Invalid("rack_id", "A valid integer is required.")
	}
	if query.DeviceTypeIDs, err = parseRackTypeSignedFilters(values["device_type_id"]); err != nil {
		return query, shared.Invalid("device_type_id", "A valid integer is required.")
	}
	if query.RoleIDs, err = parseRackTypeSignedFilters(values["role_id"]); err != nil {
		return query, shared.Invalid("role_id", "A valid integer is required.")
	}
	for _, raw := range values["ordering"] {
		query.Ordering = append(query.Ordering, strings.Split(raw, ",")...)
	}
	query.SiteSlugs = append([]string(nil), values["site_slug"]...)
	query.DeviceTypeSlugs = append([]string(nil), values["device_type_slug"]...)
	query.RoleSlugs = append([]string(nil), values["role_slug"]...)
	query.Names = append([]string(nil), values["name"]...)
	query.Statuses = append([]string(nil), values["status"]...)
	return query, nil
}
