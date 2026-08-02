package workflow

import (
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

type InterfaceTemplateService interface {
	ListInterfaceTemplates(
		context.Context,
		identity.Principal,
		applicationdcim.ListInterfaceTemplatesQuery,
	) (applicationdcim.InterfaceTemplatePage, error)
	GetInterfaceTemplate(
		context.Context,
		identity.Principal,
		applicationdcim.GetInterfaceTemplateQuery,
	) (*domaindcim.InterfaceTemplate, error)
	CreateInterfaceTemplate(
		context.Context,
		identity.Principal,
		applicationdcim.CreateInterfaceTemplateCommand,
	) (*domaindcim.InterfaceTemplate, error)
	ReplaceInterfaceTemplate(
		context.Context,
		identity.Principal,
		applicationdcim.ReplaceInterfaceTemplateCommand,
	) (*domaindcim.InterfaceTemplate, error)
	UpdateInterfaceTemplate(
		context.Context,
		identity.Principal,
		applicationdcim.UpdateInterfaceTemplateCommand,
	) (*domaindcim.InterfaceTemplate, error)
	DeleteInterfaceTemplate(
		context.Context,
		identity.Principal,
		applicationdcim.DeleteInterfaceTemplateCommand,
	) error
}

var _ InterfaceTemplateService = (*applicationdcim.InterfaceTemplateService)(nil)

type InterfaceTemplateRESTHandler struct {
	service InterfaceTemplateService
}

func NewInterfaceTemplateRESTHandler(
	service InterfaceTemplateService,
) *InterfaceTemplateRESTHandler {
	if service == nil {
		panic("REST InterfaceTemplate handler requires a typed InterfaceTemplate service")
	}
	return &InterfaceTemplateRESTHandler{service: service}
}

func (handler *InterfaceTemplateRESTHandler) Register(
	r gin.IRoutes,
	middlewares ...gin.HandlerFunc,
) {
	const base = "/api/dcim/interface-templates/"
	r.GET(base, append(middlewares, handler.list())...)
	r.POST(base, append(middlewares, handler.create())...)
	r.GET(base+":id/", append(middlewares, handler.get())...)
	r.PUT(base+":id/", append(middlewares, handler.replace())...)
	r.PATCH(base+":id/", append(middlewares, handler.update())...)
	r.DELETE(base+":id/", append(middlewares, handler.delete())...)
}

type interfaceTemplateDTO struct {
	ID          int64                      `json:"id"`
	URL         string                     `json:"url"`
	Display     string                     `json:"display"`
	DeviceType  rackTypeObjectReferenceDTO `json:"device_type"`
	Name        string                     `json:"name"`
	Label       string                     `json:"label"`
	Type        rackTypeStringChoiceDTO    `json:"type"`
	Enabled     bool                       `json:"enabled"`
	MgmtOnly    bool                       `json:"mgmt_only"`
	Description string                     `json:"description"`
	Created     time.Time                  `json:"created"`
	LastUpdated time.Time                  `json:"last_updated"`
}

type interfaceTemplateListDTO struct {
	Count    uint64                 `json:"count"`
	Next     *string                `json:"next"`
	Previous *string                `json:"previous"`
	Results  []interfaceTemplateDTO `json:"results"`
}

func newInterfaceTemplateDTO(
	template *domaindcim.InterfaceTemplate,
) interfaceTemplateDTO {
	deviceType := template.DeviceType()
	interfaceType := template.Type().String()
	return interfaceTemplateDTO{
		ID: template.ID().Int64(), URL: interfaceTemplateURL(template.ID()),
		Display: template.Display(),
		DeviceType: rackTypeObjectReferenceDTO{
			ID: deviceType.ID().Int64(), URL: deviceTypeURL(deviceType.ID()),
			Display: deviceType.Display(),
		},
		Name: template.Name(), Label: template.Label(),
		Type: rackTypeStringChoiceDTO{
			Value: interfaceType, Label: interfaceTypeResponseLabels[interfaceType],
		},
		Enabled: template.Enabled(), MgmtOnly: template.MgmtOnly(),
		Description: template.Description(), Created: template.Created().Time,
		LastUpdated: template.LastUpdated().Time,
	}
}

func interfaceTemplateURL(id shared.ID) string {
	return "/api/dcim/interface-templates/" + id.String() + "/"
}

func (handler *InterfaceTemplateRESTHandler) list() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := Principal(c)
		if !ok {
			writeError(c, shared.Unauthenticated())
			return
		}
		query, err := parseInterfaceTemplateList(c.Request.URL.Query())
		if err != nil {
			writeError(c, err)
			return
		}
		page, err := handler.service.ListInterfaceTemplates(
			c.Request.Context(), principal, query,
		)
		if err != nil {
			writeError(c, err)
			return
		}
		results := make([]interfaceTemplateDTO, 0, len(page.Results))
		for _, template := range page.Results {
			results = append(results, newInterfaceTemplateDTO(template))
		}
		c.JSON(http.StatusOK, interfaceTemplateListDTO{
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

func (handler *InterfaceTemplateRESTHandler) get() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		template, err := handler.service.GetInterfaceTemplate(
			c.Request.Context(), principal,
			applicationdcim.GetInterfaceTemplateQuery{ID: id},
		)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newInterfaceTemplateDTO(template))
	}
}

func (handler *InterfaceTemplateRESTHandler) create() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := Principal(c)
		if !ok {
			writeError(c, shared.Unauthenticated())
			return
		}
		input, err := decodeInterfaceTemplateInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		template, err := handler.service.CreateInterfaceTemplate(
			c.Request.Context(), principal, input.createCommand(),
		)
		if err != nil {
			writeError(c, err)
			return
		}
		c.Header("Location", interfaceTemplateURL(template.ID()))
		c.JSON(http.StatusCreated, newInterfaceTemplateDTO(template))
	}
}

func (handler *InterfaceTemplateRESTHandler) replace() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		input, err := decodeInterfaceTemplateInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		template, err := handler.service.ReplaceInterfaceTemplate(
			c.Request.Context(), principal, input.replaceCommand(id),
		)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newInterfaceTemplateDTO(template))
	}
}

func (handler *InterfaceTemplateRESTHandler) update() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		input, err := decodeInterfaceTemplateInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		template, err := handler.service.UpdateInterfaceTemplate(
			c.Request.Context(), principal, input.updateCommand(id),
		)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newInterfaceTemplateDTO(template))
	}
}

func (handler *InterfaceTemplateRESTHandler) delete() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		if err := handler.service.DeleteInterfaceTemplate(
			c.Request.Context(), principal,
			applicationdcim.DeleteInterfaceTemplateCommand{ID: id},
		); err != nil {
			writeError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

type interfaceTemplateInputJSON struct {
	DeviceType  json.RawMessage `json:"device_type"`
	Name        json.RawMessage `json:"name"`
	Label       json.RawMessage `json:"label"`
	Type        json.RawMessage `json:"type"`
	Enabled     json.RawMessage `json:"enabled"`
	MgmtOnly    json.RawMessage `json:"mgmt_only"`
	Description json.RawMessage `json:"description"`
}

type decodedInterfaceTemplateInput struct {
	deviceType    applicationdcim.Field[shared.ID]
	name          applicationdcim.Field[string]
	label         applicationdcim.Field[string]
	interfaceType applicationdcim.Field[string]
	enabled       applicationdcim.Field[bool]
	mgmtOnly      applicationdcim.Field[bool]
	description   applicationdcim.Field[string]
}

func decodeInterfaceTemplateInput(
	c *gin.Context,
) (decodedInterfaceTemplateInput, error) {
	input, err := decodeTypedObject[interfaceTemplateInputJSON](c)
	if err != nil {
		return decodedInterfaceTemplateInput{}, err
	}
	deviceType, err := decodeRackTypeIDField("device_type", input.DeviceType)
	if err != nil {
		return decodedInterfaceTemplateInput{}, err
	}
	name, err := decodeSiteStringField("name", input.Name)
	if err != nil {
		return decodedInterfaceTemplateInput{}, err
	}
	label, err := decodeSiteStringField("label", input.Label)
	if err != nil {
		return decodedInterfaceTemplateInput{}, err
	}
	interfaceType, err := decodeSiteStringField("type", input.Type)
	if err != nil {
		return decodedInterfaceTemplateInput{}, err
	}
	enabled, err := decodeRackTypeBoolField("enabled", input.Enabled)
	if err != nil {
		return decodedInterfaceTemplateInput{}, err
	}
	mgmtOnly, err := decodeRackTypeBoolField("mgmt_only", input.MgmtOnly)
	if err != nil {
		return decodedInterfaceTemplateInput{}, err
	}
	description, err := decodeSiteStringField("description", input.Description)
	if err != nil {
		return decodedInterfaceTemplateInput{}, err
	}
	return decodedInterfaceTemplateInput{
		deviceType: deviceType, name: name, label: label, interfaceType: interfaceType,
		enabled: enabled, mgmtOnly: mgmtOnly, description: description,
	}, nil
}

func (input decodedInterfaceTemplateInput) createCommand() applicationdcim.CreateInterfaceTemplateCommand {
	return applicationdcim.CreateInterfaceTemplateCommand{
		DeviceType: input.deviceType, Name: input.name, Label: input.label,
		Type: input.interfaceType, Enabled: input.enabled, MgmtOnly: input.mgmtOnly,
		Description: input.description,
	}
}

func (input decodedInterfaceTemplateInput) replaceCommand(
	id shared.ID,
) applicationdcim.ReplaceInterfaceTemplateCommand {
	return applicationdcim.ReplaceInterfaceTemplateCommand{
		ID: id, CreateInterfaceTemplateCommand: input.createCommand(),
	}
}

func (input decodedInterfaceTemplateInput) updateCommand(
	id shared.ID,
) applicationdcim.UpdateInterfaceTemplateCommand {
	return applicationdcim.UpdateInterfaceTemplateCommand{
		ID: id, DeviceType: input.deviceType, Name: input.name, Label: input.label,
		Type: input.interfaceType, Enabled: input.enabled, MgmtOnly: input.mgmtOnly,
		Description: input.description,
	}
}

func parseInterfaceTemplateList(
	values url.Values,
) (applicationdcim.ListInterfaceTemplatesQuery, error) {
	allowed := map[string]struct{}{
		"limit": {}, "offset": {}, "q": {}, "ordering": {}, "id": {},
		"device_type_id": {}, "name": {}, "type": {}, "enabled": {}, "mgmt_only": {},
	}
	for key := range values {
		if _, ok := allowed[key]; !ok {
			return applicationdcim.ListInterfaceTemplatesQuery{},
				shared.Invalid(key, "Unsupported filter.")
		}
	}
	query := applicationdcim.ListInterfaceTemplatesQuery{}
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
	deviceTypeIDs, err := parseRackTypeSignedFilters(values["device_type_id"])
	if err != nil {
		return query, shared.Invalid("device_type_id", "A valid integer is required.")
	}
	query.DeviceTypeIDs = deviceTypeIDs
	for _, raw := range values["ordering"] {
		query.Ordering = append(query.Ordering, strings.Split(raw, ",")...)
	}
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

func parseInterfaceTemplateBoolFilter(
	values url.Values,
	name string,
) (*bool, error) {
	rawValues, present := values[name]
	if !present || len(rawValues) == 0 || rawValues[len(rawValues)-1] == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseBool(rawValues[len(rawValues)-1])
	if err != nil {
		return nil, shared.Invalid(name, "A valid boolean is required.")
	}
	return &parsed, nil
}
