package dcim

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

var siteCommandTestTime = shared.NewTimestamp(time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC))

type siteScalarCommandField struct {
	name     string
	concrete string
}

var siteScalarCommandFields = []siteScalarCommandField{
	{name: "name", concrete: "Concrete Site"},
	{name: "slug", concrete: "concrete-site"},
	{name: "status", concrete: dcimdomain.SiteStatusPlanned.String()},
	{name: "facility", concrete: "M9"},
	{name: "description", concrete: "Core site"},
	{name: "comments", concrete: "Maintained"},
}

func TestSiteScalarCommandPresenceMatrix(t *testing.T) {
	t.Parallel()

	t.Run("POST", func(t *testing.T) {
		for _, field := range siteScalarCommandFields {
			field := field
			for _, state := range []struct {
				name  string
				value Field[string]
			}{
				{name: "omitted", value: OmittedField[string]()},
				{name: "null", value: NullField[string]()},
				{name: "blank", value: FieldValue("")},
				{name: "concrete", value: FieldValue(field.concrete)},
			} {
				state := state
				t.Run(field.name+"/"+state.name, func(t *testing.T) {
					command := baselineCreateSiteCommand()
					setCreateSiteCommandField(&command, field.name, state.value)

					values, err := command.values()
					assertPresenceViolation(t, err, expectedSitePresenceViolation(
						"POST", field.name, state.name,
					))
					assert.Equal(t, expectedCreateSiteValues(field, state.name), values)

					if err != nil {
						return
					}
					_, domainErr := dcimdomain.NewSite(values, siteCommandTestTime)
					if state.name == "blank" && (field.name == "name" || field.name == "slug") {
						assertSiteCommandViolation(
							t, domainErr, field.name, "required", "This field may not be blank.",
						)
						return
					}
					require.NoError(t, domainErr)
				})
			}
		}

		for _, status := range []string{" active ", "unknown"} {
			command := baselineCreateSiteCommand()
			command.Status = FieldValue(status)
			values, commandErr := command.values()
			require.NoError(t, commandErr)
			assert.Equal(t, status, values.Status, "command resolution must not trim status")
			_, domainErr := dcimdomain.NewSite(values, siteCommandTestTime)
			assertSiteCommandViolation(
				t,
				domainErr,
				"status",
				"invalid_choice",
				status+" is not a valid choice.",
			)
		}
	})

	t.Run("PUT", func(t *testing.T) {
		for _, field := range siteScalarCommandFields {
			field := field
			for _, state := range []struct {
				name  string
				value Field[string]
			}{
				{name: "omitted", value: OmittedField[string]()},
				{name: "null", value: NullField[string]()},
				{name: "blank", value: FieldValue("")},
				{name: "concrete", value: FieldValue(field.concrete)},
			} {
				state := state
				t.Run(field.name+"/"+state.name, func(t *testing.T) {
					command := baselineReplaceSiteCommand()
					setReplaceSiteCommandField(&command, field.name, state.value)

					patch, err := command.patch()
					assertPresenceViolation(t, err, expectedSitePresenceViolation(
						"PUT", field.name, state.name,
					))
					assert.Equal(t, expectedReplaceSitePatch(field, state.name), patch)

					if err != nil {
						return
					}
					candidate := newSiteCommandTestSite(t)
					domainErr := candidate.ApplyPatch(patch, siteCommandTestTimeAfter())
					if state.name == "blank" && (field.name == "name" || field.name == "slug") {
						assertSiteCommandViolation(
							t, domainErr, field.name, "required", "This field may not be blank.",
						)
						return
					}
					require.NoError(t, domainErr)
				})
			}
		}

		patch, err := (ReplaceSiteCommand{ID: 1}).patch()
		assert.Equal(t, dcimdomain.SitePatch{}, patch)
		assert.Equal(t, []shared.FieldViolation{
			{Field: "name", Reason: "required", Description: "This field is required."},
			{Field: "slug", Reason: "required", Description: "This field is required."},
		}, shared.ViolationsOf(err))

		for _, status := range []string{" active ", "unknown"} {
			command := baselineReplaceSiteCommand()
			command.Status = FieldValue(status)
			patch, commandErr := command.patch()
			require.NoError(t, commandErr)
			require.NotNil(t, patch.Status)
			assert.Equal(t, status, *patch.Status, "command resolution must not trim status")
			candidate := newSiteCommandTestSite(t)
			domainErr := candidate.ApplyPatch(patch, siteCommandTestTimeAfter())
			assertSiteCommandViolation(
				t,
				domainErr,
				"status",
				"invalid_choice",
				status+" is not a valid choice.",
			)
		}
	})

	t.Run("PATCH", func(t *testing.T) {
		for _, field := range siteScalarCommandFields {
			field := field
			for _, state := range []struct {
				name  string
				value Field[string]
			}{
				{name: "omitted", value: OmittedField[string]()},
				{name: "null", value: NullField[string]()},
				{name: "blank", value: FieldValue("")},
				{name: "concrete", value: FieldValue(field.concrete)},
			} {
				state := state
				t.Run(field.name+"/"+state.name, func(t *testing.T) {
					command := UpdateSiteCommand{ID: 1}
					setUpdateSiteCommandField(&command, field.name, state.value)

					patch, err := command.patch()
					assertPresenceViolation(t, err, expectedSitePresenceViolation(
						"PATCH", field.name, state.name,
					))
					assert.Equal(t, expectedUpdateSitePatch(field, state.name), patch)

					if err != nil {
						return
					}
					candidate := newSiteCommandTestSite(t)
					domainErr := candidate.ApplyPatch(patch, siteCommandTestTimeAfter())
					switch {
					case state.name == "omitted":
						assertSiteCommandViolation(
							t,
							domainErr,
							"update_mask",
							"required",
							"At least one writable field must be supplied.",
						)
					case state.name == "blank" && (field.name == "name" || field.name == "slug"):
						assertSiteCommandViolation(
							t, domainErr, field.name, "required", "This field may not be blank.",
						)
					default:
						require.NoError(t, domainErr)
						if state.name == "blank" || state.name == "concrete" {
							assert.Equal(t, patchSiteCommandField(patch, field.name), siteCommandValue(candidate, field.name))
						}
					}
				})
			}
		}

		for _, status := range []string{" active ", "unknown"} {
			command := UpdateSiteCommand{ID: 1, Status: FieldValue(status)}
			patch, commandErr := command.patch()
			require.NoError(t, commandErr)
			require.NotNil(t, patch.Status)
			assert.Equal(t, status, *patch.Status, "command resolution must not trim status")
			candidate := newSiteCommandTestSite(t)
			domainErr := candidate.ApplyPatch(patch, siteCommandTestTimeAfter())
			assertSiteCommandViolation(
				t,
				domainErr,
				"status",
				"invalid_choice",
				status+" is not a valid choice.",
			)
		}

		mixedPatch, err := (UpdateSiteCommand{
			ID:       1,
			Name:     NullField[string](),
			Facility: FieldValue("M10"),
		}).patch()
		assertPresenceViolation(t, err, &shared.FieldViolation{
			Field: "name", Reason: "null", Description: "This field may not be null.",
		})
		assert.Nil(t, mixedPatch.Name)
		require.NotNil(t, mixedPatch.Facility)
		assert.Equal(t, "M10", *mixedPatch.Facility, "valid present fields survive a sibling presence error")
	})
}

func baselineCreateSiteCommand() CreateSiteCommand {
	return CreateSiteCommand{Name: FieldValue("Site"), Slug: FieldValue("site")}
}

func baselineReplaceSiteCommand() ReplaceSiteCommand {
	return ReplaceSiteCommand{
		ID: 1, Name: FieldValue("Replacement"), Slug: FieldValue("replacement"),
	}
}

func setCreateSiteCommandField(command *CreateSiteCommand, field string, value Field[string]) {
	switch field {
	case "name":
		command.Name = value
	case "slug":
		command.Slug = value
	case "status":
		command.Status = value
	case "facility":
		command.Facility = value
	case "description":
		command.Description = value
	case "comments":
		command.Comments = value
	}
}

func setReplaceSiteCommandField(command *ReplaceSiteCommand, field string, value Field[string]) {
	switch field {
	case "name":
		command.Name = value
	case "slug":
		command.Slug = value
	case "status":
		command.Status = value
	case "facility":
		command.Facility = value
	case "description":
		command.Description = value
	case "comments":
		command.Comments = value
	}
}

func setUpdateSiteCommandField(command *UpdateSiteCommand, field string, value Field[string]) {
	switch field {
	case "name":
		command.Name = value
	case "slug":
		command.Slug = value
	case "status":
		command.Status = value
	case "facility":
		command.Facility = value
	case "description":
		command.Description = value
	case "comments":
		command.Comments = value
	}
}

func expectedSitePresenceViolation(operation, field, state string) *shared.FieldViolation {
	if state == "omitted" && operation != "PATCH" && (field == "name" || field == "slug") {
		return &shared.FieldViolation{
			Field: field, Reason: "required", Description: "This field is required.",
		}
	}
	if state == "null" {
		if field == "status" {
			return &shared.FieldViolation{
				Field: field, Reason: "blank", Description: "This field may not be blank.",
			}
		}
		return &shared.FieldViolation{
			Field: field, Reason: "null", Description: "This field may not be null.",
		}
	}
	if state == "blank" && field == "status" {
		return &shared.FieldViolation{
			Field: field, Reason: "blank", Description: "This field may not be blank.",
		}
	}
	return nil
}

func expectedCreateSiteValues(field siteScalarCommandField, state string) dcimdomain.SiteValues {
	values := dcimdomain.SiteValues{
		Name: "Site", Slug: "site", Status: dcimdomain.SiteStatusActive.String(),
	}
	value := ""
	switch state {
	case "concrete":
		value = field.concrete
	case "omitted", "null":
		if field.name == "status" {
			value = dcimdomain.SiteStatusActive.String()
		}
	}
	setSiteValuesField(&values, field.name, value)
	return values
}

func expectedReplaceSitePatch(field siteScalarCommandField, state string) dcimdomain.SitePatch {
	name := "Replacement"
	slug := "replacement"
	patch := dcimdomain.SitePatch{Name: &name, Slug: &slug}
	setExpectedSitePatchField(&patch, field, state)
	return patch
}

func expectedUpdateSitePatch(field siteScalarCommandField, state string) dcimdomain.SitePatch {
	patch := dcimdomain.SitePatch{}
	setExpectedSitePatchField(&patch, field, state)
	return patch
}

func setExpectedSitePatchField(patch *dcimdomain.SitePatch, field siteScalarCommandField, state string) {
	var value *string
	switch state {
	case "blank":
		blank := ""
		value = &blank
	case "concrete":
		concrete := field.concrete
		value = &concrete
	}
	switch field.name {
	case "name":
		patch.Name = value
	case "slug":
		patch.Slug = value
	case "status":
		patch.Status = value
	case "facility":
		patch.Facility = value
	case "description":
		patch.Description = value
	case "comments":
		patch.Comments = value
	}
}

func setSiteValuesField(values *dcimdomain.SiteValues, field, value string) {
	switch field {
	case "name":
		values.Name = value
	case "slug":
		values.Slug = value
	case "status":
		values.Status = value
	case "facility":
		values.Facility = value
	case "description":
		values.Description = value
	case "comments":
		values.Comments = value
	}
}

func patchSiteCommandField(patch dcimdomain.SitePatch, field string) string {
	var value *string
	switch field {
	case "name":
		value = patch.Name
	case "slug":
		value = patch.Slug
	case "status":
		value = patch.Status
	case "facility":
		value = patch.Facility
	case "description":
		value = patch.Description
	case "comments":
		value = patch.Comments
	}
	if value == nil {
		return ""
	}
	return *value
}

func siteCommandValue(site *dcimdomain.Site, field string) string {
	switch field {
	case "name":
		return site.Name()
	case "slug":
		return site.Slug().String()
	case "status":
		return site.Status().String()
	case "facility":
		return site.Facility()
	case "description":
		return site.Description()
	case "comments":
		return site.Comments()
	default:
		return ""
	}
}

func newSiteCommandTestSite(t *testing.T) *dcimdomain.Site {
	t.Helper()
	site, err := dcimdomain.NewSite(dcimdomain.SiteValues{
		Name:        "Original",
		Slug:        "original",
		Status:      dcimdomain.SiteStatusActive.String(),
		Facility:    "M9",
		Description: "Original description",
		Comments:    "Original comments",
	}, siteCommandTestTime)
	require.NoError(t, err)
	return site
}

func siteCommandTestTimeAfter() shared.Timestamp {
	return shared.NewTimestamp(siteCommandTestTime.Add(time.Minute))
}

func assertPresenceViolation(t *testing.T, err error, expected *shared.FieldViolation) {
	t.Helper()
	if expected == nil {
		require.NoError(t, err)
		return
	}
	require.Error(t, err)
	assert.Equal(t, []shared.FieldViolation{*expected}, shared.ViolationsOf(err))
}

func assertSiteCommandViolation(
	t *testing.T,
	err error,
	field string,
	reason string,
	description string,
) {
	t.Helper()
	require.Error(t, err)
	assert.Equal(t, []shared.FieldViolation{{
		Field: field, Reason: reason, Description: description,
	}}, shared.ViolationsOf(err))
}
