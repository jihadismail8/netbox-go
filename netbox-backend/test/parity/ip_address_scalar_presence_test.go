package parity

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	ipamv1 "netbox-go/gen/go/netbox/ipam/v1"
	postgreschangelog "netbox-go/internal/adapters/postgres/changelog"
	ipamrow "netbox-go/internal/adapters/postgres/ipam/row"
	"netbox-go/internal/application/authz"
	domainipam "netbox-go/internal/domain/ipam"
)

func TestIPAddressScalarPresenceRESTGRPCParity(t *testing.T) {
	environment := newProfileParityEnvironment(t, authz.AllowAll{})
	fixtures := environment.seedProfileFixtures(t)

	var rowsBefore, changesBefore int64
	require.NoError(t, environment.db.Model(&ipamrow.IPAddressRow{}).Count(&rowsBefore).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Count(&changesBefore).Error)

	createRejections := []struct {
		name string
		body map[string]any
		want map[string]any
	}{
		{
			name: "address omitted", body: map[string]any{},
			want: map[string]any{"address": []any{"This field is required."}},
		},
		{
			name: "address null", body: map[string]any{"address": nil},
			want: map[string]any{"address": []any{"This field may not be null."}},
		},
		{
			name: "address blank", body: map[string]any{"address": ""},
			want: map[string]any{"address": []any{"This field may not be blank."}},
		},
		{
			name: "address invalid", body: map[string]any{"address": "invalid"},
			want: map[string]any{"address": []any{"Invalid IP address format: invalid"}},
		},
		{
			name: "address leading whitespace", body: map[string]any{"address": " 198.51.100.200"},
			want: map[string]any{"address": []any{"Invalid IP address format:  198.51.100.200"}},
		},
		{
			name: "status null",
			body: map[string]any{"address": "198.51.100.200", "status": nil},
			want: map[string]any{"status": []any{"This field may not be blank."}},
		},
		{
			name: "status blank",
			body: map[string]any{"address": "198.51.100.200", "status": ""},
			want: map[string]any{"status": []any{"This field may not be blank."}},
		},
		{
			name: "status invalid choice",
			body: map[string]any{"address": "198.51.100.200", "status": " invalid "},
			want: map[string]any{"status": []any{" invalid  is not a valid choice."}},
		},
		{
			name: "status boolean-like choice",
			body: map[string]any{"address": "198.51.100.200", "status": "true"},
			want: map[string]any{"status": []any{"True is not a valid choice."}},
		},
		{
			name: "role invalid choice",
			body: map[string]any{"address": "198.51.100.200", "role": " invalid "},
			want: map[string]any{"role": []any{" invalid  is not a valid choice."}},
		},
		{
			name: "role integer-like choice",
			body: map[string]any{"address": "198.51.100.200", "role": "001"},
			want: map[string]any{"role": []any{"1 is not a valid choice."}},
		},
		{
			name: "role Unicode integer-like choice",
			body: map[string]any{"address": "198.51.100.200", "role": "٠٠١"},
			want: map[string]any{"role": []any{"1 is not a valid choice."}},
		},
		{
			name: "dns name null",
			body: map[string]any{"address": "198.51.100.200", "dns_name": nil},
			want: map[string]any{"dns_name": []any{"This field may not be null."}},
		},
		{
			name: "dns name invalid",
			body: map[string]any{"address": "198.51.100.200", "dns_name": "bad name"},
			want: map[string]any{"dns_name": []any{"Only alphanumeric characters, asterisks, hyphens, periods, and underscores are allowed in DNS names"}},
		},
		{
			name: "description null",
			body: map[string]any{"address": "198.51.100.200", "description": nil},
			want: map[string]any{"description": []any{"This field may not be null."}},
		},
		{
			name: "comments null",
			body: map[string]any{"address": "198.51.100.200", "comments": nil},
			want: map[string]any{"comments": []any{"This field may not be null."}},
		},
		{
			name: "description null character",
			body: map[string]any{"address": "198.51.100.200", "description": "contains\x00null"},
			want: map[string]any{"description": []any{"Null characters are not allowed."}},
		},
		{
			name: "dns name null character",
			body: map[string]any{"address": "198.51.100.200", "dns_name": "contains\x00null"},
			want: map[string]any{"dns_name": []any{
				"Only alphanumeric characters, asterisks, hyphens, periods, and underscores are allowed in DNS names",
				"Null characters are not allowed.",
			}},
		},
	}
	for _, test := range createRejections {
		t.Run("REST create/"+test.name, func(t *testing.T) {
			restError := requestJSON(
				t, environment.router, http.MethodPost, "/api/ipam/ip-addresses",
				test.body, http.StatusBadRequest,
			)
			require.Equal(t, test.want, restError)
		})
	}

	empty := ""
	validAddress := "198.51.100.200"
	trueStatus := "true"
	unicodeIntegerRole := "٠٠١"
	nullDescription := "contains\x00null"
	nullDNSName := "contains\x00null"
	for _, input := range []*ipamv1.IPAddressInput{
		{Address: &empty},
		{Address: &validAddress, Status: &trueStatus},
		{Address: &validAddress, Role: wrapperspb.String("001")},
		{Address: &validAddress, Role: wrapperspb.String(unicodeIntegerRole)},
		{Address: &validAddress, DnsName: &nullDNSName},
		{Address: &validAddress, Description: &nullDescription},
	} {
		_, grpcErr := environment.ipam.CreateIPAddress(
			environment.ctx,
			&ipamv1.CreateIPAddressRequest{IpAddress: input},
		)
		require.Equal(t, codes.InvalidArgument, status.Code(grpcErr))
		if grpcErr == nil || status.Convert(grpcErr).Message() != "Invalid input." {
			t.Fatal("gRPC did not preserve the canonical validation envelope")
		}
	}

	var rowsAfter, changesAfter int64
	require.NoError(t, environment.db.Model(&ipamrow.IPAddressRow{}).Count(&rowsAfter).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Count(&changesAfter).Error)
	require.Equal(t, rowsBefore, rowsAfter)
	require.Equal(t, changesBefore, changesAfter)

	created := requestJSON(
		t, environment.router, http.MethodPost, "/api/ipam/ip-addresses",
		map[string]any{
			"address":              "192.0.2.44/٣٢",
			"vrf":                  fixtures.vrf,
			"status":               "reserved",
			"role":                 nil,
			"dns_name":             " Edge.EXAMPLE.Test ",
			"description":          " created description ",
			"comments":             " created comments ",
			"assigned_object_type": "dcim.interface",
			"assigned_object_id":   fixtures.iface,
		},
		http.StatusCreated,
	)
	addressID := jsonID(t, created["id"])
	require.Equal(t, "192.0.2.44/32", created["address"])
	require.Nil(t, created["role"])
	require.Equal(t, "edge.example.test", created["dns_name"])
	require.Equal(t, "created description", created["description"])
	require.Equal(t, "created comments", created["comments"])

	grpcRead, err := environment.ipam.GetIPAddress(
		environment.ctx,
		&ipamv1.GetIPAddressRequest{Id: addressID},
	)
	require.NoError(t, err)
	require.Equal(t, "192.0.2.44/32", grpcRead.IpAddress.Address)
	require.Nil(t, grpcRead.IpAddress.Role)
	require.Equal(t, "edge.example.test", grpcRead.IpAddress.DnsName)
	require.Equal(t, "created description", grpcRead.IpAddress.Description)
	require.Equal(t, "created comments", grpcRead.IpAddress.Comments)

	replaced := requestJSON(
		t, environment.router, http.MethodPut,
		"/api/ipam/ip-addresses/"+strconv.FormatInt(addressID, 10),
		map[string]any{"address": "2001:db8::44"}, http.StatusOK,
	)
	require.Equal(t, "2001:db8::44/128", replaced["address"])
	requireChoiceValue(t, replaced["status"], "reserved")
	require.Nil(t, replaced["role"])
	require.Equal(t, "edge.example.test", replaced["dns_name"])
	require.Equal(t, "created description", replaced["description"])
	require.Equal(t, "created comments", replaced["comments"])
	requireReferenceID(t, replaced["vrf"], fixtures.vrf)
	require.Equal(t, float64(fixtures.iface), replaced["assigned_object_id"])

	grpcRead, err = environment.ipam.GetIPAddress(
		environment.ctx,
		&ipamv1.GetIPAddressRequest{Id: addressID},
	)
	require.NoError(t, err)
	require.Equal(t, "2001:db8::44/128", grpcRead.IpAddress.Address)
	require.Equal(t, "reserved", grpcRead.IpAddress.Status)
	require.Nil(t, grpcRead.IpAddress.Role)
	require.Equal(t, fixtures.vrf, grpcRead.IpAddress.VrfId.Value)
	require.Equal(t, fixtures.iface, grpcRead.IpAddress.AssignedObjectId.Value)

	var beforeRejected ipamrow.IPAddressRow
	require.NoError(t, environment.db.First(&beforeRejected, addressID).Error)
	changesBeforeRejected := countParityIPAddressChanges(t, environment, addressID)
	itemPath := "/api/ipam/ip-addresses/" + strconv.FormatInt(addressID, 10)
	mutationRejections := []struct {
		name   string
		method string
		body   map[string]any
		want   map[string]any
	}{
		{
			name: "PUT address omitted", method: http.MethodPut,
			body: map[string]any{},
			want: map[string]any{"address": []any{"This field is required."}},
		},
		{
			name: "PUT address null", method: http.MethodPut,
			body: map[string]any{"address": nil},
			want: map[string]any{"address": []any{"This field may not be null."}},
		},
		{
			name: "PATCH address blank", method: http.MethodPatch,
			body: map[string]any{"address": ""},
			want: map[string]any{"address": []any{"This field may not be blank."}},
		},
		{
			name: "PATCH status null", method: http.MethodPatch,
			body: map[string]any{"status": nil},
			want: map[string]any{"status": []any{"This field may not be blank."}},
		},
		{
			name: "PATCH status blank", method: http.MethodPatch,
			body: map[string]any{"status": ""},
			want: map[string]any{"status": []any{"This field may not be blank."}},
		},
		{
			name: "PATCH status invalid choice", method: http.MethodPatch,
			body: map[string]any{"status": " invalid "},
			want: map[string]any{"status": []any{" invalid  is not a valid choice."}},
		},
		{
			name: "PATCH status boolean-like choice", method: http.MethodPatch,
			body: map[string]any{"status": "true"},
			want: map[string]any{"status": []any{"True is not a valid choice."}},
		},
		{
			name: "PATCH role invalid choice", method: http.MethodPatch,
			body: map[string]any{"role": " invalid "},
			want: map[string]any{"role": []any{" invalid  is not a valid choice."}},
		},
		{
			name: "PATCH dns name null", method: http.MethodPatch,
			body: map[string]any{"dns_name": nil},
			want: map[string]any{"dns_name": []any{"This field may not be null."}},
		},
		{
			name: "PATCH description null", method: http.MethodPatch,
			body: map[string]any{"description": nil},
			want: map[string]any{"description": []any{"This field may not be null."}},
		},
		{
			name: "PATCH comments null", method: http.MethodPatch,
			body: map[string]any{"comments": nil},
			want: map[string]any{"comments": []any{"This field may not be null."}},
		},
		{
			name: "PATCH comments null character", method: http.MethodPatch,
			body: map[string]any{"comments": "contains\x00null"},
			want: map[string]any{"comments": []any{"Null characters are not allowed."}},
		},
	}
	for _, test := range mutationRejections {
		t.Run("REST mutation/"+test.name, func(t *testing.T) {
			restError := requestJSON(
				t, environment.router, test.method, itemPath, test.body,
				http.StatusBadRequest,
			)
			require.Equal(t, test.want, restError)
			var unchanged ipamrow.IPAddressRow
			require.NoError(t, environment.db.First(&unchanged, addressID).Error)
			require.Equal(t, beforeRejected, unchanged)
			require.Equal(
				t, changesBeforeRejected,
				countParityIPAddressChanges(t, environment, addressID),
			)
		})
	}
	_, grpcErr := environment.ipam.UpdateIPAddress(
		environment.ctx,
		&ipamv1.UpdateIPAddressRequest{
			Id: addressID, IpAddress: &ipamv1.IPAddressInput{},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"status"}},
		},
	)
	require.Equal(t, codes.InvalidArgument, status.Code(grpcErr))
	if grpcErr == nil || status.Convert(grpcErr).Message() != "Invalid input." {
		t.Fatal("gRPC did not preserve the canonical update validation envelope")
	}
	var afterRejected ipamrow.IPAddressRow
	require.NoError(t, environment.db.First(&afterRejected, addressID).Error)
	require.Equal(t, beforeRejected, afterRejected)
	require.Equal(
		t, changesBeforeRejected,
		countParityIPAddressChanges(t, environment, addressID),
	)

	roleFixture := requestJSON(
		t, environment.router, http.MethodPost, "/api/ipam/ip-addresses",
		map[string]any{"address": "192.0.2.60", "role": "loopback"},
		http.StatusCreated,
	)
	roleFixtureID := jsonID(t, roleFixture["id"])
	clearedByGRPC, err := environment.ipam.UpdateIPAddress(
		environment.ctx,
		&ipamv1.UpdateIPAddressRequest{
			Id: roleFixtureID, IpAddress: &ipamv1.IPAddressInput{},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"role"}},
		},
	)
	require.NoError(t, err)
	require.Nil(t, clearedByGRPC.IpAddress.Role)
	roleRead := requestJSON(
		t, environment.router, http.MethodGet,
		"/api/ipam/ip-addresses/"+strconv.FormatInt(roleFixtureID, 10),
		nil, http.StatusOK,
	)
	require.Nil(t, roleRead["role"])

	createdByGRPC, err := environment.ipam.CreateIPAddress(
		environment.ctx,
		&ipamv1.CreateIPAddressRequest{IpAddress: &ipamv1.IPAddressInput{
			Address:     pointer("2001:db8::45/١٢٨"),
			Role:        wrapperspb.String(""),
			DnsName:     pointer(" GRPC.EXAMPLE.Test "),
			Description: pointer(" grpc description "),
			Comments:    pointer(" grpc comments "),
		}},
	)
	require.NoError(t, err)
	require.Equal(t, "2001:db8::45/128", createdByGRPC.IpAddress.Address)
	require.Nil(t, createdByGRPC.IpAddress.Role)
	require.Equal(t, "grpc.example.test", createdByGRPC.IpAddress.DnsName)
	require.Equal(t, "grpc description", createdByGRPC.IpAddress.Description)
	require.Equal(t, "grpc comments", createdByGRPC.IpAddress.Comments)

	restRead := requestJSON(
		t, environment.router, http.MethodGet,
		"/api/ipam/ip-addresses/"+strconv.FormatInt(createdByGRPC.IpAddress.Id, 10),
		nil, http.StatusOK,
	)
	require.Equal(t, "2001:db8::45/128", restRead["address"])
	require.Nil(t, restRead["role"])
	require.Equal(t, "grpc.example.test", restRead["dns_name"])
}

func requireChoiceValue(t *testing.T, raw any, expected string) {
	t.Helper()
	choice, ok := raw.(map[string]any)
	require.True(t, ok)
	require.Equal(t, expected, choice["value"])
}

func requireReferenceID(t *testing.T, raw any, expected int64) {
	t.Helper()
	reference, ok := raw.(map[string]any)
	require.True(t, ok)
	require.Equal(t, float64(expected), reference["id"])
}

func countParityIPAddressChanges(
	t *testing.T,
	environment *profileParityEnvironment,
	id int64,
) int64 {
	t.Helper()
	var count int64
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ? AND object_id = ?", domainipam.IPAddressObjectType, id,
	).Count(&count).Error)
	return count
}
