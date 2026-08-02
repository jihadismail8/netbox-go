package workflow

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"

	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

type DeviceRoleService interface {
	ListDeviceRoles(context.Context, identity.Principal, applicationdcim.ListDeviceRolesQuery) (applicationdcim.DeviceRolePage, error)
	GetDeviceRole(context.Context, identity.Principal, applicationdcim.GetDeviceRoleQuery) (*domaindcim.DeviceRole, error)
	CreateDeviceRole(context.Context, identity.Principal, applicationdcim.CreateDeviceRoleCommand) (*domaindcim.DeviceRole, error)
	ReplaceDeviceRole(context.Context, identity.Principal, applicationdcim.ReplaceDeviceRoleCommand) (*domaindcim.DeviceRole, error)
	UpdateDeviceRole(context.Context, identity.Principal, applicationdcim.UpdateDeviceRoleCommand) (*domaindcim.DeviceRole, error)
	DeleteDeviceRole(context.Context, identity.Principal, applicationdcim.DeleteDeviceRoleCommand) error
}

var _ DeviceRoleService = (*applicationdcim.DeviceRoleService)(nil)

// DeviceRoleHandler owns the typed DeviceRole HTTP surface. It remains
// standalone so route composition can cut over without retaining a dependency
// on the dynamic workflow resource handler.
type DeviceRoleHandler struct {
	service DeviceRoleService
}

func NewDeviceRoleHandler(service DeviceRoleService) *DeviceRoleHandler {
	if service == nil {
		panic("REST DeviceRole handler requires a typed service")
	}
	return &DeviceRoleHandler{service: service}
}

func (handler *DeviceRoleHandler) Register(r gin.IRoutes, middlewares ...gin.HandlerFunc) {
	const base = "/api/dcim/device-roles/"
	r.GET(base, append(middlewares, handler.list())...)
	r.POST(base, append(middlewares, handler.create())...)
	r.GET(base+":id/", append(middlewares, handler.get())...)
	r.PUT(base+":id/", append(middlewares, handler.replace())...)
	r.PATCH(base+":id/", append(middlewares, handler.update())...)
	r.DELETE(base+":id/", append(middlewares, handler.delete())...)
}

type deviceRoleReferenceDTO struct {
	ID      int64  `json:"id"`
	URL     string `json:"url"`
	Display string `json:"display"`
}

type deviceRoleDTO struct {
	ID          int64                   `json:"id"`
	URL         string                  `json:"url"`
	Display     string                  `json:"display"`
	Parent      *deviceRoleReferenceDTO `json:"parent"`
	Name        string                  `json:"name"`
	Slug        string                  `json:"slug"`
	Color       string                  `json:"color"`
	VMRole      bool                    `json:"vm_role"`
	Description string                  `json:"description"`
	Comments    string                  `json:"comments"`
	Created     time.Time               `json:"created"`
	LastUpdated time.Time               `json:"last_updated"`
	DeviceCount uint64                  `json:"device_count"`
	Depth       uint32                  `json:"_depth"`
}

type deviceRoleListDTO struct {
	Count    uint64          `json:"count"`
	Next     *string         `json:"next"`
	Previous *string         `json:"previous"`
	Results  []deviceRoleDTO `json:"results"`
}

func deviceRoleURL(id shared.ID) string {
	return "/api/dcim/device-roles/" + id.String() + "/"
}

func newDeviceRoleDTO(role *domaindcim.DeviceRole) deviceRoleDTO {
	dto := deviceRoleDTO{
		ID: role.ID().Int64(), URL: deviceRoleURL(role.ID()), Display: role.Display(),
		Name: role.Name(), Slug: role.Slug().String(), Color: role.Color().String(),
		VMRole: role.VMRole(), Description: role.Description(), Comments: role.Comments(),
		Created: role.Created().Time, LastUpdated: role.LastUpdated().Time,
		DeviceCount: role.DeviceCount(), Depth: role.Depth(),
	}
	if parent, present := role.ParentReference(); present {
		dto.Parent = &deviceRoleReferenceDTO{
			ID: parent.ID.Int64(), URL: deviceRoleURL(parent.ID), Display: parent.Display,
		}
	}
	return dto
}

func (handler *DeviceRoleHandler) list() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := Principal(c)
		if !ok {
			writeError(c, shared.Unauthenticated())
			return
		}
		query, err := parseDeviceRoleList(c.Request.URL.Query())
		if err != nil {
			writeError(c, err)
			return
		}
		page, err := handler.service.ListDeviceRoles(c.Request.Context(), principal, query)
		if err != nil {
			writeError(c, err)
			return
		}
		results := make([]deviceRoleDTO, 0, len(page.Results))
		for _, role := range page.Results {
			results = append(results, newDeviceRoleDTO(role))
		}
		c.JSON(http.StatusOK, deviceRoleListDTO{
			Count:    page.Count,
			Next:     organizationPageURL(c, query.Offset, query.EffectiveLimit(), page.Count, true),
			Previous: organizationPageURL(c, query.Offset, query.EffectiveLimit(), page.Count, false),
			Results:  results,
		})
	}
}

func (handler *DeviceRoleHandler) get() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		role, err := handler.service.GetDeviceRole(
			c.Request.Context(), principal, applicationdcim.GetDeviceRoleQuery{ID: id},
		)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newDeviceRoleDTO(role))
	}
}

func (handler *DeviceRoleHandler) create() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := Principal(c)
		if !ok {
			writeError(c, shared.Unauthenticated())
			return
		}
		input, err := decodeDeviceRoleInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		role, err := handler.service.CreateDeviceRole(c.Request.Context(), principal, input.createCommand())
		if err != nil {
			writeError(c, err)
			return
		}
		c.Header("Location", deviceRoleURL(role.ID()))
		c.JSON(http.StatusCreated, newDeviceRoleDTO(role))
	}
}

func (handler *DeviceRoleHandler) replace() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		input, err := decodeDeviceRoleInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		role, err := handler.service.ReplaceDeviceRole(
			c.Request.Context(), principal, input.replaceCommand(id),
		)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newDeviceRoleDTO(role))
	}
}

func (handler *DeviceRoleHandler) update() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		input, err := decodeDeviceRoleInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		role, err := handler.service.UpdateDeviceRole(
			c.Request.Context(), principal, input.updateCommand(id),
		)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, newDeviceRoleDTO(role))
	}
}

func (handler *DeviceRoleHandler) delete() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := organizationRequestIdentity(c)
		if !ok {
			return
		}
		if err := handler.service.DeleteDeviceRole(
			c.Request.Context(), principal, applicationdcim.DeleteDeviceRoleCommand{ID: id},
		); err != nil {
			writeError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

type deviceRoleInputJSON struct {
	Parent      json.RawMessage `json:"parent"`
	Name        json.RawMessage `json:"name"`
	Slug        json.RawMessage `json:"slug"`
	Color       json.RawMessage `json:"color"`
	VMRole      json.RawMessage `json:"vm_role"`
	Description json.RawMessage `json:"description"`
	Comments    json.RawMessage `json:"comments"`
}

type decodedDeviceRoleInput struct {
	parent      applicationdcim.Field[shared.ID]
	name        applicationdcim.Field[string]
	slug        applicationdcim.Field[string]
	color       applicationdcim.Field[string]
	vmRole      applicationdcim.Field[bool]
	description applicationdcim.Field[string]
	comments    applicationdcim.Field[string]
}

func decodeDeviceRoleInput(c *gin.Context) (decodedDeviceRoleInput, error) {
	input, err := decodeTypedObject[deviceRoleInputJSON](c)
	if err != nil {
		return decodedDeviceRoleInput{}, err
	}
	parent, err := decodeRackTypeIDField("parent", input.Parent)
	if err != nil {
		return decodedDeviceRoleInput{}, err
	}
	name, err := decodeSiteStringField("name", input.Name)
	if err != nil {
		return decodedDeviceRoleInput{}, err
	}
	slug, err := decodeSiteStringField("slug", input.Slug)
	if err != nil {
		return decodedDeviceRoleInput{}, err
	}
	color, err := decodeSiteStringField("color", input.Color)
	if err != nil {
		return decodedDeviceRoleInput{}, err
	}
	vmRole, err := decodeRackTypeBoolField("vm_role", input.VMRole)
	if err != nil {
		return decodedDeviceRoleInput{}, err
	}
	description, err := decodeSiteStringField("description", input.Description)
	if err != nil {
		return decodedDeviceRoleInput{}, err
	}
	comments, err := decodeSiteStringField("comments", input.Comments)
	if err != nil {
		return decodedDeviceRoleInput{}, err
	}
	return decodedDeviceRoleInput{
		parent: parent, name: name, slug: slug, color: color, vmRole: vmRole,
		description: description, comments: comments,
	}, nil
}

func (input decodedDeviceRoleInput) createCommand() applicationdcim.CreateDeviceRoleCommand {
	return applicationdcim.CreateDeviceRoleCommand{
		Parent: input.parent, Name: input.name, Slug: input.slug, Color: input.color,
		VMRole: input.vmRole, Description: input.description, Comments: input.comments,
	}
}

func (input decodedDeviceRoleInput) replaceCommand(id shared.ID) applicationdcim.ReplaceDeviceRoleCommand {
	return applicationdcim.ReplaceDeviceRoleCommand{
		ID: id, Parent: input.parent, Name: input.name, Slug: input.slug, Color: input.color,
		VMRole: input.vmRole, Description: input.description, Comments: input.comments,
	}
}

func (input decodedDeviceRoleInput) updateCommand(id shared.ID) applicationdcim.UpdateDeviceRoleCommand {
	return applicationdcim.UpdateDeviceRoleCommand{
		ID: id, Parent: input.parent, Name: input.name, Slug: input.slug, Color: input.color,
		VMRole: input.vmRole, Description: input.description, Comments: input.comments,
	}
}

func parseDeviceRoleList(values url.Values) (applicationdcim.ListDeviceRolesQuery, error) {
	parameters, err := parseOrganizationList(values)
	if err != nil {
		return applicationdcim.ListDeviceRolesQuery{}, err
	}
	return applicationdcim.ListDeviceRolesQuery{
		Limit: parameters.limit, LimitPresent: parameters.limitPresent,
		Offset: parameters.offset, Query: parameters.query, IDs: parameters.ids,
		Ordering: parameters.ordering, Names: parameters.names, Slugs: parameters.slugs,
	}, nil
}
