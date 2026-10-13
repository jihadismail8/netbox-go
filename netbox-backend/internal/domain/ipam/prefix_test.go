package ipam_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

var prefixDomainTime = shared.NewTimestamp(time.Date(2026, time.July, 22, 14, 0, 0, 0, time.UTC))

func TestPrefixRejectsHostBitsWithCanonicalSuggestionAndSlashZero(t *testing.T) {
	t.Parallel()

	_, err := domainipam.NewPrefix(validPrefixValues("192.0.2.7/24"), prefixDomainTime)
	require.Error(t, err)
	violations := shared.ViolationsOf(err)
	require.Len(t, violations, 1)
	assert.Equal(t, "prefix", violations[0].Field)
	assert.Equal(t, "192.0.2.7/24 is not a valid prefix. Did you mean 192.0.2.0/24?", violations[0].Description)

	for _, value := range []string{"0.0.0.0/0", "::/0"} {
		_, err = domainipam.NewPrefix(validPrefixValues(value), prefixDomainTime)
		require.Error(t, err)
		assert.Equal(t, "Cannot create prefix with /0 mask.", shared.ViolationsOf(err)[0].Description)
	}
}

func TestPrefixAcceptsBareHostsCanonicalNetworksAndDefaults(t *testing.T) {
	t.Parallel()

	prefix, err := domainipam.NewPrefix(validPrefixValues(" 2001:db8::1 "), prefixDomainTime)
	require.NoError(t, err)
	assert.Equal(t, "2001:db8::1/128", prefix.Network().String())
	assert.Equal(t, uint32(6), prefix.Family())
	assert.Equal(t, domainipam.PrefixStatusActive, prefix.Status())
	assert.False(t, prefix.IsPool())
	assert.False(t, prefix.MarkUtilized())
	assert.True(t, prefix.VRF().IsNull())
}

func TestPrefixValidatesChoiceReferenceAndDescription(t *testing.T) {
	t.Parallel()

	values := validPrefixValues("192.0.2.0/24")
	values.Status = "unknown"
	values.Description = strings.Repeat("x", domainipam.PrefixDescriptionMaxLength+1)
	values.VRF = domainipam.NonNullVRFReference(domainipam.VRFReference{})
	_, err := domainipam.NewPrefix(values, prefixDomainTime)
	require.Error(t, err)
	assert.Len(t, shared.ViolationsOf(err), 3)
}

func TestPrefixPatchPreservesNullableVRFAndComputedProjection(t *testing.T) {
	t.Parallel()

	reference := prefixVRFReference(t, 9, "Tenant", "65000:9", true)
	prefix, err := domainipam.RestorePrefix(domainipam.PrefixState{
		ID: 3, Prefix: "10.0.0.0/8", VRF: domainipam.NonNullVRFReference(reference),
		Status:  domainipam.PrefixStatusContainer.String(),
		Created: prefixDomainTime, LastUpdated: prefixDomainTime, Children: 7, Depth: 2,
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(7), prefix.Children())
	assert.Equal(t, uint32(2), prefix.Depth())

	nullVRF := domainipam.NullVRFReference()
	updated := shared.NewTimestamp(prefixDomainTime.Add(time.Hour))
	require.NoError(t, prefix.ApplyPatch(domainipam.PrefixPatch{
		VRF: &nullVRF, Description: stringPointer(" moved "),
	}, updated))
	assert.True(t, prefix.VRF().IsNull())
	assert.Equal(t, "moved", prefix.Description())
	assert.Equal(t, updated, prefix.LastUpdated())
	assert.Equal(t, shared.ID(9), reference.ID())
}

func TestPrefixNetworkContainmentAndOrdering(t *testing.T) {
	t.Parallel()

	outer := requiredPrefixNetwork(t, "10.0.0.0/8")
	inner := requiredPrefixNetwork(t, "10.1.0.0/16")
	equal := requiredPrefixNetwork(t, "10.0.0.0/8")
	v6 := requiredPrefixNetwork(t, "2001:db8::/32")
	assert.True(t, outer.Contains(inner, false))
	assert.False(t, outer.Contains(equal, false))
	assert.True(t, outer.Contains(equal, true))
	assert.False(t, outer.Contains(v6, true))
	assert.Less(t, outer.Compare(inner), 0)
	assert.Less(t, inner.Compare(v6), 0)

	filtered, explicit, err := domainipam.ParsePrefixFilter("10.1.2.3/16")
	require.NoError(t, err)
	assert.True(t, explicit)
	assert.Equal(t, "10.1.0.0/16", filtered.String())
}

func TestPrefixScalarNormalizationContract(t *testing.T) {
	t.Parallel()

	for _, network := range []string{"192.0.2.0/24", "2001:db8::/64"} {
		for _, status := range []string{"container", "active", "reserved", "deprecated"} {
			values := validPrefixValues(network)
			values.Status = status
			values.IsPool, values.MarkUtilized = true, true
			values.Description = " \t" + strings.Repeat("界", 200) + "\n"
			values.Comments = " \t" + strings.Repeat("界", 1000) + "\n"
			prefix, err := domainipam.NewPrefix(values, prefixDomainTime)
			require.NoError(t, err)
			assert.Equal(t, network, prefix.Network().String())
			assert.Equal(t, status, prefix.Status().String())
			assert.True(t, prefix.IsPool())
			assert.True(t, prefix.MarkUtilized())
			assert.Equal(t, strings.Repeat("界", 200), prefix.Description())
			assert.Equal(t, strings.Repeat("界", 1000), prefix.Comments())
		}
	}

	cases := []struct {
		name   string
		change func(*domainipam.PrefixValues)
		want   []shared.FieldViolation
	}{
		{"empty Prefix", func(v *domainipam.PrefixValues) { v.Prefix = "" },
			[]shared.FieldViolation{{Field: "prefix", Reason: "blank", Description: "This field may not be blank."}}},
		{"whitespace Prefix", func(v *domainipam.PrefixValues) { v.Prefix = " \t\n" },
			[]shared.FieldViolation{{Field: "prefix", Reason: "blank", Description: "This field may not be blank."}}},
		{"empty status", func(v *domainipam.PrefixValues) { v.Status = "" },
			[]shared.FieldViolation{{Field: "status", Reason: "invalid", Description: "This field may not be blank."}}},
		{"padded status", func(v *domainipam.PrefixValues) { v.Status = " active " },
			[]shared.FieldViolation{{Field: "status", Reason: "invalid", Description: " active  is not a valid choice."}}},
		{"uppercase status", func(v *domainipam.PrefixValues) { v.Status = "ACTIVE" },
			[]shared.FieldViolation{{Field: "status", Reason: "invalid", Description: "ACTIVE is not a valid choice."}}},
		{"whitespace status", func(v *domainipam.PrefixValues) { v.Status = " " },
			[]shared.FieldViolation{{Field: "status", Reason: "invalid", Description: "  is not a valid choice."}}},
		{"unknown status", func(v *domainipam.PrefixValues) { v.Status = "unknown" },
			[]shared.FieldViolation{{Field: "status", Reason: "invalid", Description: "unknown is not a valid choice."}}},
		{"Unicode description length", func(v *domainipam.PrefixValues) { v.Description = strings.Repeat("界", 201) },
			[]shared.FieldViolation{{Field: "description", Reason: "max_length", Description: "Ensure this field has no more than 200 characters."}}},
		{"description NUL", func(v *domainipam.PrefixValues) { v.Description = "note\x00" },
			[]shared.FieldViolation{{Field: "description", Reason: "null_characters_not_allowed", Description: "Null characters are not allowed."}}},
		{"description length and NUL", func(v *domainipam.PrefixValues) { v.Description = strings.Repeat("界", 200) + "\x00" },
			[]shared.FieldViolation{
				{Field: "description", Reason: "max_length", Description: "Ensure this field has no more than 200 characters."},
				{Field: "description", Reason: "null_characters_not_allowed", Description: "Null characters are not allowed."},
			}},
		{"comments NUL", func(v *domainipam.PrefixValues) { v.Comments = "note\x00" },
			[]shared.FieldViolation{{Field: "comments", Reason: "null_characters_not_allowed", Description: "Null characters are not allowed."}}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			values := validPrefixValues("198.51.100.0/24")
			test.change(&values)
			prefix, err := domainipam.NewPrefix(values, prefixDomainTime)
			assert.Nil(t, prefix)
			assert.Equal(t, test.want, shared.ViolationsOf(err))

			loaded, err := domainipam.NewPrefix(validPrefixValues("192.0.2.0/24"), prefixDomainTime)
			require.NoError(t, err)
			require.NoError(t, loaded.AssignID(3))
			before := loaded.State()
			err = loaded.Replace(values, shared.NewTimestamp(prefixDomainTime.Add(time.Hour)))
			assert.Equal(t, test.want, shared.ViolationsOf(err))
			assert.Equal(t, before, loaded.State(), "rejected scalar replacement must not mutate the aggregate")
		})
	}

	loaded, err := domainipam.NewPrefix(validPrefixValues("192.0.2.0/24"), prefixDomainTime)
	require.NoError(t, err)
	before := loaded.State()
	invalid := "note\x00"
	err = loaded.ApplyPatch(domainipam.PrefixPatch{Comments: &invalid}, prefixDomainTime)
	assert.Equal(t, []shared.FieldViolation{{
		Field: "comments", Reason: "null_characters_not_allowed", Description: "Null characters are not allowed.",
	}}, shared.ViolationsOf(err))
	assert.Equal(t, before, loaded.State())
	blank := " \t\n"
	require.NoError(t, loaded.ApplyPatch(domainipam.PrefixPatch{Description: &blank, Comments: &blank}, prefixDomainTime))
	assert.Empty(t, loaded.Description())
	assert.Empty(t, loaded.Comments())
	before = loaded.State()
	paddedStatus := " active "
	err = loaded.ValidatePatch(domainipam.PrefixPatch{Status: &paddedStatus})
	assert.Equal(t, []shared.FieldViolation{{
		Field: "status", Reason: "invalid", Description: " active  is not a valid choice.",
	}}, shared.ViolationsOf(err))
	assert.Equal(t, before, loaded.State())
	require.NoError(t, loaded.ValidatePatch(domainipam.PrefixPatch{}))
	err = loaded.ApplyPatch(domainipam.PrefixPatch{}, prefixDomainTime)
	assert.Equal(t, []shared.FieldViolation{{
		Field: "update_mask", Reason: "required", Description: "At least one writable field must be supplied.",
	}}, shared.ViolationsOf(err))
	assert.Equal(t, before, loaded.State(), "pure value preview must not change the existing empty-write policy")
}

func TestPrefixNetworkNormalizationContract(t *testing.T) {
	t.Parallel()

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
	for _, operation := range []string{"New", "Replace", "ApplyPatch"} {
		for _, test := range cases {
			t.Run(operation+"/"+test.name, func(t *testing.T) {
				vrf := domainipam.NonNullVRFReference(prefixVRFReference(t, 9, "Network VRF", "65000:9", true))
				values := validPrefixValues(test.input)
				values.VRF, values.Status = vrf, "reserved"
				values.IsPool, values.MarkUtilized = true, true
				values.Description, values.Comments = "original description", "original comments"
				var prefix *domainipam.Prefix
				var before domainipam.PrefixState
				var err error
				now := prefixDomainTime
				if operation == "New" {
					prefix, err = domainipam.NewPrefix(values, now)
				} else {
					seedValues := values
					seedValues.Prefix = "192.0.2.0/24"
					prefix, err = domainipam.NewPrefix(seedValues, prefixDomainTime)
					require.NoError(t, err)
					require.NoError(t, prefix.AssignID(3))
					before = prefix.State()
					now = shared.NewTimestamp(prefixDomainTime.Add(time.Hour))
					if operation == "Replace" {
						err = prefix.Replace(values, now)
					} else {
						err = prefix.ApplyPatch(domainipam.PrefixPatch{Prefix: stringPointer(test.input)}, now)
					}
				}
				if test.description != "" {
					require.Error(t, err)
					assert.Equal(t, []shared.FieldViolation{{Field: "prefix", Reason: "invalid", Description: test.description}}, shared.ViolationsOf(err))
					if operation == "New" {
						assert.Nil(t, prefix)
					} else {
						assert.Equal(t, before, prefix.State(), "rejected network must preserve complete aggregate state")
					}
					return
				}
				require.NoError(t, err)
				require.NotNil(t, prefix)
				wantID := shared.ID(0)
				if operation != "New" {
					wantID = 3
				}
				assert.True(t, prefix.Network().Valid())
				assert.Equal(t, test.canonical, prefix.Network().String())
				assert.Equal(t, test.family, prefix.Network().Family())
				assert.Equal(t, test.family, prefix.Family())
				assert.Equal(t, test.bits, prefix.Network().Bits())
				assert.Equal(t, test.canonical, prefix.Display())
				assert.Equal(t, domainipam.PrefixState{
					ID: wantID, Prefix: test.canonical, VRF: vrf, Status: "reserved",
					IsPool: true, MarkUtilized: true, Description: "original description", Comments: "original comments",
					Created: prefixDomainTime, LastUpdated: now, Children: 0, Depth: 0,
				}, prefix.State())
				assert.Equal(t, domainipam.PrefixSnapshot{
					Prefix: test.canonical, VRF: vrf, Status: "reserved", IsPool: true, MarkUtilized: true,
					Description: "original description", Comments: "original comments",
				}, prefix.Snapshot())
			})
		}
	}
}

func validPrefixValues(value string) domainipam.PrefixValues {
	return domainipam.PrefixValues{
		Prefix: value, VRF: domainipam.NullVRFReference(), Status: domainipam.PrefixStatusActive.String(),
	}
}

func prefixVRFReference(
	t *testing.T,
	id shared.ID,
	name string,
	rdValue string,
	enforceUnique bool,
) domainipam.VRFReference {
	t.Helper()
	rd, err := domainipam.ParseRouteDistinguisher(rdValue)
	require.NoError(t, err)
	reference, err := domainipam.NewVRFReference(
		id, name, domainipam.NonNullRouteDistinguisher(rd), enforceUnique,
	)
	require.NoError(t, err)
	return reference
}

func requiredPrefixNetwork(t *testing.T, value string) domainipam.PrefixNetwork {
	t.Helper()
	network, err := domainipam.ParsePrefixNetwork(value)
	require.NoError(t, err)
	return network
}

func stringPointer(value string) *string { return &value }
