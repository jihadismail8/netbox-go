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

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	postgreschangelog "netbox-go/internal/adapters/postgres/changelog"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	"netbox-go/internal/application/authz"
	domaindcim "netbox-go/internal/domain/dcim"
)

func TestDeviceRoleScalarPresenceRESTGRPCParity(t *testing.T) {
	environment := newProfileParityEnvironment(t, authz.AllowAll{})

	root := requestJSON(
		t,
		environment.router,
		http.MethodPost,
		"/api/dcim/device-roles",
		map[string]any{"name": "  Root Role  ", "slug": "  root-role  "},
		http.StatusCreated,
	)
	rootID := jsonID(t, root["id"])
	requireDeviceRoleRESTScalars(
		t, root, nil, "Root Role", "root-role", domaindcim.DeviceRoleDefaultColor, true, "", "",
	)
	rootState := loadParityDeviceRolePresenceState(t, environment, rootID)
	require.Equal(t, int64(1), rootState.changeCount)

	rootGRPC, err := environment.dcim.GetDeviceRole(
		environment.ctx,
		&dcimv1.GetDeviceRoleRequest{Id: rootID},
	)
	require.NoError(t, err)
	requireDeviceRoleProtoScalars(
		t, rootGRPC.DeviceRole, nil, "Root Role", "root-role",
		domaindcim.DeviceRoleDefaultColor, true, "", "",
	)

	vmRoleFalse := false
	childCreated, err := environment.dcim.CreateDeviceRole(
		environment.ctx,
		&dcimv1.CreateDeviceRoleRequest{DeviceRole: &dcimv1.DeviceRoleInput{
			Parent: wrapperspb.Int64(rootID),
			Name:   pointer("  Child Role  "), Slug: pointer("  child-role  "),
			Color: pointer("  00ff00  "), VmRole: &vmRoleFalse,
			Description: pointer("  child description  "), Comments: pointer("  child comments  "),
		}},
	)
	require.NoError(t, err)
	childID := childCreated.DeviceRole.Id
	requireDeviceRoleProtoScalars(
		t, childCreated.DeviceRole, &rootID, "Child Role", "child-role", "00ff00", false,
		"child description", "child comments",
	)
	childState := loadParityDeviceRolePresenceState(t, environment, childID)
	require.Equal(t, rootState.roleCount+1, childState.roleCount)
	require.Equal(t, rootState.totalChangeCount+1, childState.totalChangeCount)
	require.Equal(t, int64(1), childState.changeCount)

	itemPath := "/api/dcim/device-roles/" + strconv.FormatInt(childID, 10)
	childREST := requestJSON(
		t, environment.router, http.MethodGet, itemPath, nil, http.StatusOK,
	)
	requireDeviceRoleRESTScalars(
		t, childREST, &rootID, "Child Role", "child-role", "00ff00", false,
		"child description", "child comments",
	)

	beforePut := loadParityDeviceRolePresenceState(t, environment, childID)
	replaced := requestJSON(
		t,
		environment.router,
		http.MethodPut,
		itemPath,
		map[string]any{"name": "  Replaced Child  ", "slug": "  replaced-child  "},
		http.StatusOK,
	)
	requireDeviceRoleRESTScalars(
		t, replaced, nil, "Replaced Child", "replaced-child", "00ff00", false,
		"child description", "child comments",
	)
	afterPut := loadParityDeviceRolePresenceState(t, environment, childID)
	requireDeviceRoleParityUpdateRecorded(t, beforePut, afterPut)
	require.Nil(t, afterPut.row.ParentID, "PUT parent omission must reset the relationship to root")

	grpcReplacementName := "  gRPC Replaced Child  "
	grpcReplacementSlug := "  grpc-replaced-child  "
	replacedByGRPC, err := environment.dcim.ReplaceDeviceRole(
		environment.ctx,
		&dcimv1.ReplaceDeviceRoleRequest{
			Id: childID,
			DeviceRole: &dcimv1.DeviceRoleInput{
				Name: &grpcReplacementName,
				Slug: &grpcReplacementSlug,
			},
		},
	)
	require.NoError(t, err)
	requireDeviceRoleProtoScalars(
		t, replacedByGRPC.DeviceRole, nil, "gRPC Replaced Child", "grpc-replaced-child",
		"00ff00", false, "child description", "child comments",
	)
	afterGRPCReplace := loadParityDeviceRolePresenceState(t, environment, childID)
	requireDeviceRoleParityUpdateRecorded(t, afterPut, afterGRPCReplace)
	require.Nil(t, afterGRPCReplace.row.ParentID, "gRPC PUT parent omission must reset to root")

	patchedByGRPC, err := environment.dcim.UpdateDeviceRole(
		environment.ctx,
		&dcimv1.UpdateDeviceRoleRequest{
			Id: childID,
			DeviceRole: &dcimv1.DeviceRoleInput{
				Parent: wrapperspb.Int64(rootID), VmRole: &vmRoleFalse,
				Description: pointer(""), Comments: pointer(""),
			},
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: []string{"parent", "vm_role", "description", "comments"},
			},
		},
	)
	require.NoError(t, err)
	requireDeviceRoleProtoScalars(
		t, patchedByGRPC.DeviceRole, &rootID, "gRPC Replaced Child", "grpc-replaced-child", "00ff00",
		false, "", "",
	)
	afterGRPCPatch := loadParityDeviceRolePresenceState(t, environment, childID)
	requireDeviceRoleParityUpdateRecorded(t, afterGRPCReplace, afterGRPCPatch)

	clearedByGRPC, err := environment.dcim.UpdateDeviceRole(
		environment.ctx,
		&dcimv1.UpdateDeviceRoleRequest{
			Id: childID, DeviceRole: &dcimv1.DeviceRoleInput{},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"parent"}},
		},
	)
	require.NoError(t, err)
	requireDeviceRoleProtoScalars(
		t, clearedByGRPC.DeviceRole, nil, "gRPC Replaced Child", "grpc-replaced-child", "00ff00",
		false, "", "",
	)
	afterGRPCClear := loadParityDeviceRolePresenceState(t, environment, childID)
	requireDeviceRoleParityUpdateRecorded(t, afterGRPCPatch, afterGRPCClear)
	require.Nil(t, afterGRPCClear.row.ParentID)

	patchedByREST := requestJSON(
		t,
		environment.router,
		http.MethodPatch,
		itemPath,
		map[string]any{"color": "  1122aa  ", "vm_role": false, "description": "", "comments": ""},
		http.StatusOK,
	)
	requireDeviceRoleRESTScalars(
		t, patchedByREST, nil, "gRPC Replaced Child", "grpc-replaced-child", "1122aa", false, "", "",
	)
	afterRESTPatch := loadParityDeviceRolePresenceState(t, environment, childID)
	requireDeviceRoleParityUpdateRecorded(t, afterGRPCClear, afterRESTPatch)

	restRejections := []struct {
		name string
		body map[string]any
		want map[string]any
	}{
		{
			name: "name null", body: map[string]any{"name": nil},
			want: map[string]any{"name": []any{"This field may not be null."}},
		},
		{
			name: "slug blank", body: map[string]any{"slug": ""},
			want: map[string]any{"slug": []any{"This field may not be blank."}},
		},
		{
			name: "color uppercase", body: map[string]any{"color": "ABCDEF"},
			want: map[string]any{"color": []any{"Enter a valid hexadecimal RGB color code."}},
		},
		{
			name: "vm role null", body: map[string]any{"vm_role": nil},
			want: map[string]any{"vm_role": []any{"This field may not be null."}},
		},
		{
			name: "description null", body: map[string]any{"description": nil},
			want: map[string]any{"description": []any{"This field may not be null."}},
		},
		{
			name: "comments null", body: map[string]any{"comments": nil},
			want: map[string]any{"comments": []any{"This field may not be null."}},
		},
	}
	for _, test := range restRejections {
		test := test
		t.Run("REST PATCH rejection/"+test.name, func(t *testing.T) {
			before := loadParityDeviceRolePresenceState(t, environment, childID)
			got := requestJSON(
				t, environment.router, http.MethodPatch, itemPath, test.body, http.StatusBadRequest,
			)
			require.Equal(t, test.want, got)
			require.Equal(t, before, loadParityDeviceRolePresenceState(t, environment, childID))
		})
	}

	for _, field := range []string{"name", "slug", "color", "vm_role", "description", "comments"} {
		field := field
		t.Run("gRPC FieldMask null/"+field, func(t *testing.T) {
			before := loadParityDeviceRolePresenceState(t, environment, childID)
			_, updateErr := environment.dcim.UpdateDeviceRole(
				environment.ctx,
				&dcimv1.UpdateDeviceRoleRequest{
					Id: childID, DeviceRole: &dcimv1.DeviceRoleInput{},
					UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{field}},
				},
			)
			requireDeviceRoleGRPCInvalid(t, updateErr)
			require.Equal(t, before, loadParityDeviceRolePresenceState(t, environment, childID))
		})
	}

	finalGRPC, err := environment.dcim.GetDeviceRole(
		environment.ctx,
		&dcimv1.GetDeviceRoleRequest{Id: childID},
	)
	require.NoError(t, err)
	requireDeviceRoleProtoScalars(
		t, finalGRPC.DeviceRole, nil, "gRPC Replaced Child", "grpc-replaced-child", "1122aa", false, "", "",
	)
}

type parityDeviceRolePresenceState struct {
	row              dcimrow.DeviceRoleRow
	roleCount        int64
	changeCount      int64
	totalChangeCount int64
}

func loadParityDeviceRolePresenceState(
	t *testing.T,
	environment *profileParityEnvironment,
	id int64,
) parityDeviceRolePresenceState {
	t.Helper()
	var state parityDeviceRolePresenceState
	require.NoError(t, environment.db.First(&state.row, id).Error)
	require.NoError(t, environment.db.Model(&dcimrow.DeviceRoleRow{}).Count(&state.roleCount).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ? AND object_id = ?", domaindcim.DeviceRoleObjectType, id,
	).Count(&state.changeCount).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Count(&state.totalChangeCount).Error)
	return state
}

func requireDeviceRoleParityUpdateRecorded(
	t *testing.T,
	before parityDeviceRolePresenceState,
	after parityDeviceRolePresenceState,
) {
	t.Helper()
	require.Equal(t, before.roleCount, after.roleCount)
	require.Equal(t, before.changeCount+1, after.changeCount)
	require.Equal(t, before.totalChangeCount+1, after.totalChangeCount)
	require.Equal(t, before.row.Created, after.row.Created)
}

func requireDeviceRoleGRPCInvalid(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	require.Equal(t, "Invalid input.", status.Convert(err).Message())
}

func requireDeviceRoleProtoScalars(
	t *testing.T,
	role *dcimv1.DeviceRole,
	parentID *int64,
	name string,
	slug string,
	color string,
	vmRole bool,
	description string,
	comments string,
) {
	t.Helper()
	require.NotNil(t, role)
	if parentID == nil {
		require.Nil(t, role.ParentId)
	} else {
		require.NotNil(t, role.ParentId)
		require.Equal(t, *parentID, role.ParentId.Value)
	}
	require.Equal(t, name, role.Name)
	require.Equal(t, slug, role.Slug)
	require.Equal(t, color, role.Color)
	require.Equal(t, vmRole, role.VmRole)
	require.Equal(t, description, role.Description)
	require.Equal(t, comments, role.Comments)
}

func requireDeviceRoleRESTScalars(
	t *testing.T,
	role map[string]any,
	parentID *int64,
	name string,
	slug string,
	color string,
	vmRole bool,
	description string,
	comments string,
) {
	t.Helper()
	if parentID == nil {
		require.Nil(t, role["parent"])
	} else {
		parent, ok := role["parent"].(map[string]any)
		require.True(t, ok)
		require.Equal(t, float64(*parentID), parent["id"])
	}
	require.Equal(t, name, role["name"])
	require.Equal(t, slug, role["slug"])
	require.Equal(t, color, role["color"])
	require.Equal(t, vmRole, role["vm_role"])
	require.Equal(t, description, role["description"])
	require.Equal(t, comments, role["comments"])
}
