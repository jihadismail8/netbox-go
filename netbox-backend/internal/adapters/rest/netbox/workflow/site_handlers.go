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

	"github.com/gin-gonic/gin"

	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

type SiteService interface {
	ListSites(context.Context, identity.Principal, applicationdcim.ListSitesQuery) (applicationdcim.SitePage, error)
	GetSite(context.Context, identity.Principal, applicationdcim.GetSiteQuery) (*domaindcim.Site, error)
	CreateSite(context.Context, identity.Principal, applicationdcim.CreateSiteCommand) (*domaindcim.Site, error)
	ReplaceSite(context.Context, identity.Principal, applicationdcim.ReplaceSiteCommand) (*domaindcim.Site, error)
	UpdateSite(context.Context, identity.Principal, applicationdcim.UpdateSiteCommand) (*domaindcim.Site, error)
	DeleteSite(context.Context, identity.Principal, applicationdcim.DeleteSiteCommand) error
}

func (h *Handler) registerSites(r gin.IRoutes, middlewares ...gin.HandlerFunc) {
	const base = "/api/dcim/sites/"
	r.GET(base, append(middlewares, h.listSites())...)
	r.POST(base, append(middlewares, h.createSite())...)
	r.GET(base+":id/", append(middlewares, h.getSite())...)
	r.PUT(base+":id/", append(middlewares, h.replaceSite())...)
	r.PATCH(base+":id/", append(middlewares, h.updateSite())...)
	r.DELETE(base+":id/", append(middlewares, h.deleteSite())...)
}

func (h *Handler) listSites() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := Principal(c)
		if !ok {
			writeError(c, shared.Unauthenticated())
			return
		}
		query, err := parseSiteList(c.Request.URL.Query())
		if err != nil {
			writeError(c, err)
			return
		}
		page, err := h.sites.ListSites(c.Request.Context(), principal, query)
		if err != nil {
			writeError(c, err)
			return
		}
		results := make([]gin.H, 0, len(page.Results))
		for _, site := range page.Results {
			results = append(results, siteDTO(site, true))
		}
		c.JSON(http.StatusOK, gin.H{
			"count":    page.Count,
			"next":     sitePageURL(c, query, page.Count, true),
			"previous": sitePageURL(c, query, page.Count, false),
			"results":  results,
		})
	}
}

func (h *Handler) getSite() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := siteRequestIdentity(c)
		if !ok {
			return
		}
		site, err := h.sites.GetSite(c.Request.Context(), principal, applicationdcim.GetSiteQuery{ID: id})
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, siteDTO(site, true))
	}
}

func (h *Handler) createSite() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := Principal(c)
		if !ok {
			writeError(c, shared.Unauthenticated())
			return
		}
		input, err := decodeSiteInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		command := input.createCommand()
		site, err := h.sites.CreateSite(c.Request.Context(), principal, command)
		if err != nil {
			writeError(c, err)
			return
		}
		c.Header("Location", siteURL(site.ID()))
		c.JSON(http.StatusCreated, siteDTO(site, false))
	}
}

func (h *Handler) replaceSite() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := siteRequestIdentity(c)
		if !ok {
			return
		}
		input, err := decodeSiteInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		command := input.replaceCommand(id)
		site, err := h.sites.ReplaceSite(c.Request.Context(), principal, command)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, siteDTO(site, true))
	}
}

func (h *Handler) updateSite() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := siteRequestIdentity(c)
		if !ok {
			return
		}
		input, err := decodeSiteInput(c)
		if err != nil {
			writeError(c, err)
			return
		}
		command := input.updateCommand(id)
		site, err := h.sites.UpdateSite(c.Request.Context(), principal, command)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, siteDTO(site, true))
	}
}

func (h *Handler) deleteSite() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, id, ok := siteRequestIdentity(c)
		if !ok {
			return
		}
		if err := h.sites.DeleteSite(
			c.Request.Context(),
			principal,
			applicationdcim.DeleteSiteCommand{ID: id},
		); err != nil {
			writeError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func siteRequestIdentity(c *gin.Context) (identity.Principal, shared.ID, bool) {
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

func siteDTO(site *domaindcim.Site, includeRelatedCounts bool) gin.H {
	status := choiceEnvelope(site.Status().String(), commonChoiceLabels)
	out := gin.H{
		"id":           site.ID().Int64(),
		"url":          siteURL(site.ID()),
		"display":      site.Display(),
		"name":         site.Name(),
		"slug":         site.Slug().String(),
		"status":       status,
		"facility":     site.Facility(),
		"description":  site.Description(),
		"comments":     site.Comments(),
		"created":      site.Created().Time,
		"last_updated": site.LastUpdated().Time,
	}
	if includeRelatedCounts {
		out["device_count"] = site.DeviceCount()
		out["prefix_count"] = site.PrefixCount()
		out["rack_count"] = site.RackCount()
	}
	return out
}

func siteURL(id shared.ID) string {
	return "/api/dcim/sites/" + id.String() + "/"
}

type siteInputJSON struct {
	Name        json.RawMessage `json:"name"`
	Slug        json.RawMessage `json:"slug"`
	Status      json.RawMessage `json:"status"`
	Facility    json.RawMessage `json:"facility"`
	Description json.RawMessage `json:"description"`
	Comments    json.RawMessage `json:"comments"`
}

type decodedSiteInput struct {
	name        applicationdcim.Field[string]
	slug        applicationdcim.Field[string]
	status      applicationdcim.Field[string]
	facility    applicationdcim.Field[string]
	description applicationdcim.Field[string]
	comments    applicationdcim.Field[string]
}

func decodeSiteInput(c *gin.Context) (decodedSiteInput, error) {
	raw, err := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20))
	if err != nil {
		return decodedSiteInput{}, shared.Invalid("non_field_errors", "Expected a JSON object.")
	}
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return decodedSiteInput{}, shared.Invalid("non_field_errors", "Expected a JSON object.")
	}

	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.DisallowUnknownFields()
	var input siteInputJSON
	if err := decoder.Decode(&input); err != nil {
		if field, unknown := unknownJSONField(err); unknown {
			return decodedSiteInput{}, shared.Invalid(
				field,
				"This field is not supported by the active capability profile.")

		}
		return decodedSiteInput{}, shared.Invalid("non_field_errors", "Expected a JSON object.")
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return decodedSiteInput{}, shared.Invalid(
			"non_field_errors",
			"Request body must contain one JSON object.")

	}

	name, err := decodeSiteStringField("name", input.Name)
	if err != nil {
		return decodedSiteInput{}, err
	}
	slug, err := decodeSiteStringField("slug", input.Slug)
	if err != nil {
		return decodedSiteInput{}, err
	}
	status, err := decodeSiteStringField("status", input.Status)
	if err != nil {
		return decodedSiteInput{}, err
	}
	facility, err := decodeSiteStringField("facility", input.Facility)
	if err != nil {
		return decodedSiteInput{}, err
	}
	description, err := decodeSiteStringField("description", input.Description)
	if err != nil {
		return decodedSiteInput{}, err
	}
	comments, err := decodeSiteStringField("comments", input.Comments)
	if err != nil {
		return decodedSiteInput{}, err
	}
	return decodedSiteInput{
		name: name, slug: slug, status: status, facility: facility,
		description: description, comments: comments,
	}, nil
}

func decodeSiteStringField(name string, raw json.RawMessage) (applicationdcim.Field[string], error) {
	if len(raw) == 0 {
		return applicationdcim.OmittedField[string](), nil
	}
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return applicationdcim.NullField[string](), nil
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return applicationdcim.Field[string]{}, shared.Invalid(name, "Must be a string.")
	}
	return applicationdcim.FieldValue(value), nil
}

func unknownJSONField(err error) (string, bool) {
	const prefix = "json: unknown field \""
	message := err.Error()
	if !strings.HasPrefix(message, prefix) || !strings.HasSuffix(message, "\"") {
		return "", false
	}
	return strings.TrimSuffix(strings.TrimPrefix(message, prefix), "\""), true
}

func (input decodedSiteInput) createCommand() applicationdcim.CreateSiteCommand {
	return applicationdcim.CreateSiteCommand{
		Name: input.name, Slug: input.slug, Status: input.status,
		Facility: input.facility, Description: input.description, Comments: input.comments,
	}
}

func (input decodedSiteInput) replaceCommand(id shared.ID) applicationdcim.ReplaceSiteCommand {
	return applicationdcim.ReplaceSiteCommand{
		ID: id, Name: input.name, Slug: input.slug, Status: input.status,
		Facility: input.facility, Description: input.description, Comments: input.comments,
	}
}

func (input decodedSiteInput) updateCommand(id shared.ID) applicationdcim.UpdateSiteCommand {
	return applicationdcim.UpdateSiteCommand{
		ID: id, Name: input.name, Slug: input.slug, Status: input.status,
		Facility: input.facility, Description: input.description, Comments: input.comments,
	}
}

func parseSiteList(values url.Values) (applicationdcim.ListSitesQuery, error) {
	allowed := map[string]struct{}{
		"limit": {}, "offset": {}, "q": {}, "ordering": {}, "id": {},
		"name": {}, "slug": {}, "status": {},
	}
	for key := range values {
		if _, ok := allowed[key]; !ok {
			return applicationdcim.ListSitesQuery{}, shared.Invalid(key, "Unsupported filter.")
		}
	}

	query := applicationdcim.ListSitesQuery{}
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
	query.Slugs = append([]string(nil), values["slug"]...)
	query.Statuses = append([]string(nil), values["status"]...)
	return query, nil
}

func sitePageURL(
	c *gin.Context,
	query applicationdcim.ListSitesQuery,
	count uint64,
	next bool,
) any {
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
	return c.Request.URL.Path + "?" + values.Encode()
}
