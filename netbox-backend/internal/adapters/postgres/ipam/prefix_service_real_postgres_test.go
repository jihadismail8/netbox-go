package ipam

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"netbox-go/internal/adapters/postgres/bootstrap"
	postgreschangelog "netbox-go/internal/adapters/postgres/changelog"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	identityrow "netbox-go/internal/adapters/postgres/identity"
	ipamrow "netbox-go/internal/adapters/postgres/ipam/row"
	postgrestransaction "netbox-go/internal/adapters/postgres/transaction"
	"netbox-go/internal/application/authz"
	applicationchangelog "netbox-go/internal/application/changelog"
	applicationipam "netbox-go/internal/application/ipam"
	"netbox-go/internal/domain/identity"
	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

func TestPostgresPrefixScalarPresenceDurability(t *testing.T) {
	db, principal, vrfs := newPrefixScalarPostgres(t)
	repository := NewPrefixRepository(db)
	recorder := postgreschangelog.NewRecorder(db)
	now := time.Date(2026, time.October, 13, 8, 10, 0, 0, time.UTC)
	serviceAt := func(at time.Time) *applicationipam.PrefixService {
		return newPostgresPrefixPresenceService(t, db, at, repository, recorder)
	}
	state := loadPostgresPrefixPresenceState(t, db)
	require.Empty(t, state.prefixes)
	require.Empty(t, state.changes)

	for index, vrf := range []applicationipam.Field[int64]{
		applicationipam.OmittedField[int64](), applicationipam.NullField[int64](),
	} {
		t.Run(fmt.Sprintf("create defaults/VRF-state-%d", vrf.State()), func(t *testing.T) {
			network := []string{"192.0.2.0/24", "198.51.100.0/24"}[index]
			created, err := serviceAt(now).CreatePrefix(t.Context(), principal, applicationipam.CreatePrefixCommand{
				Prefix: applicationipam.FieldValue(network), VRF: vrf,
			})
			require.NoError(t, err)
			expected := ipamrow.PrefixRow{
				RowMetadata: ipamrow.RowMetadata{ID: created.ID().Int64(), Created: now, LastUpdated: now},
				Prefix:      network, Status: "active",
			}
			state = requirePostgresPrefixPresenceMutation(t, db, state, expected, true, principal.ID)
			requirePostgresPrefixPresenceReload(t, repository, created)
		})
	}

	created, err := serviceAt(now).CreatePrefix(t.Context(), principal, applicationipam.CreatePrefixCommand{
		Prefix: applicationipam.FieldValue("203.0.113.0/24"), VRF: applicationipam.FieldValue(vrfs[0].ID),
		Status: applicationipam.FieldValue("reserved"), IsPool: applicationipam.FieldValue(true),
		MarkUtilized: applicationipam.FieldValue(true), Description: applicationipam.FieldValue("  durable description  "),
		Comments: applicationipam.FieldValue("  " + strings.Repeat("c", 240) + "  "),
	})
	require.NoError(t, err)
	expected := ipamrow.PrefixRow{
		RowMetadata: ipamrow.RowMetadata{ID: created.ID().Int64(), Created: now, LastUpdated: now},
		Prefix:      "203.0.113.0/24", VRFID: int64Pointer(vrfs[0].ID), Status: "reserved",
		IsPool: true, MarkUtilized: true, Description: "durable description", Comments: strings.Repeat("c", 240),
	}
	state = requirePostgresPrefixPresenceMutation(t, db, state, expected, true, principal.ID)
	requirePostgresPrefixPresenceReload(t, repository, created)

	putAt := now.Add(time.Minute)
	replaced, err := serviceAt(putAt).ReplacePrefix(t.Context(), principal, applicationipam.ReplacePrefixCommand{
		ID: created.ID(), CreatePrefixCommand: applicationipam.CreatePrefixCommand{
			Prefix: applicationipam.FieldValue("2001:db8:13::/64"),
		},
	})
	require.NoError(t, err)
	expected.Prefix, expected.LastUpdated = "2001:db8:13::/64", putAt
	state = requirePostgresPrefixPresenceMutation(t, db, state, expected, false, principal.ID)
	requirePostgresPrefixPresenceReload(t, repository, replaced)

	patchAt := putAt.Add(time.Minute)
	patched, err := serviceAt(patchAt).UpdatePrefix(t.Context(), principal, applicationipam.UpdatePrefixCommand{
		ID: created.ID(), Prefix: applicationipam.FieldValue("2001:db8:14::/64"),
	})
	require.NoError(t, err)
	expected.Prefix, expected.LastUpdated = "2001:db8:14::/64", patchAt
	state = requirePostgresPrefixPresenceMutation(t, db, state, expected, false, principal.ID)
	requirePostgresPrefixPresenceReload(t, repository, patched)
	now = patchAt

	for _, operation := range []string{"replace", "update"} {
		for _, value := range []bool{false, true} {
			now = now.Add(time.Minute)
			result, err := mutatePostgresPrefixPresence(t.Context(), serviceAt(now), principal, operation, created.ID(), applicationipam.CreatePrefixCommand{
				Prefix: applicationipam.FieldValue(expected.Prefix), IsPool: applicationipam.FieldValue(value),
				MarkUtilized: applicationipam.FieldValue(value),
			})
			require.NoError(t, err)
			expected.IsPool, expected.MarkUtilized, expected.LastUpdated = value, value, now
			state = requirePostgresPrefixPresenceMutation(t, db, state, expected, false, principal.ID)
			requirePostgresPrefixPresenceReload(t, repository, result)
		}
	}

	for _, status := range []string{"container", "active", "reserved", "deprecated"} {
		now = now.Add(time.Minute)
		result, err := serviceAt(now).UpdatePrefix(t.Context(), principal, applicationipam.UpdatePrefixCommand{
			ID: created.ID(), Status: applicationipam.FieldValue(status),
		})
		require.NoError(t, err)
		expected.Status, expected.LastUpdated = status, now
		state = requirePostgresPrefixPresenceMutation(t, db, state, expected, false, principal.ID)
		requirePostgresPrefixPresenceReload(t, repository, result)
	}

	now = now.Add(time.Minute)
	cleared, err := serviceAt(now).ReplacePrefix(t.Context(), principal, applicationipam.ReplacePrefixCommand{
		ID: created.ID(), CreatePrefixCommand: applicationipam.CreatePrefixCommand{
			Prefix: applicationipam.FieldValue(expected.Prefix), VRF: applicationipam.NullField[int64](),
			IsPool: applicationipam.FieldValue(false), MarkUtilized: applicationipam.FieldValue(false),
			Description: applicationipam.FieldValue(" \t "), Comments: applicationipam.FieldValue(" \t "),
		},
	})
	require.NoError(t, err)
	expected.VRFID, expected.IsPool, expected.MarkUtilized = nil, false, false
	expected.Description, expected.Comments, expected.LastUpdated = "", "", now
	state = requirePostgresPrefixPresenceMutation(t, db, state, expected, false, principal.ID)
	requirePostgresPrefixPresenceReload(t, repository, cleared)

	now = now.Add(time.Minute)
	reassigned, err := serviceAt(now).UpdatePrefix(t.Context(), principal, applicationipam.UpdatePrefixCommand{
		ID: created.ID(), VRF: applicationipam.FieldValue(vrfs[1].ID),
		Description: applicationipam.FieldValue("  " + strings.Repeat("界", 200) + "  "),
		Comments:    applicationipam.FieldValue("  " + strings.Repeat("文", 240) + "  "),
	})
	require.NoError(t, err)
	expected.VRFID, expected.Description, expected.Comments, expected.LastUpdated = int64Pointer(vrfs[1].ID), strings.Repeat("界", 200), strings.Repeat("文", 240), now
	state = requirePostgresPrefixPresenceMutation(t, db, state, expected, false, principal.ID)
	requirePostgresPrefixPresenceReload(t, repository, reassigned)

	now = now.Add(time.Minute)
	null, err := serviceAt(now).UpdatePrefix(t.Context(), principal, applicationipam.UpdatePrefixCommand{
		ID: created.ID(), VRF: applicationipam.NullField[int64](),
	})
	require.NoError(t, err)
	expected.VRFID, expected.LastUpdated = nil, now
	state = requirePostgresPrefixPresenceMutation(t, db, state, expected, false, principal.ID)
	requirePostgresPrefixPresenceReload(t, repository, null)

	for _, operation := range []string{"create", "replace", "update"} {
		for _, test := range []struct {
			name    string
			fields  applicationipam.CreatePrefixCommand
			field   string
			reasons []string
		}{
			{"null prefix", applicationipam.CreatePrefixCommand{Prefix: applicationipam.NullField[string]()}, "prefix", []string{"null"}},
			{"blank prefix", applicationipam.CreatePrefixCommand{Prefix: applicationipam.FieldValue(" \t ")}, "prefix", []string{"blank"}},
			{"null status", applicationipam.CreatePrefixCommand{Status: applicationipam.NullField[string]()}, "status", []string{"invalid"}},
			{"blank status", applicationipam.CreatePrefixCommand{Status: applicationipam.FieldValue("")}, "status", []string{"invalid"}},
			{"padded status", applicationipam.CreatePrefixCommand{Status: applicationipam.FieldValue(" reserved ")}, "status", []string{"invalid"}},
			{"case changed status", applicationipam.CreatePrefixCommand{Status: applicationipam.FieldValue("Reserved")}, "status", []string{"invalid"}},
			{"unknown status", applicationipam.CreatePrefixCommand{Status: applicationipam.FieldValue("unknown")}, "status", []string{"invalid"}},
			{"VRF zero", applicationipam.CreatePrefixCommand{VRF: applicationipam.FieldValue(int64(0))}, "vrf", []string{"invalid"}},
			{"VRF negative", applicationipam.CreatePrefixCommand{VRF: applicationipam.FieldValue(int64(-1))}, "vrf", []string{"invalid"}},
			{"VRF absent positive ID", applicationipam.CreatePrefixCommand{VRF: applicationipam.FieldValue(int64(999999))}, "vrf", []string{"invalid"}},
			{"null is_pool", applicationipam.CreatePrefixCommand{IsPool: applicationipam.NullField[bool]()}, "is_pool", []string{"null"}},
			{"null mark_utilized", applicationipam.CreatePrefixCommand{MarkUtilized: applicationipam.NullField[bool]()}, "mark_utilized", []string{"null"}},
			{"null description", applicationipam.CreatePrefixCommand{Description: applicationipam.NullField[string]()}, "description", []string{"null"}},
			{"null comments", applicationipam.CreatePrefixCommand{Comments: applicationipam.NullField[string]()}, "comments", []string{"null"}},
			{"description NUL", applicationipam.CreatePrefixCommand{Description: applicationipam.FieldValue("bad\x00description")}, "description", []string{"null_characters_not_allowed"}},
			{"comments NUL", applicationipam.CreatePrefixCommand{Comments: applicationipam.FieldValue("bad\x00comments")}, "comments", []string{"null_characters_not_allowed"}},
			{"description length and NUL", applicationipam.CreatePrefixCommand{Description: applicationipam.FieldValue(strings.Repeat("界", 200) + "\x00")}, "description", []string{"max_length", "null_characters_not_allowed"}},
		} {
			t.Run("validation non-effects/"+operation+"/"+test.name, func(t *testing.T) {
				before := loadPostgresPrefixPresenceState(t, db)
				fields := test.fields
				if operation != "update" && fields.Prefix.State() == applicationipam.FieldOmitted {
					fields.Prefix = applicationipam.FieldValue("10.250.0.0/24")
				}
				result, err := mutatePostgresPrefixPresence(t.Context(), serviceAt(now.Add(time.Minute)), principal, operation, created.ID(), fields)
				require.Nil(t, result)
				require.True(t, shared.HasReason(err, shared.ErrorReasonValidation), err)
				var reasons []string
				for _, violation := range shared.ViolationsOf(err) {
					require.Equal(t, test.field, violation.Field)
					reasons = append(reasons, violation.Reason)
				}
				require.Equal(t, test.reasons, reasons)
				require.Equal(t, before, loadPostgresPrefixPresenceState(t, db))
			})
		}
		if operation != "update" {
			t.Run("required omission/"+operation, func(t *testing.T) {
				before := loadPostgresPrefixPresenceState(t, db)
				result, err := mutatePostgresPrefixPresence(t.Context(), serviceAt(now.Add(time.Minute)), principal, operation, created.ID(), applicationipam.CreatePrefixCommand{})
				require.Nil(t, result)
				require.Equal(t, []shared.FieldViolation{{Field: "prefix", Reason: "required", Description: "This field is required."}}, shared.ViolationsOf(err))
				require.Equal(t, before, loadPostgresPrefixPresenceState(t, db))
			})
		}
		t.Run("mixed null diagnostics/"+operation, func(t *testing.T) {
			before := loadPostgresPrefixPresenceState(t, db)
			result, err := mutatePostgresPrefixPresence(t.Context(), serviceAt(now.Add(time.Minute)), principal, operation, created.ID(), applicationipam.CreatePrefixCommand{
				Prefix: applicationipam.NullField[string](), VRF: applicationipam.FieldValue(int64(0)),
				Status: applicationipam.FieldValue(" unknown "), IsPool: applicationipam.NullField[bool](),
				MarkUtilized: applicationipam.NullField[bool](),
				Description:  applicationipam.FieldValue(strings.Repeat("界", 200) + "\x00"),
				Comments:     applicationipam.FieldValue("c\x00"),
			})
			require.Nil(t, result)
			require.Equal(t, []shared.FieldViolation{
				{Field: "prefix", Reason: "null", Description: "This field may not be null."},
				{Field: "vrf", Reason: "invalid", Description: "Related object not found using the provided numeric ID: 0"},
				{Field: "status", Reason: "invalid", Description: " unknown  is not a valid choice."},
				{Field: "is_pool", Reason: "null", Description: "This field may not be null."},
				{Field: "mark_utilized", Reason: "null", Description: "This field may not be null."},
				{Field: "description", Reason: "max_length", Description: "Ensure this field has no more than 200 characters."},
				{Field: "description", Reason: "null_characters_not_allowed", Description: "Null characters are not allowed."},
				{Field: "comments", Reason: "null_characters_not_allowed", Description: "Null characters are not allowed."},
			}, shared.ViolationsOf(err), "presence must suppress only the fallback blank and never fabricate an update_mask error")
			require.Equal(t, before, loadPostgresPrefixPresenceState(t, db))
		})
	}

	for _, failure := range []string{"repository", "recorder"} {
		for _, operation := range []string{"create", "replace", "update"} {
			t.Run("delegate then error rollback/"+failure+"/"+operation, func(t *testing.T) {
				before := loadPostgresPrefixPresenceState(t, db)
				sentinel := errors.New("forced Prefix failure after actual delegate")
				observed := 0
				at := now.Add(2 * time.Minute)
				observe := func(ctx context.Context, item *domainipam.Prefix) {
					observed++
					tx, bound := postgrestransaction.FromContext(ctx)
					require.True(t, bound)
					inside := loadPostgresPrefixPresenceState(t, tx)
					current := postgresPrefixPresenceRow(t, inside, item.ID().Int64())
					require.Equal(t, "2001:db8:1300::/64", current.Prefix)
					require.Equal(t, int64Pointer(vrfs[0].ID), current.VRFID)
					require.Equal(t, "reserved", current.Status)
					require.True(t, current.IsPool)
					require.True(t, current.MarkUtilized)
					require.Equal(t, "rollback description", current.Description)
					require.Equal(t, "rollback comments", current.Comments)
					require.True(t, current.LastUpdated.Equal(at))
					if failure == "recorder" {
						requirePostgresPrefixPresenceChange(t, inside, before, current, operation == "create", principal.ID)
					} else {
						require.Equal(t, before.changes, inside.changes)
					}
				}
				var activeRepository applicationipam.PrefixRepository = repository
				var activeRecorder applicationchangelog.Recorder = recorder
				if failure == "repository" {
					activeRepository = &prefixPresenceFailingRepository{PrefixRepository: repository, observe: observe, err: sentinel}
				} else {
					activeRecorder = &prefixPresenceFailingRecorder{delegate: recorder, repository: repository, observe: observe, err: sentinel}
				}
				service := newPostgresPrefixPresenceService(t, db, at, activeRepository, activeRecorder)
				result, err := mutatePostgresPrefixPresence(t.Context(), service, principal, operation, created.ID(), applicationipam.CreatePrefixCommand{
					Prefix: applicationipam.FieldValue("2001:db8:1300::/64"), VRF: applicationipam.FieldValue(vrfs[0].ID),
					Status: applicationipam.FieldValue("reserved"), IsPool: applicationipam.FieldValue(true),
					MarkUtilized: applicationipam.FieldValue(true), Description: applicationipam.FieldValue("rollback description"),
					Comments: applicationipam.FieldValue("rollback comments"),
				})
				require.ErrorIs(t, err, sentinel)
				require.Nil(t, result)
				require.Equal(t, 1, observed, "observe actual delegated durable writes before raising the sentinel")
				require.Equal(t, before, loadPostgresPrefixPresenceState(t, db), "rows, timestamps, relations and object changes must all roll back")
			})
		}
	}
}

func TestPostgresPrefixNetworkCanonicalDurabilityAndRollback(t *testing.T) {
	cases := []struct {
		name, input, canonical, description string
		family                              uint32
		bits                                int
	}{
		{"canonical IPv4", "198.51.100.0/24", "198.51.100.0/24", "", 4, 24},
		{"bare IPv4 host", "203.0.113.17", "203.0.113.17/32", "", 4, 32},
		{"canonical IPv6", "2001:db8:10::/64", "2001:db8:10::/64", "", 6, 64},
		{"expanded IPv6 host", "2001:0DB8:0020:0000:0000:0000:0000:0017", "2001:db8:20::17/128", "", 6, 128},
		{"expanded IPv6 network", "2001:0DB8:0030:0000:0000:0000:0000:0000/64", "2001:db8:30::/64", "", 6, 64},
		{"IPv4 host bits", "203.0.113.17/24", "", "203.0.113.17/24 is not a valid prefix. Did you mean 203.0.113.0/24?", 0, 0},
		{"IPv6 host bits", "2001:db8:40::17/64", "", "2001:db8:40::17/64 is not a valid prefix. Did you mean 2001:db8:40::/64?", 0, 0},
		{"IPv4 slash zero", "0.0.0.0/0", "", "Cannot create prefix with /0 mask.", 0, 0},
		{"IPv6 slash zero", "::/0", "", "Cannot create prefix with /0 mask.", 0, 0},
	}
	seedAt := time.Date(2026, time.October, 13, 18, 0, 0, 0, time.UTC)
	at := seedAt.Add(time.Minute)
	for _, test := range cases {
		for _, operation := range []string{"create", "replace", "update"} {
			t.Run("matrix/"+test.name+"/"+operation, func(t *testing.T) {
				db, principal, vrfs := newPrefixScalarPostgres(t)
				repository, recorder := NewPrefixRepository(db), postgreschangelog.NewRecorder(db)
				id, before, expected := preparePostgresPrefixNetworkMutation(t, db, principal, repository, recorder, vrfs[0].ID, operation, seedAt)
				fields := applicationipam.CreatePrefixCommand{Prefix: applicationipam.FieldValue(test.input)}
				if operation == "create" {
					fields = prefixNetworkPostgresFields(test.input, vrfs[0].ID)
				}
				service := newPostgresPrefixPresenceService(t, db, at, repository, recorder)
				result, err := mutatePostgresPrefixPresence(t.Context(), service, principal, operation, id, fields)
				if test.description != "" {
					require.Nil(t, result)
					require.True(t, shared.HasReason(err, shared.ErrorReasonValidation), err)
					require.Equal(t, []shared.FieldViolation{{Field: "prefix", Reason: "invalid", Description: test.description}}, shared.ViolationsOf(err))
					require.Equal(t, before, loadPostgresPrefixPresenceState(t, db))
					return
				}
				require.NoError(t, err)
				require.NotNil(t, result)
				require.True(t, result.ID().IsValid())
				if operation == "create" {
					expected.ID, expected.Created = result.ID().Int64(), at
				} else {
					require.Equal(t, id, result.ID())
				}
				expected.Prefix, expected.LastUpdated = test.canonical, at
				requirePostgresPrefixPresenceMutation(t, db, before, expected, operation == "create", principal.ID)
				require.Equal(t, test.canonical, result.Network().String())
				require.Equal(t, test.canonical, result.Display())
				require.Equal(t, test.canonical, result.State().Prefix)
				require.Equal(t, test.canonical, result.Snapshot().Prefix)
				require.Equal(t, test.family, result.Family())
				require.Equal(t, test.bits, result.Network().Bits())
				requirePostgresPrefixPresenceReload(t, repository, result)
			})
		}
	}
	for _, failure := range []string{"repository", "recorder"} {
		for _, operation := range []string{"create", "replace", "update"} {
			t.Run("late error/"+failure+"/"+operation, func(t *testing.T) {
				db, principal, vrfs := newPrefixScalarPostgres(t)
				repository, recorder := NewPrefixRepository(db), postgreschangelog.NewRecorder(db)
				id, before, expected := preparePostgresPrefixNetworkMutation(t, db, principal, repository, recorder, vrfs[0].ID, operation, seedAt)
				sentinel := errors.New("forced network proof error after actual Prefix delegate")
				observed := 0
				observe := func(ctx context.Context, item *domainipam.Prefix) {
					observed++
					tx, bound := postgrestransaction.FromContext(ctx)
					require.True(t, bound)
					inside := loadPostgresPrefixPresenceState(t, tx)
					if operation == "create" {
						expected.ID, expected.Created = item.ID().Int64(), at
					} else {
						require.Equal(t, id, item.ID())
					}
					expected.Prefix, expected.LastUpdated = "2001:db8:50::/64", at
					actual := postgresPrefixPresenceRow(t, inside, expected.ID)
					require.True(t, actual.Created.Equal(expected.Created))
					require.True(t, actual.LastUpdated.Equal(expected.LastUpdated))
					actual.Created, actual.LastUpdated = expected.Created, expected.LastUpdated
					require.Equal(t, expected, actual)
					require.Equal(t, before.vrfs, inside.vrfs)
					count := len(before.prefixes)
					if operation == "create" {
						count++
					}
					require.Len(t, inside.prefixes, count)
					for _, prior := range before.prefixes {
						if prior.ID != expected.ID {
							require.Equal(t, prior, postgresPrefixPresenceRow(t, inside, prior.ID))
						}
					}
					require.Equal(t, "2001:db8:50::/64", item.Network().String())
					require.Equal(t, "2001:db8:50::/64", item.State().Prefix)
					require.Equal(t, "2001:db8:50::/64", item.Snapshot().Prefix)
					require.Equal(t, uint32(6), item.Family())
					require.Equal(t, 64, item.Network().Bits())
					if failure == "recorder" {
						requirePostgresPrefixPresenceChange(t, inside, before, expected, operation == "create", principal.ID)
					} else {
						require.Equal(t, before.changes, inside.changes)
					}
				}
				var activeRepository applicationipam.PrefixRepository = repository
				var activeRecorder applicationchangelog.Recorder = recorder
				if failure == "repository" {
					activeRepository = &prefixPresenceFailingRepository{PrefixRepository: repository, observe: observe, err: sentinel}
				} else {
					activeRecorder = &prefixPresenceFailingRecorder{delegate: recorder, repository: repository, observe: observe, err: sentinel}
				}
				fields := applicationipam.CreatePrefixCommand{Prefix: applicationipam.FieldValue("2001:0DB8:0050:0000:0000:0000:0000:0000/64")}
				if operation == "create" {
					fields = prefixNetworkPostgresFields("2001:0DB8:0050:0000:0000:0000:0000:0000/64", vrfs[0].ID)
				}
				service := newPostgresPrefixPresenceService(t, db, at, activeRepository, activeRecorder)
				result, err := mutatePostgresPrefixPresence(t.Context(), service, principal, operation, id, fields)
				require.ErrorIs(t, err, sentinel)
				require.Nil(t, result)
				require.Equal(t, 1, observed)
				require.Equal(t, before, loadPostgresPrefixPresenceState(t, db), "delegated rows, relations, times and canonical change snapshots must roll back")
			})
		}
	}
}

func prefixNetworkPostgresFields(network string, vrfID int64) applicationipam.CreatePrefixCommand {
	return applicationipam.CreatePrefixCommand{
		Prefix: applicationipam.FieldValue(network), VRF: applicationipam.FieldValue(vrfID),
		Status: applicationipam.FieldValue("reserved"), IsPool: applicationipam.FieldValue(true), MarkUtilized: applicationipam.FieldValue(true),
		Description: applicationipam.FieldValue("network proof description"), Comments: applicationipam.FieldValue("network proof comments"),
	}
}

func preparePostgresPrefixNetworkMutation(t *testing.T, db *gorm.DB, principal identity.Principal, repository *PrefixRepository, recorder *postgreschangelog.Recorder, vrfID int64, operation string, seedAt time.Time) (shared.ID, postgresPrefixPresenceState, ipamrow.PrefixRow) {
	t.Helper()
	before := loadPostgresPrefixPresenceState(t, db)
	require.Empty(t, before.prefixes)
	require.Empty(t, before.changes)
	expected := ipamrow.PrefixRow{
		Prefix: "192.0.2.0/24", VRFID: int64Pointer(vrfID), Status: "reserved", IsPool: true, MarkUtilized: true,
		Description: "network proof description", Comments: "network proof comments",
	}
	if operation == "create" {
		return 0, before, expected
	}
	service := newPostgresPrefixPresenceService(t, db, seedAt, repository, recorder)
	seed, err := service.CreatePrefix(t.Context(), principal, prefixNetworkPostgresFields("192.0.2.0/24", vrfID))
	require.NoError(t, err)
	require.NotNil(t, seed)
	expected.ID, expected.Created, expected.LastUpdated = seed.ID().Int64(), seedAt, seedAt
	before = requirePostgresPrefixPresenceMutation(t, db, before, expected, true, principal.ID)
	return seed.ID(), before, expected
}

func mutatePostgresPrefixPresence(ctx context.Context, service *applicationipam.PrefixService, principal identity.Principal, operation string, id shared.ID, fields applicationipam.CreatePrefixCommand) (*domainipam.Prefix, error) {
	switch operation {
	case "create":
		return service.CreatePrefix(ctx, principal, fields)
	case "replace":
		return service.ReplacePrefix(ctx, principal, applicationipam.ReplacePrefixCommand{ID: id, CreatePrefixCommand: fields})
	case "update":
		return service.UpdatePrefix(ctx, principal, applicationipam.UpdatePrefixCommand{
			ID: id, Prefix: fields.Prefix, VRF: fields.VRF, Status: fields.Status, IsPool: fields.IsPool,
			MarkUtilized: fields.MarkUtilized, Description: fields.Description, Comments: fields.Comments,
		})
	default:
		panic("unsupported Prefix scalar test mutation")
	}
}

type postgresPrefixPresenceState struct {
	prefixes []ipamrow.PrefixRow
	vrfs     []ipamrow.VRFRow
	changes  []postgreschangelog.ChangeRow
}

func loadPostgresPrefixPresenceState(t *testing.T, db *gorm.DB) postgresPrefixPresenceState {
	t.Helper()
	var state postgresPrefixPresenceState
	require.NoError(t, db.Order("id").Find(&state.prefixes).Error)
	require.NoError(t, db.Order("id").Find(&state.vrfs).Error)
	require.NoError(t, db.Order("id").Find(&state.changes).Error)
	var addresses int64
	require.NoError(t, db.Model(&ipamrow.IPAddressRow{}).Count(&addresses).Error)
	require.Zero(t, addresses, "empty reader-support tables earn no downstream IPAddress claim")
	return state
}

func postgresPrefixPresenceRow(t *testing.T, state postgresPrefixPresenceState, id int64) ipamrow.PrefixRow {
	t.Helper()
	for _, row := range state.prefixes {
		if row.ID == id {
			return row
		}
	}
	t.Fatalf("Prefix %d is absent", id)
	return ipamrow.PrefixRow{}
}

func requirePostgresPrefixPresenceMutation(t *testing.T, db *gorm.DB, before postgresPrefixPresenceState, expected ipamrow.PrefixRow, create bool, actorID int64) postgresPrefixPresenceState {
	t.Helper()
	after := loadPostgresPrefixPresenceState(t, db)
	count := len(before.prefixes)
	if create {
		count++
	}
	require.Len(t, after.prefixes, count)
	require.Equal(t, before.vrfs, after.vrfs)
	for _, row := range before.prefixes {
		if row.ID != expected.ID {
			require.Equal(t, row, postgresPrefixPresenceRow(t, after, row.ID))
		}
	}
	actual := postgresPrefixPresenceRow(t, after, expected.ID)
	require.True(t, actual.Created.Equal(expected.Created))
	require.True(t, actual.LastUpdated.Equal(expected.LastUpdated))
	actual.Created, actual.LastUpdated = expected.Created, expected.LastUpdated
	require.Equal(t, expected, actual)
	requirePostgresPrefixPresenceChange(t, after, before, expected, create, actorID)
	return after
}

// Changelog snapshots own seven writable fields, not raw-row identity or timestamps.
type prefixPresencePGSnapshot struct {
	Prefix       string `json:"prefix"`
	VRFID        *int64 `json:"vrf"`
	Status       string `json:"status"`
	IsPool       bool   `json:"is_pool"`
	MarkUtilized bool   `json:"mark_utilized"`
	Description  string `json:"description"`
	Comments     string `json:"comments"`
}

func requirePostgresPrefixPresencePayload(t *testing.T, data []byte, row ipamrow.PrefixRow) {
	t.Helper()
	var payload map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &payload))
	require.Len(t, payload, 7)
	for _, key := range []string{"prefix", "vrf", "status", "is_pool", "mark_utilized", "description", "comments"} {
		require.Contains(t, payload, key)
	}
	encoded, err := json.Marshal(prefixPresencePGSnapshot{
		Prefix: row.Prefix, VRFID: row.VRFID, Status: row.Status, IsPool: row.IsPool,
		MarkUtilized: row.MarkUtilized, Description: row.Description, Comments: row.Comments,
	})
	require.NoError(t, err)
	require.JSONEq(t, string(encoded), string(data))
}

func requirePostgresPrefixPresenceChange(t *testing.T, after, before postgresPrefixPresenceState, current ipamrow.PrefixRow, create bool, actorID int64) {
	t.Helper()
	require.Len(t, after.changes, len(before.changes)+1)
	require.Equal(t, before.changes, after.changes[:len(before.changes)])
	change := after.changes[len(after.changes)-1]
	require.Positive(t, change.ID)
	require.Equal(t, actorID, change.ActorID)
	require.Equal(t, domainipam.PrefixObjectType, change.Kind)
	require.Equal(t, current.ID, change.ObjectID)
	require.True(t, change.OccurredAt.Equal(current.LastUpdated))
	if create {
		require.Equal(t, "create", change.Action)
		require.Empty(t, change.BeforeData)
	} else {
		require.Equal(t, "update", change.Action)
		prior := postgresPrefixPresenceRow(t, before, current.ID)
		require.True(t, prior.Created.Equal(current.Created))
		requirePostgresPrefixPresencePayload(t, change.BeforeData, prior)
	}
	requirePostgresPrefixPresencePayload(t, change.AfterData, current)
}

func requirePostgresPrefixPresenceReload(t *testing.T, repository *PrefixRepository, expected *domainipam.Prefix) {
	t.Helper()
	actual, err := repository.Get(t.Context(), expected.ID())
	require.NoError(t, err)
	actualState, expectedState := actual.State(), expected.State()
	require.True(t, actualState.Created.Equal(expectedState.Created.Time))
	require.True(t, actualState.LastUpdated.Equal(expectedState.LastUpdated.Time))
	actualState.Created, actualState.LastUpdated = expectedState.Created, expectedState.LastUpdated
	require.Equal(t, expectedState, actualState)
	require.Equal(t, expected.Snapshot(), actual.Snapshot())
}

type prefixPresenceFailingRepository struct {
	applicationipam.PrefixRepository
	observe func(context.Context, *domainipam.Prefix)
	err     error
}

func (repository *prefixPresenceFailingRepository) Create(ctx context.Context, item *domainipam.Prefix) error {
	if err := repository.PrefixRepository.Create(ctx, item); err != nil {
		return err
	}
	repository.observe(ctx, item)
	return repository.err
}

func (repository *prefixPresenceFailingRepository) Update(ctx context.Context, item *domainipam.Prefix) error {
	if err := repository.PrefixRepository.Update(ctx, item); err != nil {
		return err
	}
	repository.observe(ctx, item)
	return repository.err
}

type prefixPresenceFailingRecorder struct {
	delegate   applicationchangelog.Recorder
	repository *PrefixRepository
	observe    func(context.Context, *domainipam.Prefix)
	err        error
}

func (recorder *prefixPresenceFailingRecorder) Record(ctx context.Context, change applicationchangelog.Change) error {
	if err := recorder.delegate.Record(ctx, change); err != nil {
		return err
	}
	item, err := recorder.repository.Get(ctx, change.ObjectID)
	if err != nil {
		return err
	}
	recorder.observe(ctx, item)
	return recorder.err
}

func newPostgresPrefixPresenceService(t *testing.T, db *gorm.DB, now time.Time, repository applicationipam.PrefixRepository, recorder applicationchangelog.Recorder) *applicationipam.PrefixService {
	t.Helper()
	service, err := applicationipam.NewPrefixService(
		repository, NewVRFRepository(db), postgrestransaction.NewUnitOfWork(db), recorder,
		authz.AllowAll{}, ipamPostgresConcurrencyClock{now: shared.NewTimestamp(now)},
	)
	require.NoError(t, err)
	return service
}

func newPrefixScalarPostgres(t *testing.T) (*gorm.DB, identity.Principal, []ipamrow.VRFRow) {
	t.Helper()
	dsn := os.Getenv("NETBOX_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("NETBOX_TEST_POSTGRES_DSN is not set")
	}
	base, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	baseSQL, err := base.DB()
	require.NoError(t, err)
	schema := fmt.Sprintf("ipam_prefix_scalar_presence_%d", time.Now().UnixNano())
	var existing int64
	require.NoError(t, base.Raw("SELECT count(*) FROM pg_namespace WHERE nspname = ?", schema).Scan(&existing).Error)
	require.Zero(t, existing)
	require.NoError(t, base.Exec(`CREATE SCHEMA "`+schema+`"`).Error)
	t.Cleanup(func() {
		if err := base.Exec(`DROP SCHEMA "` + schema + `" CASCADE`).Error; err != nil {
			t.Error("could not drop the owned Prefix scalar-presence schema")
		}
		var remaining int64
		if err := base.Raw("SELECT count(*) FROM pg_namespace WHERE nspname = ?", schema).Scan(&remaining).Error; err != nil || remaining != 0 {
			t.Error("the owned Prefix scalar-presence schema remained after cleanup")
		}
		if err := baseSQL.Close(); err != nil {
			t.Error("could not close the PostgreSQL cleanup connection")
		}
	})
	db, err := gorm.Open(postgres.Open(ipamPostgresDSNWithSearchPath(t, dsn, schema)), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Error("could not close the scoped PostgreSQL Prefix connection")
		}
	})
	entries := []bootstrap.Entry{{Name: "go_identity_users", Model: &identityrow.UserRow{}}}
	for _, descriptor := range dcimrow.Descriptors() {
		entries = append(entries, bootstrap.Entry{Name: descriptor.Name, Model: descriptor.Model, Dependencies: descriptor.Dependencies})
	}
	for _, descriptor := range ipamrow.Descriptors() {
		entries = append(entries, bootstrap.Entry{Name: descriptor.Name, Model: descriptor.Model, Dependencies: descriptor.Dependencies})
	}
	entries = append(entries, bootstrap.Entry{Name: "go_object_changes", Model: &postgreschangelog.ChangeRow{}, Dependencies: []string{"go_identity_users"}})
	for _, entry := range entries {
		require.False(t, db.Migrator().HasTable(entry.Model))
	}
	registry, err := bootstrap.NewRegistry(entries...)
	require.NoError(t, err)
	result, err := bootstrap.Run(t.Context(), db, registry)
	require.NoError(t, err)
	require.Len(t, result.Created, len(entries))
	require.Empty(t, result.Existing)
	now := time.Date(2026, time.October, 13, 8, 10, 0, 0, time.UTC)
	require.NoError(t, db.Create(&identityrow.UserRow{
		ID: 1, Username: "prefix-presence-test", PasswordHash: "not-used-by-tests",
		IsActive: true, IsStaff: true, IsSuperuser: true, Permissions: []byte("[]"), Created: now, Updated: now,
	}).Error)
	vrfs := []ipamrow.VRFRow{
		{RowMetadata: ipamrow.RowMetadata{Created: now, LastUpdated: now}, Name: "Prefix VRF one", RD: stringPointer("64512:1301"), EnforceUnique: true},
		{RowMetadata: ipamrow.RowMetadata{Created: now, LastUpdated: now}, Name: "Prefix VRF two", RD: stringPointer("64512:1302"), EnforceUnique: true},
	}
	require.NoError(t, db.Create(&vrfs).Error)
	return db, identity.Principal{ID: 1, Username: "prefix-presence-test", IsSuperuser: true}, vrfs
}
