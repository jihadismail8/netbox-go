package workflow

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
)

func TestInterfaceTemplateRPCDispatchUsesTypedService(t *testing.T) {
	typed := &grpcInterfaceTemplateServiceSpy{}
	services := completeTypedDCIMTestServices()
	services.interfaceTemplates = typed
	server := services.server()
	ctx := identity.WithPrincipal(
		t.Context(),
		identity.Principal{ID: 1, Username: "typed-interface-template"},
	)

	response, err := server.ListInterfaceTemplates(
		ctx, &dcimv1.ListInterfaceTemplatesRequest{},
	)
	require.NoError(t, err)
	require.Equal(t, uint64(0), response.Page.Count)
	require.Equal(t, 1, typed.listCalls)
}

type grpcInterfaceTemplateServiceSpy struct {
	listCalls      int
	createCalls    int
	replaceCalls   int
	updateCalls    int
	deleteCalls    int
	listQuery      applicationdcim.ListInterfaceTemplatesQuery
	createCommand  applicationdcim.CreateInterfaceTemplateCommand
	replaceCommand applicationdcim.ReplaceInterfaceTemplateCommand
	updateCommand  applicationdcim.UpdateInterfaceTemplateCommand
	deleteCommand  applicationdcim.DeleteInterfaceTemplateCommand
	template       *domaindcim.InterfaceTemplate
}

func (spy *grpcInterfaceTemplateServiceSpy) ListInterfaceTemplates(
	_ context.Context,
	_ identity.Principal,
	query applicationdcim.ListInterfaceTemplatesQuery,
) (applicationdcim.InterfaceTemplatePage, error) {
	spy.listCalls++
	spy.listQuery = query
	return applicationdcim.InterfaceTemplatePage{}, nil
}

func (spy *grpcInterfaceTemplateServiceSpy) GetInterfaceTemplate(
	context.Context,
	identity.Principal,
	applicationdcim.GetInterfaceTemplateQuery,
) (*domaindcim.InterfaceTemplate, error) {
	return spy.template, nil
}

func (spy *grpcInterfaceTemplateServiceSpy) CreateInterfaceTemplate(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.CreateInterfaceTemplateCommand,
) (*domaindcim.InterfaceTemplate, error) {
	spy.createCalls++
	spy.createCommand = command
	return spy.template, nil
}

func (spy *grpcInterfaceTemplateServiceSpy) ReplaceInterfaceTemplate(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.ReplaceInterfaceTemplateCommand,
) (*domaindcim.InterfaceTemplate, error) {
	spy.replaceCalls++
	spy.replaceCommand = command
	return spy.template, nil
}

func (spy *grpcInterfaceTemplateServiceSpy) UpdateInterfaceTemplate(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.UpdateInterfaceTemplateCommand,
) (*domaindcim.InterfaceTemplate, error) {
	spy.updateCalls++
	spy.updateCommand = command
	return spy.template, nil
}

func (spy *grpcInterfaceTemplateServiceSpy) DeleteInterfaceTemplate(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.DeleteInterfaceTemplateCommand,
) error {
	spy.deleteCalls++
	spy.deleteCommand = command
	return nil
}
