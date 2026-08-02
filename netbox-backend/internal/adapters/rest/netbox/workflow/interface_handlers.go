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

type InterfaceService interface {
	ListInterfaces(
		context.Context,
		identity.Principal,
		applicationdcim.ListInterfacesQuery,
	) (applicationdcim.InterfacePage, error)
	GetInterface(
		context.Context,
		identity.Principal,
		applicationdcim.GetInterfaceQuery,
	) (*domaindcim.Interface, error)
	CreateInterface(
		context.Context,
		identity.Principal,
		applicationdcim.CreateInterfaceCommand,
	) (*domaindcim.Interface, error)
	ReplaceInterface(
		context.Context,
		identity.Principal,
		applicationdcim.ReplaceInterfaceCommand,
	) (*domaindcim.Interface, error)
	UpdateInterface(
		context.Context,
		identity.Principal,
		applicationdcim.UpdateInterfaceCommand,
	) (*domaindcim.Interface, error)
	DeleteInterface(
		context.Context,
		identity.Principal,
		applicationdcim.DeleteInterfaceCommand,
	) error
}

var _ InterfaceService = (*applicationdcim.InterfaceService)(nil)

type InterfaceRESTHandler struct {
	service InterfaceService
}

func NewInterfaceRESTHandler(service InterfaceService) *InterfaceRESTHandler {
	if service == nil {
		panic("REST Interface handler requires a typed Interface service")
	}
	return &InterfaceRESTHandler{service: service}
}

func (handler *InterfaceRESTHandler) Register(
	r gin.IRoutes,
	middlewares ...gin.HandlerFunc,
) {
	const base = "/api/dcim/interfaces/"
	r.GET(base, append(middlewares, handler.list())...)
	r.POST(base, append(middlewares, handler.create())...)
	r.GET(base+":id/", append(middlewares, handler.get())...)
	r.PUT(base+":id/", append(middlewares, handler.replace())...)
	r.PATCH(base+":id/", append(middlewares, handler.update())...)
	r.DELETE(base+":id/", append(middlewares, handler.delete())...)
}

type interfaceDTO struct {
	ID               int64                      `json:"id"`
	URL              string                     `json:"url"`
	Display          string                     `json:"display"`
	Device           rackTypeObjectReferenceDTO `json:"device"`
	Name             string                     `json:"name"`
	Label            string                     `json:"label"`
	Type             rackTypeStringChoiceDTO    `json:"type"`
	Enabled          bool                       `json:"enabled"`
	MgmtOnly         bool                       `json:"mgmt_only"`
	MTU              *uint32                    `json:"mtu"`
	Speed            *uint64                    `json:"speed"`
	Duplex           *rackTypeStringChoiceDTO   `json:"duplex"`
	Description      string                     `json:"description"`
	Created          time.Time                  `json:"created"`
	LastUpdated      time.Time                  `json:"last_updated"`
	CountIPAddresses uint64                     `json:"count_ipaddresses"`
}

type interfaceListDTO struct {
	Count    uint64         `json:"count"`
	Next     *string        `json:"next"`
	Previous *string        `json:"previous"`
	Results  []interfaceDTO `json:"results"`
}

func newInterfaceDTO(networkInterface *domaindcim.Interface) interfaceDTO {
	device := networkInterface.Device()
	interfaceType := networkInterface.Type().String()
	var mtu *uint32
	if value, present := networkInterface.MTU().Get(); present {
		mtu = &value
	}
	var speed *uint64
	if value, present := networkInterface.Speed().Get(); present {
		speed = &value
	}
	var duplex *rackTypeStringChoiceDTO
	if value, present := networkInterface.Duplex().Get(); present && value.String() != "" {
		duplex = &rackTypeStringChoiceDTO{
			Value: value.String(),
			Label: map[string]string{
				"half": "Half", "full": "Full", "auto": "Auto",
			}[value.String()],
		}
	}
	return interfaceDTO{
		ID: networkInterface.ID().Int64(), URL: interfaceURL(networkInterface.ID()),
		Display: networkInterface.Display(),
		Device: rackTypeObjectReferenceDTO{
			ID: device.ID().Int64(), URL: deviceURL(device.ID()), Display: device.Display(),
		},
		Name: networkInterface.Name(), Label: networkInterface.Label(),
		Type: rackTypeStringChoiceDTO{
			Value: interfaceType, Label: interfaceTypeResponseLabels[interfaceType],
		},
		Enabled: networkInterface.Enabled(), MgmtOnly: networkInterface.MgmtOnly(),
		MTU: mtu, Speed: speed, Duplex: duplex,
		Description: networkInterface.Description(),
		Created:     networkInterface.Created().Time, LastUpdated: networkInterface.LastUpdated().Time,
		CountIPAddresses: networkInterface.IPAddressCount(),
	}
}

func interfaceURL(id shared.ID) string {
	return "/api/dcim/interfaces/" + id.String() + "/"
}

func deviceURL(id shared.ID) string {
	return "/api/dcim/devices/" + id.String() + "/"
}

func (handler *InterfaceRESTHandler) list() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := Principal(c)
		if !ok {
			writeError(c, shared.Unauthenticated())
			return
		}
		query, err := parseInterfaceList(c.Request.URL.Query())
		if err != nil {
			writeError(c, err)
			return
		}
		page, err := handler.service.ListInterfaces(c.Request.Context(), principal, query)
		if err != nil {
			writeError(c, err)
			return
		}
		results := make([]interfaceDTO, 0, len(page.Results))
		for _, networkInterface := range page.Results {
			results = append(results, newInterfaceDTO(networkInterface))
		}
		c.JSON(http.StatusOK, interfaceListDTO{
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

func (handler *InterfaceRESTHandler) get() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		networkInterface, err := handler.service.GetInterface(
			c.Request.Context(), principal, applicationdcim.GetInterfaceQuery{ID: id},
		)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newInterfaceDTO(networkInterface))
	}
}

func (handler *InterfaceRESTHandler) create() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := Principal(c)
		if !ok {
			writeError(c, shared.Unauthenticated())
			return
		}
		input, err := decodeInterfaceInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		networkInterface, err := handler.service.CreateInterface(
			c.Request.Context(), principal, input.createCommand(),
		)
		if err != nil {
			writeError(c, err)
			return
		}
		c.Header("Location", interfaceURL(networkInterface.ID()))
		c.JSON(http.StatusCreated, newInterfaceDTO(networkInterface))
	}
}

func (handler *InterfaceRESTHandler) replace() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		input, err := decodeInterfaceInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		networkInterface, err := handler.service.ReplaceInterface(
			c.Request.Context(), principal, input.replaceCommand(id),
		)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newInterfaceDTO(networkInterface))
	}
}

func (handler *InterfaceRESTHandler) update() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		input, err := decodeInterfaceInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		networkInterface, err := handler.service.UpdateInterface(
			c.Request.Context(), principal, input.updateCommand(id),
		)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newInterfaceDTO(networkInterface))
	}
}

func (handler *InterfaceRESTHandler) delete() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		if err := handler.service.DeleteInterface(
			c.Request.Context(), principal,
			applicationdcim.DeleteInterfaceCommand{ID: id},
		); err != nil {
			writeError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

type interfaceInputJSON struct {
	Device      json.RawMessage `json:"device"`
	Name        json.RawMessage `json:"name"`
	Label       json.RawMessage `json:"label"`
	Type        json.RawMessage `json:"type"`
	Enabled     json.RawMessage `json:"enabled"`
	MgmtOnly    json.RawMessage `json:"mgmt_only"`
	MTU         json.RawMessage `json:"mtu"`
	Speed       json.RawMessage `json:"speed"`
	Duplex      json.RawMessage `json:"duplex"`
	Description json.RawMessage `json:"description"`
}

type decodedInterfaceInput struct {
	device        applicationdcim.Field[shared.ID]
	name          applicationdcim.Field[string]
	label         applicationdcim.Field[string]
	interfaceType applicationdcim.Field[string]
	enabled       applicationdcim.Field[bool]
	mgmtOnly      applicationdcim.Field[bool]
	mtu           applicationdcim.Field[uint32]
	speed         applicationdcim.Field[uint64]
	duplex        applicationdcim.Field[string]
	description   applicationdcim.Field[string]
}

func decodeInterfaceInput(c *gin.Context) (decodedInterfaceInput, error) {
	input, err := decodeTypedObject[interfaceInputJSON](c)
	if err != nil {
		return decodedInterfaceInput{}, err
	}
	device, err := decodeRackTypeIDField("device", input.Device)
	if err != nil {
		return decodedInterfaceInput{}, err
	}
	name, err := decodeSiteStringField("name", input.Name)
	if err != nil {
		return decodedInterfaceInput{}, err
	}
	label, err := decodeSiteStringField("label", input.Label)
	if err != nil {
		return decodedInterfaceInput{}, err
	}
	interfaceType, err := decodeSiteStringField("type", input.Type)
	if err != nil {
		return decodedInterfaceInput{}, err
	}
	enabled, err := decodeRackTypeBoolField("enabled", input.Enabled)
	if err != nil {
		return decodedInterfaceInput{}, err
	}
	mgmtOnly, err := decodeRackTypeBoolField("mgmt_only", input.MgmtOnly)
	if err != nil {
		return decodedInterfaceInput{}, err
	}
	mtu, err := decodeRackTypeUint32Field("mtu", input.MTU)
	if err != nil {
		return decodedInterfaceInput{}, err
	}
	speed, err := decodeInterfaceUint64Field("speed", input.Speed)
	if err != nil {
		return decodedInterfaceInput{}, err
	}
	duplex, err := decodeSiteStringField("duplex", input.Duplex)
	if err != nil {
		return decodedInterfaceInput{}, err
	}
	description, err := decodeSiteStringField("description", input.Description)
	if err != nil {
		return decodedInterfaceInput{}, err
	}
	return decodedInterfaceInput{
		device: device, name: name, label: label, interfaceType: interfaceType,
		enabled: enabled, mgmtOnly: mgmtOnly, mtu: mtu, speed: speed,
		duplex: duplex, description: description,
	}, nil
}

func decodeInterfaceUint64Field(
	name string,
	raw json.RawMessage,
) (applicationdcim.Field[uint64], error) {
	if len(raw) == 0 {
		return applicationdcim.OmittedField[uint64](), nil
	}
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return applicationdcim.NullField[uint64](), nil
	}
	value, valid := exactJSONInteger(raw)
	if !valid || value.Sign() < 0 || value.BitLen() > 64 {
		return applicationdcim.Field[uint64]{},
			shared.Invalid(name, "A valid integer is required.")
	}
	return applicationdcim.FieldValue(value.Uint64()), nil
}

func (input decodedInterfaceInput) createCommand() applicationdcim.CreateInterfaceCommand {
	return applicationdcim.CreateInterfaceCommand{
		Device: input.device, Name: input.name, Label: input.label,
		Type: input.interfaceType, Enabled: input.enabled, MgmtOnly: input.mgmtOnly,
		MTU: input.mtu, Speed: input.speed, Duplex: input.duplex,
		Description: input.description,
	}
}

func (input decodedInterfaceInput) replaceCommand(
	id shared.ID,
) applicationdcim.ReplaceInterfaceCommand {
	return applicationdcim.ReplaceInterfaceCommand{
		ID: id, CreateInterfaceCommand: input.createCommand(),
	}
}

func (input decodedInterfaceInput) updateCommand(
	id shared.ID,
) applicationdcim.UpdateInterfaceCommand {
	return applicationdcim.UpdateInterfaceCommand{
		ID: id, Device: input.device, Name: input.name, Label: input.label,
		Type: input.interfaceType, Enabled: input.enabled, MgmtOnly: input.mgmtOnly,
		MTU: input.mtu, Speed: input.speed, Duplex: input.duplex,
		Description: input.description,
	}
}

func parseInterfaceList(values url.Values) (applicationdcim.ListInterfacesQuery, error) {
	allowed := map[string]struct{}{
		"limit": {}, "offset": {}, "q": {}, "ordering": {}, "id": {},
		"device_id": {}, "device_name": {}, "name": {}, "type": {},
		"enabled": {}, "mgmt_only": {},
	}
	for key := range values {
		if _, ok := allowed[key]; !ok {
			return applicationdcim.ListInterfacesQuery{},
				shared.Invalid(key, "Unsupported filter.")
		}
	}
	query := applicationdcim.ListInterfacesQuery{}
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
	deviceIDs, err := parseRackTypeSignedFilters(values["device_id"])
	if err != nil {
		return query, shared.Invalid("device_id", "A valid integer is required.")
	}
	query.DeviceIDs = deviceIDs
	for _, raw := range values["ordering"] {
		query.Ordering = append(query.Ordering, strings.Split(raw, ",")...)
	}
	query.DeviceNames = append([]string(nil), values["device_name"]...)
	query.Names = append([]string(nil), values["name"]...)
	query.Types = append([]string(nil), values["type"]...)
	query.Enabled, err = parseInterfaceTemplateBoolFilter(values, "enabled")
	if err != nil {
		return query, err
	}
	query.MgmtOnly, err = parseInterfaceTemplateBoolFilter(values, "mgmt_only")
	if err != nil {
		return query, err
	}
	return query, nil
}
