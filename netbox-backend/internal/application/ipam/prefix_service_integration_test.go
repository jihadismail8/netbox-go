package ipam_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	postgreschangelog "netbox-go/internal/adapters/postgres/changelog"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	identityrows "netbox-go/internal/adapters/postgres/identity"
	postgresipam "netbox-go/internal/adapters/postgres/ipam"
	ipamrow "netbox-go/internal/adapters/postgres/ipam/row"
	postgrestransaction "netbox-go/internal/adapters/postgres/transaction"
	"netbox-go/internal/application/authz"
	"netbox-go/internal/application/changelog"
	applicationipam "netbox-go/internal/application/ipam"
	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

var prefixApplicationFixtureSequence atomic.Uint64

func TestPrefixServiceEnforcesGlobalAndPerVRFUniqueness(t *testing.T) {
	db, service, _, vrfs := newPrefixApplicationService(t, nil, &authz.AllowAll{})
	principal := testPrincipal()

	global, err := service.CreatePrefix(t.Context(), principal, createPrefixCommand("192.0.2.0/24", nil))
	require.NoError(t, err)
	assert.Equal(t, domainipam.PrefixStatusActive, global.Status())
	assert.False(t, global.IsPool())
	assert.False(t, global.MarkUtilized())
	assert.True(t, global.VRF().IsNull())

	_, err = service.CreatePrefix(t.Context(), principal, createPrefixCommand("192.0.2.0/24", nil))
	require.Error(t, err)
	assert.Equal(t, "prefix", shared.ViolationsOf(err)[0].Field)
	assert.Contains(t, shared.ViolationsOf(err)[0].Description, "global table")

	nonUnique := createApplicationVRF(t, vrfs, "Non-unique", "65000:1", false)
	for range 2 {
		_, err = service.CreatePrefix(
			t.Context(), principal, createPrefixCommand("198.51.100.0/24", idPointer(nonUnique.ID())),
		)
		require.NoError(t, err)
	}

	unique := createApplicationVRF(t, vrfs, "Unique", "65000:2", true)
	_, err = service.CreatePrefix(
		t.Context(), principal, createPrefixCommand("203.0.113.0/24", idPointer(unique.ID())),
	)
	require.NoError(t, err)
	_, err = service.CreatePrefix(
		t.Context(), principal, createPrefixCommand("203.0.113.0/24", idPointer(unique.ID())),
	)
	require.Error(t, err)
	assert.Contains(t, shared.ViolationsOf(err)[0].Description, "VRF Unique (65000:2)")

	var prefixCount, changeCount int64
	require.NoError(t, db.Model(&ipamrow.PrefixRow{}).Count(&prefixCount).Error)
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Count(&changeCount).Error)
	assert.Equal(t, int64(4), prefixCount)
	assert.Equal(t, int64(4), changeCount)
}

func TestPrefixServiceProjectsHierarchyAndExactListSemantics(t *testing.T) {
	_, service, _, _ := newPrefixApplicationService(t, nil, &authz.AllowAll{})
	principal := testPrincipal()
	outer, err := service.CreatePrefix(t.Context(), principal, createPrefixCommand("10.0.0.0/8", nil))
	require.NoError(t, err)
	inner, err := service.CreatePrefix(t.Context(), principal, createPrefixCommand("10.1.0.0/16", nil))
	require.NoError(t, err)
	_, err = service.CreatePrefix(t.Context(), principal, createPrefixCommand("2001:db8::/32", nil))
	require.NoError(t, err)

	outer, err = service.GetPrefix(t.Context(), principal, applicationipam.GetPrefixQuery{ID: outer.ID()})
	require.NoError(t, err)
	inner, err = service.GetPrefix(t.Context(), principal, applicationipam.GetPrefixQuery{ID: inner.ID()})
	require.NoError(t, err)
	assert.Equal(t, uint64(1), outer.Children())
	assert.Equal(t, uint32(0), outer.Depth())
	assert.Equal(t, uint64(0), inner.Children())
	assert.Equal(t, uint32(1), inner.Depth())

	zero := uint32(0)
	within := "10.0.0.0/8"
	page, err := service.ListPrefixes(t.Context(), principal, applicationipam.ListPrefixesQuery{
		Limit: zero, LimitPresent: true, IDs: []int64{-1, outer.ID().Int64(), inner.ID().Int64()},
		Prefixes: []string{"invalid", "10.0.0.0/8", "10.1.2.3/16"},
		Family:   int64PointerForPrefix(4), WithinInclude: &within,
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(2), page.Count)
	require.Len(t, page.Results, 2)
	assert.Equal(t, applicationipam.MaximumPrefixPageLimit, applicationipam.ListPrefixesQuery{
		LimitPresent: true,
	}.EffectiveLimit())
	assert.Equal(t, "10.0.0.0/8", page.Results[0].Display())
	assert.Equal(t, "10.1.0.0/16", page.Results[1].Display())
}

func TestPrefixServiceRollsBackMutationAndChangeTogether(t *testing.T) {
	db, service, _, _ := newPrefixApplicationService(
		t, failingPrefixRecorder{}, &authz.AllowAll{},
	)
	_, err := service.CreatePrefix(
		t.Context(), testPrincipal(), createPrefixCommand("192.0.2.0/24", nil),
	)
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonInternal))

	var count int64
	require.NoError(t, db.Model(&ipamrow.PrefixRow{}).Count(&count).Error)
	assert.Zero(t, count)
}

func TestPrefixServiceDeniedCreateStopsBeforeWrite(t *testing.T) {
	db, service, _, _ := newPrefixApplicationService(
		t, nil, &trackingAuthorizer{denyAction: authz.Add},
	)
	_, err := service.CreatePrefix(
		t.Context(), testPrincipal(), createPrefixCommand("192.0.2.0/24", nil),
	)
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonForbidden))
	var count int64
	require.NoError(t, db.Model(&ipamrow.PrefixRow{}).Count(&count).Error)
	assert.Zero(t, count)
}

func TestReplacePrefixPreservesOmittedOptionalState(t *testing.T) {
	db, seedService, repository, vrfs := newPrefixApplicationService(t, nil, &authz.AllowAll{})
	vrf := createApplicationVRF(t, vrfs, "Scalar VRF", "65000:9", true)
	command := createPrefixCommand("192.0.2.0/24", idPointer(vrf.ID()))
	command.Status = applicationipam.FieldValue("reserved")
	command.IsPool, command.MarkUtilized = applicationipam.FieldValue(true), applicationipam.FieldValue(true)
	command.Description, command.Comments = applicationipam.FieldValue("original description"), applicationipam.FieldValue("original comments")
	seed, err := seedService.CreatePrefix(t.Context(), testPrincipal(), command)
	require.NoError(t, err)
	before := seed.Snapshot()
	clock := &prefixScalarCountingClock{now: shared.NewTimestamp(applicationUpdatedAt.Add(time.Hour))}
	service, err := applicationipam.NewPrefixService(
		repository, vrfs, postgrestransaction.NewUnitOfWork(db),
		postgreschangelog.NewRecorder(db), &authz.AllowAll{}, clock,
	)
	require.NoError(t, err)

	updated, err := service.ReplacePrefix(t.Context(), testPrincipal(), applicationipam.ReplacePrefixCommand{
		ID: seed.ID(), CreatePrefixCommand: createPrefixCommand("198.51.100.0/24", nil),
	})
	require.NoError(t, err)
	want := before
	want.Prefix = "198.51.100.0/24"
	assert.Equal(t, want, updated.Snapshot())
	assert.Equal(t, seed.ID(), updated.ID())
	assert.Equal(t, seed.Created(), updated.Created())
	assert.Equal(t, clock.now, updated.LastUpdated())
	var row ipamrow.PrefixRow
	require.NoError(t, db.First(&row, seed.ID().Int64()).Error)
	require.NotNil(t, row.VRFID)
	assert.Equal(t, vrf.ID().Int64(), *row.VRFID)
	assert.Equal(t, "reserved", row.Status)
	assert.True(t, row.IsPool)
	assert.True(t, row.MarkUtilized)
	assert.Equal(t, "original description", row.Description)
	assert.Equal(t, "original comments", row.Comments)
	assert.Equal(t, seed.Created().Time, row.Created)
	assert.Equal(t, clock.now.Time, row.LastUpdated)
	var changes []postgreschangelog.ChangeRow
	require.NoError(t, db.Order("id").Find(&changes).Error)
	require.Len(t, changes, 2)
	assert.Equal(t, int64(17), changes[1].ActorID)
	assert.Equal(t, domainipam.PrefixObjectType, changes[1].Kind)
	assert.Equal(t, seed.ID().Int64(), changes[1].ObjectID)
	assert.Equal(t, string(changelog.ActionUpdate), changes[1].Action)
	assert.Equal(t, clock.now.Time, changes[1].OccurredAt)
	assert.JSONEq(t, prefixScalarSnapshotJSON(t, before), string(changes[1].BeforeData))
	assert.JSONEq(t, prefixScalarSnapshotJSON(t, want), string(changes[1].AfterData))

	updated, err = service.UpdatePrefix(t.Context(), testPrincipal(), applicationipam.UpdatePrefixCommand{
		ID: seed.ID(), VRF: applicationipam.NullField[int64](),
		IsPool: applicationipam.FieldValue(false), MarkUtilized: applicationipam.FieldValue(false),
		Description: applicationipam.FieldValue(""), Comments: applicationipam.FieldValue(""),
	})
	require.NoError(t, err)
	want.VRF = domainipam.NullVRFReference()
	want.IsPool, want.MarkUtilized, want.Description, want.Comments = false, false, "", ""
	assert.Equal(t, want, updated.Snapshot())
	updated, err = service.ReplacePrefix(t.Context(), testPrincipal(), applicationipam.ReplacePrefixCommand{
		ID: seed.ID(), CreatePrefixCommand: createPrefixCommand("2001:db8:1::/64", nil),
	})
	require.NoError(t, err)
	want.Prefix = "2001:db8:1::/64"
	assert.Equal(t, want, updated.Snapshot())
	assert.Equal(t, seed.Created(), updated.Created())
	assert.Equal(t, shared.NewTimestamp(clock.now.Add(2*time.Hour)), updated.LastUpdated())
	assert.Equal(t, 3, clock.calls)
	reloaded, err := repository.Get(t.Context(), seed.ID())
	require.NoError(t, err)
	assert.Equal(t, updated.State(), reloaded.State())
	row = ipamrow.PrefixRow{}
	require.NoError(t, db.First(&row, seed.ID().Int64()).Error)
	assert.Nil(t, row.VRFID)
	assert.Equal(t, "reserved", row.Status)
	assert.False(t, row.IsPool)
	assert.False(t, row.MarkUtilized)
	assert.Empty(t, row.Description)
	assert.Empty(t, row.Comments)
	changes = nil
	require.NoError(t, db.Order("id").Find(&changes).Error)
	assert.Len(t, changes, 4)
}

func TestPrefixScalarValidationLeavesStateUnchanged(t *testing.T) {
	null := func(field string) []shared.FieldViolation {
		return []shared.FieldViolation{{Field: field, Reason: "null", Description: "This field may not be null."}}
	}
	cases := []struct {
		name   string
		change func(*applicationipam.CreatePrefixCommand)
		want   []shared.FieldViolation
	}{
		{"null Prefix", func(c *applicationipam.CreatePrefixCommand) { c.Prefix = applicationipam.NullField[string]() }, null("prefix")},
		{"null status", func(c *applicationipam.CreatePrefixCommand) { c.Status = applicationipam.NullField[string]() },
			[]shared.FieldViolation{{Field: "status", Reason: "invalid", Description: "This field may not be blank."}}},
		{"null is_pool", func(c *applicationipam.CreatePrefixCommand) { c.IsPool = applicationipam.NullField[bool]() }, null("is_pool")},
		{"null mark_utilized", func(c *applicationipam.CreatePrefixCommand) { c.MarkUtilized = applicationipam.NullField[bool]() }, null("mark_utilized")},
		{"null description", func(c *applicationipam.CreatePrefixCommand) { c.Description = applicationipam.NullField[string]() }, null("description")},
		{"null comments", func(c *applicationipam.CreatePrefixCommand) { c.Comments = applicationipam.NullField[string]() }, null("comments")},
		{"blank Prefix", func(c *applicationipam.CreatePrefixCommand) { c.Prefix = applicationipam.FieldValue("") },
			[]shared.FieldViolation{{Field: "prefix", Reason: "blank", Description: "This field may not be blank."}}},
		{"whitespace Prefix", func(c *applicationipam.CreatePrefixCommand) { c.Prefix = applicationipam.FieldValue(" \t\n") },
			[]shared.FieldViolation{{Field: "prefix", Reason: "blank", Description: "This field may not be blank."}}},
		{"blank status", func(c *applicationipam.CreatePrefixCommand) { c.Status = applicationipam.FieldValue("") },
			[]shared.FieldViolation{{Field: "status", Reason: "invalid", Description: "This field may not be blank."}}},
		{"padded status", func(c *applicationipam.CreatePrefixCommand) { c.Status = applicationipam.FieldValue(" active ") },
			[]shared.FieldViolation{{Field: "status", Reason: "invalid", Description: " active  is not a valid choice."}}},
		{"uppercase status", func(c *applicationipam.CreatePrefixCommand) { c.Status = applicationipam.FieldValue("ACTIVE") },
			[]shared.FieldViolation{{Field: "status", Reason: "invalid", Description: "ACTIVE is not a valid choice."}}},
		{"description length and NUL", func(c *applicationipam.CreatePrefixCommand) {
			c.Description = applicationipam.FieldValue(strings.Repeat("界", 200) + "\x00")
		}, []shared.FieldViolation{
			{Field: "description", Reason: "max_length", Description: "Ensure this field has no more than 200 characters."},
			{Field: "description", Reason: "null_characters_not_allowed", Description: "Null characters are not allowed."},
		}},
		{"comments NUL", func(c *applicationipam.CreatePrefixCommand) { c.Comments = applicationipam.FieldValue("note\x00") },
			[]shared.FieldViolation{{Field: "comments", Reason: "null_characters_not_allowed", Description: "Null characters are not allowed."}}},
	}
	for _, id := range []int64{0, -9, 9999} {
		cases = append(cases, struct {
			name   string
			change func(*applicationipam.CreatePrefixCommand)
			want   []shared.FieldViolation
		}{
			fmt.Sprintf("VRF %d", id),
			func(c *applicationipam.CreatePrefixCommand) { c.VRF = applicationipam.FieldValue(id) },
			[]shared.FieldViolation{{Field: "vrf", Reason: "invalid",
				Description: fmt.Sprintf("Related object not found using the provided numeric ID: %d", id)}},
		})
	}
	for _, operation := range []string{"POST", "PUT", "PATCH"} {
		for _, test := range cases {
			t.Run(operation+"/"+test.name, func(t *testing.T) {
				db, service, observer, recorder, clock, seed := newPrefixScalarObservedService(t, &authz.AllowAll{})
				before := prefixScalarDatabaseState(t, db)
				command := createPrefixCommand("198.51.100.0/24", nil)
				test.change(&command)
				result, err := invokePrefixScalarMutation(t, service, operation, seed.ID(), command)
				assert.Nil(t, result)
				assert.Equal(t, test.want, shared.ViolationsOf(err))
				assert.Equal(t, 0, clock.calls, "owned validation must precede the clock")
				assert.Zero(t, observer.writes)
				assert.Zero(t, recorder.calls)
				assert.Equal(t, before, prefixScalarDatabaseState(t, db))
				if observer.loaded != nil {
					assert.Equal(t, observer.before, observer.loaded.State())
				}
			})
		}
		for _, omitPrefix := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/mixed omitted Prefix %t", operation, omitPrefix), func(t *testing.T) {
				db, service, observer, recorder, clock, seed := newPrefixScalarObservedService(t, &authz.AllowAll{})
				before := prefixScalarDatabaseState(t, db)
				command := applicationipam.CreatePrefixCommand{
					Prefix: applicationipam.NullField[string](), VRF: applicationipam.FieldValue(int64(-9)),
					Status: applicationipam.NullField[string](), IsPool: applicationipam.NullField[bool](),
					MarkUtilized: applicationipam.NullField[bool](),
					Description:  applicationipam.FieldValue(strings.Repeat("界", 200) + "\x00"),
					Comments:     applicationipam.FieldValue("note\x00"),
				}
				var want []shared.FieldViolation
				if omitPrefix {
					command.Prefix = applicationipam.OmittedField[string]()
					if operation != "PATCH" {
						want = append(want, shared.FieldViolation{Field: "prefix", Reason: "required", Description: "This field is required."})
					}
				} else {
					want = append(want, null("prefix")...)
				}
				want = append(want,
					shared.FieldViolation{Field: "vrf", Reason: "invalid", Description: "Related object not found using the provided numeric ID: -9"},
					shared.FieldViolation{Field: "status", Reason: "invalid", Description: "This field may not be blank."},
					null("is_pool")[0], null("mark_utilized")[0],
					shared.FieldViolation{Field: "description", Reason: "max_length", Description: "Ensure this field has no more than 200 characters."},
					shared.FieldViolation{Field: "description", Reason: "null_characters_not_allowed", Description: "Null characters are not allowed."},
					shared.FieldViolation{Field: "comments", Reason: "null_characters_not_allowed", Description: "Null characters are not allowed."},
				)
				result, err := invokePrefixScalarMutation(t, service, operation, seed.ID(), command)
				assert.Nil(t, result)
				assert.Equal(t, want, shared.ViolationsOf(err))
				assert.Zero(t, clock.calls)
				assert.Zero(t, observer.writes)
				assert.Zero(t, recorder.calls)
				assert.Equal(t, before, prefixScalarDatabaseState(t, db))
				if observer.loaded != nil {
					assert.Equal(t, observer.before, observer.loaded.State())
				}
			})
		}
	}

	for _, operation := range []string{"POST", "PUT", "PATCH"} {
		for _, boundary := range []string{"model denial", "object denial", "missing target", "zero clock"} {
			if operation == "POST" && (boundary == "object denial" || boundary == "missing target") {
				continue
			}
			t.Run(operation+"/"+boundary, func(t *testing.T) {
				authorizer := &trackingAuthorizer{}
				db, service, observer, recorder, clock, seed := newPrefixScalarObservedService(t, authorizer)
				before := prefixScalarDatabaseState(t, db)
				id, reason := seed.ID(), shared.ErrorReasonForbidden
				switch boundary {
				case "model denial":
					authorizer.denyAction = authz.Change
					if operation == "POST" {
						authorizer.denyAction = authz.Add
					}
				case "object denial":
					authorizer.deniedObjectIDs = map[int64]struct{}{id.Int64(): {}}
				case "missing target":
					id, reason = 9999, shared.ErrorReasonNotFound
				case "zero clock":
					clock.now, reason = shared.Timestamp{}, shared.ErrorReasonInternal
				}
				result, err := invokePrefixScalarMutation(t, service, operation, id, createPrefixCommand("198.51.100.0/24", nil))
				assert.Nil(t, result)
				assert.True(t, shared.HasReason(err, reason))
				wantCalls := 0
				if boundary == "zero clock" {
					wantCalls = 1
				}
				assert.Equal(t, wantCalls, clock.calls)
				assert.Zero(t, observer.writes)
				assert.Zero(t, recorder.calls)
				assert.Equal(t, before, prefixScalarDatabaseState(t, db))
				if observer.loaded != nil {
					assert.Equal(t, observer.before, observer.loaded.State())
				}
			})
		}
	}
}

func TestPrefixNetworkMutationStateAndEffects(t *testing.T) {
	cases := []struct {
		name, input, canonical, description string
		family                              uint32
		bits                                int
	}{
		{name: "IPv4 network", input: "198.51.100.0/24", canonical: "198.51.100.0/24", family: 4, bits: 24},
		{name: "IPv4 bare host", input: "203.0.113.17", canonical: "203.0.113.17/32", family: 4, bits: 32},
		{name: "IPv6 network", input: "2001:db8:10::/64", canonical: "2001:db8:10::/64", family: 6, bits: 64},
		{name: "expanded IPv6 bare host", input: "2001:0DB8:0020:0000:0000:0000:0000:0017", canonical: "2001:db8:20::17/128", family: 6, bits: 128},
		{name: "expanded IPv6 network", input: "2001:0DB8:0030:0000:0000:0000:0000:0000/64", canonical: "2001:db8:30::/64", family: 6, bits: 64},
		{name: "IPv4 host bits", input: "203.0.113.17/24", description: "203.0.113.17/24 is not a valid prefix. Did you mean 203.0.113.0/24?"},
		{name: "IPv6 host bits", input: "2001:db8:40::17/64", description: "2001:db8:40::17/64 is not a valid prefix. Did you mean 2001:db8:40::/64?"},
		{name: "IPv4 slash zero", input: "0.0.0.0/0", description: "Cannot create prefix with /0 mask."},
		{name: "IPv6 slash zero", input: "::/0", description: "Cannot create prefix with /0 mask."},
	}
	for _, operation := range []string{"POST", "PUT", "PATCH"} {
		for _, test := range cases {
			t.Run(operation+"/"+test.name, func(t *testing.T) {
				db, service, observer, recorder, clock, seed := newPrefixScalarObservedService(t, &authz.AllowAll{})
				before := prefixScalarDatabaseState(t, db)
				result, err := invokePrefixScalarMutation(t, service, operation, seed.ID(), createPrefixCommand(test.input, nil))
				if test.description != "" {
					require.Error(t, err)
					assert.Nil(t, result)
					assert.Equal(t, []shared.FieldViolation{{Field: "prefix", Reason: "invalid", Description: test.description}}, shared.ViolationsOf(err))
					assert.Zero(t, clock.calls)
					assert.Zero(t, observer.writes)
					assert.Zero(t, recorder.calls)
					assert.Equal(t, before, prefixScalarDatabaseState(t, db))
					if operation != "POST" {
						require.NotNil(t, observer.loaded)
						assert.Equal(t, observer.before, observer.loaded.State())
					}
					return
				}
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, 1, clock.calls)
				assert.Equal(t, 1, observer.writes)
				assert.Equal(t, 1, recorder.calls)
				assert.True(t, result.ID().IsValid())
				want := domainipam.PrefixSnapshot{Prefix: test.canonical, VRF: domainipam.NullVRFReference(), Status: "active"}
				created, action := clock.now, changelog.ActionCreate
				if operation != "POST" {
					want.VRF, want.Status = seed.VRF(), "reserved"
					want.IsPool, want.MarkUtilized = true, true
					want.Description, want.Comments = "original description", "original comments"
					created, action = seed.Created(), changelog.ActionUpdate
					assert.Equal(t, seed.ID(), result.ID())
				} else {
					assert.NotEqual(t, seed.ID(), result.ID())
				}
				assert.Equal(t, test.family, result.Family())
				assert.Equal(t, test.bits, result.Network().Bits())
				assert.Equal(t, test.canonical, result.Display())
				assert.Equal(t, want, result.Snapshot())
				assert.Equal(t, domainipam.PrefixState{
					ID: result.ID(), Prefix: test.canonical, VRF: want.VRF, Status: want.Status,
					IsPool: want.IsPool, MarkUtilized: want.MarkUtilized,
					Description: want.Description, Comments: want.Comments,
					Created: created, LastUpdated: clock.now, Children: 0, Depth: 0,
				}, result.State())
				reloaded, err := observer.Get(t.Context(), result.ID())
				require.NoError(t, err)
				assert.Equal(t, result.State(), reloaded.State())
				after := prefixScalarDatabaseState(t, db)
				assert.Equal(t, before.vrfs, after.vrfs)
				require.Len(t, after.changes, len(before.changes)+1)
				assert.Equal(t, before.changes, after.changes[:len(before.changes)])
				rowIndex := 0
				if operation == "POST" {
					require.Len(t, after.prefixes, len(before.prefixes)+1)
					assert.Equal(t, before.prefixes, after.prefixes[:len(before.prefixes)])
					rowIndex = len(before.prefixes)
				} else {
					require.Len(t, after.prefixes, len(before.prefixes))
				}
				row := after.prefixes[rowIndex]
				assert.Equal(t, result.ID().Int64(), row.ID)
				assert.Equal(t, test.canonical, row.Prefix)
				assert.Equal(t, want.Status, row.Status)
				assert.Equal(t, want.IsPool, row.IsPool)
				assert.Equal(t, want.MarkUtilized, row.MarkUtilized)
				assert.Equal(t, want.Description, row.Description)
				assert.Equal(t, want.Comments, row.Comments)
				assert.Equal(t, created.Time, row.Created)
				assert.Equal(t, clock.now.Time, row.LastUpdated)
				if operation == "POST" {
					assert.Nil(t, row.VRFID)
				} else {
					require.NotNil(t, row.VRFID)
					reference, present := seed.VRF().Get()
					require.True(t, present)
					assert.Equal(t, reference.ID().Int64(), *row.VRFID)
				}
				change := after.changes[len(before.changes)]
				assert.Equal(t, int64(17), change.ActorID)
				assert.Equal(t, domainipam.PrefixObjectType, change.Kind)
				assert.Equal(t, result.ID().Int64(), change.ObjectID)
				assert.Equal(t, string(action), change.Action)
				assert.Equal(t, clock.now.Time, change.OccurredAt)
				if operation == "POST" {
					assert.Empty(t, change.BeforeData)
				} else {
					assert.JSONEq(t, prefixScalarSnapshotJSON(t, domainipam.PrefixSnapshot{
						Prefix: "192.0.2.0/24", VRF: seed.VRF(), Status: "reserved", IsPool: true, MarkUtilized: true,
						Description: "original description", Comments: "original comments",
					}), string(change.BeforeData))
				}
				assert.JSONEq(t, prefixScalarSnapshotJSON(t, want), string(change.AfterData))
			})
		}
	}
}

func invokePrefixScalarMutation(
	t *testing.T, service *applicationipam.PrefixService, operation string, id shared.ID,
	command applicationipam.CreatePrefixCommand,
) (*domainipam.Prefix, error) {
	t.Helper()
	switch operation {
	case "POST":
		return service.CreatePrefix(t.Context(), testPrincipal(), command)
	case "PUT":
		return service.ReplacePrefix(t.Context(), testPrincipal(), applicationipam.ReplacePrefixCommand{
			ID: id, CreatePrefixCommand: command,
		})
	case "PATCH":
		return service.UpdatePrefix(t.Context(), testPrincipal(), applicationipam.UpdatePrefixCommand{
			ID: id, Prefix: command.Prefix, VRF: command.VRF, Status: command.Status,
			IsPool: command.IsPool, MarkUtilized: command.MarkUtilized,
			Description: command.Description, Comments: command.Comments,
		})
	default:
		t.Fatalf("unsupported scalar test operation %q", operation)
		return nil, nil
	}
}

type prefixScalarCountingClock struct {
	now   shared.Timestamp
	calls int
}

func (clock *prefixScalarCountingClock) Now() shared.Timestamp {
	clock.calls++
	if clock.now.IsZero() {
		return shared.Timestamp{}
	}
	return shared.NewTimestamp(clock.now.Add(time.Duration(clock.calls-1) * time.Hour))
}

type prefixScalarObservedRepository struct {
	applicationipam.PrefixRepository
	loaded *domainipam.Prefix
	before domainipam.PrefixState
	writes int
}

func (repository *prefixScalarObservedRepository) GetForUpdate(ctx context.Context, id shared.ID) (*domainipam.Prefix, error) {
	loaded, err := repository.PrefixRepository.GetForUpdate(ctx, id)
	if err == nil && loaded != nil {
		repository.loaded, repository.before = loaded, loaded.State()
	}
	return loaded, err
}

func (repository *prefixScalarObservedRepository) Create(ctx context.Context, prefix *domainipam.Prefix) error {
	repository.writes++
	return repository.PrefixRepository.Create(ctx, prefix)
}

func (repository *prefixScalarObservedRepository) Update(ctx context.Context, prefix *domainipam.Prefix) error {
	repository.writes++
	return repository.PrefixRepository.Update(ctx, prefix)
}

type prefixScalarObservedRecorder struct {
	changelog.Recorder
	calls int
}

func (recorder *prefixScalarObservedRecorder) Record(ctx context.Context, change changelog.Change) error {
	recorder.calls++
	return recorder.Recorder.Record(ctx, change)
}

func newPrefixScalarObservedService(
	t *testing.T, authorizer authz.ResourceAuthorizer,
) (*gorm.DB, *applicationipam.PrefixService, *prefixScalarObservedRepository, *prefixScalarObservedRecorder, *prefixScalarCountingClock, *domainipam.Prefix) {
	t.Helper()
	db, seedService, prefixes, vrfs := newPrefixApplicationService(t, nil, &authz.AllowAll{})
	vrf := createApplicationVRF(t, vrfs, "Scalar VRF", "65000:9", true)
	command := createPrefixCommand("192.0.2.0/24", idPointer(vrf.ID()))
	command.Status = applicationipam.FieldValue("reserved")
	command.IsPool, command.MarkUtilized = applicationipam.FieldValue(true), applicationipam.FieldValue(true)
	command.Description, command.Comments = applicationipam.FieldValue("original description"), applicationipam.FieldValue("original comments")
	seed, err := seedService.CreatePrefix(t.Context(), testPrincipal(), command)
	require.NoError(t, err)
	repository := &prefixScalarObservedRepository{PrefixRepository: prefixes}
	recorder := &prefixScalarObservedRecorder{Recorder: postgreschangelog.NewRecorder(db)}
	clock := &prefixScalarCountingClock{now: shared.NewTimestamp(applicationUpdatedAt.Add(time.Hour))}
	service, err := applicationipam.NewPrefixService(
		repository, vrfs, postgrestransaction.NewUnitOfWork(db), recorder, authorizer, clock,
	)
	require.NoError(t, err)
	return db, service, repository, recorder, clock, seed
}

type prefixScalarPersistedState struct {
	prefixes []ipamrow.PrefixRow
	vrfs     []ipamrow.VRFRow
	changes  []postgreschangelog.ChangeRow
}

func prefixScalarDatabaseState(t *testing.T, db *gorm.DB) prefixScalarPersistedState {
	t.Helper()
	var state prefixScalarPersistedState
	require.NoError(t, db.Order("id").Find(&state.prefixes).Error)
	require.NoError(t, db.Order("id").Find(&state.vrfs).Error)
	require.NoError(t, db.Order("id").Find(&state.changes).Error)
	return state
}

func prefixScalarSnapshotJSON(t *testing.T, snapshot domainipam.PrefixSnapshot) string {
	t.Helper()
	var vrfID any
	if reference, present := snapshot.VRF.Get(); present {
		vrfID = reference.ID().Int64()
	}
	payload, err := json.Marshal(map[string]any{
		"prefix": snapshot.Prefix, "vrf": vrfID, "status": snapshot.Status,
		"is_pool": snapshot.IsPool, "mark_utilized": snapshot.MarkUtilized,
		"description": snapshot.Description, "comments": snapshot.Comments,
	})
	require.NoError(t, err)
	return string(payload)
}

func newPrefixApplicationService(
	t *testing.T,
	recorder changelog.Recorder,
	authorizer authz.ResourceAuthorizer,
) (*gorm.DB, *applicationipam.PrefixService, *postgresipam.PrefixRepository, *postgresipam.VRFRepository) {
	t.Helper()
	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(
		sqlite.Open(fmt.Sprintf("file:%s_%d?mode=memory&cache=shared&_foreign_keys=on", name, prefixApplicationFixtureSequence.Add(1))),
		&gorm.Config{TranslateError: true, Logger: logger.Default.LogMode(logger.Silent)},
	)
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, sqlDB.Close()) })
	models := append(identityrows.Models(), dcimrow.Models()...)
	models = append(models, ipamrow.Models()...)
	models = append(models, &postgreschangelog.ChangeRow{})
	for _, model := range models {
		require.False(t, db.Migrator().HasTable(model), "fixture tables must be absent before migration")
	}
	require.NoError(t, db.AutoMigrate(models...))
	require.NoError(t, db.Create(&identityrows.UserRow{
		ID: 17, Username: "operator", PasswordHash: "test", IsActive: true,
		Permissions: datatypes.JSON([]byte(`[]`)),
		Created:     applicationCreatedAt.Time, Updated: applicationUpdatedAt.Time,
	}).Error)
	if recorder == nil {
		recorder = postgreschangelog.NewRecorder(db)
	}
	prefixes := postgresipam.NewPrefixRepository(db)
	vrfs := postgresipam.NewVRFRepository(db)
	service, err := applicationipam.NewPrefixService(
		prefixes, vrfs, postgrestransaction.NewUnitOfWork(db), recorder,
		authorizer, fixedClock{now: applicationUpdatedAt},
	)
	require.NoError(t, err)
	return db, service, prefixes, vrfs
}

func createApplicationVRF(
	t *testing.T,
	repository *postgresipam.VRFRepository,
	name, rdValue string,
	enforceUnique bool,
) *domainipam.VRF {
	t.Helper()
	rd, err := domainipam.ParseRouteDistinguisher(rdValue)
	require.NoError(t, err)
	vrf, err := domainipam.NewVRF(domainipam.VRFValues{
		Name: name, RD: domainipam.NonNullRouteDistinguisher(rd), EnforceUnique: enforceUnique,
	}, applicationCreatedAt)
	require.NoError(t, err)
	require.NoError(t, repository.Create(t.Context(), vrf))
	return vrf
}

func createPrefixCommand(value string, vrfID *shared.ID) applicationipam.CreatePrefixCommand {
	command := applicationipam.CreatePrefixCommand{Prefix: applicationipam.FieldValue(value)}
	if vrfID != nil {
		command.VRF = applicationipam.FieldValue(vrfID.Int64())
	}
	return command
}

func idPointer(value shared.ID) *shared.ID { return &value }

func int64PointerForPrefix(value int64) *int64 { return &value }

type failingPrefixRecorder struct{}

func (failingPrefixRecorder) Record(context.Context, changelog.Change) error {
	return errors.New("forced Prefix change failure")
}
