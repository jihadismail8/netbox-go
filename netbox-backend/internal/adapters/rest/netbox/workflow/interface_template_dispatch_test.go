package workflow

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
)

func TestInterfaceTemplateRoutesUseTypedService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	typed := &interfaceTemplateServiceSpy{}
	router := gin.New()
	newCompleteTypedHandler(
		&typedSiteCallSpy{},
		WithInterfaceTemplateService(typed),
	).Register(router, func(c *gin.Context) {
		SetPrincipal(c, identity.Principal{ID: 1, Username: "typed-interface-template"})
		c.Next()
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(
		response,
		httptest.NewRequest(http.MethodGet, "/api/dcim/interface-templates/", nil),
	)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.Equal(t, 1, typed.listCalls)
}

type interfaceTemplateServiceSpy struct {
	listCalls      int
	createCalls    int
	updateCalls    int
	replaceCalls   int
	deleteCalls    int
	listQuery      applicationdcim.ListInterfaceTemplatesQuery
	createCommand  applicationdcim.CreateInterfaceTemplateCommand
	updateCommand  applicationdcim.UpdateInterfaceTemplateCommand
	replaceCommand applicationdcim.ReplaceInterfaceTemplateCommand
	deleteCommand  applicationdcim.DeleteInterfaceTemplateCommand
	template       *domaindcim.InterfaceTemplate
}

func (spy *interfaceTemplateServiceSpy) ListInterfaceTemplates(
	_ context.Context,
	_ identity.Principal,
	query applicationdcim.ListInterfaceTemplatesQuery,
) (applicationdcim.InterfaceTemplatePage, error) {
	spy.listCalls++
	spy.listQuery = query
	return applicationdcim.InterfaceTemplatePage{}, nil
}

func (spy *interfaceTemplateServiceSpy) GetInterfaceTemplate(
	context.Context,
	identity.Principal,
	applicationdcim.GetInterfaceTemplateQuery,
) (*domaindcim.InterfaceTemplate, error) {
	return spy.template, nil
}

func (spy *interfaceTemplateServiceSpy) CreateInterfaceTemplate(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.CreateInterfaceTemplateCommand,
) (*domaindcim.InterfaceTemplate, error) {
	spy.createCalls++
	spy.createCommand = command
	return spy.template, nil
}

func (spy *interfaceTemplateServiceSpy) ReplaceInterfaceTemplate(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.ReplaceInterfaceTemplateCommand,
) (*domaindcim.InterfaceTemplate, error) {
	spy.replaceCalls++
	spy.replaceCommand = command
	return spy.template, nil
}

func (spy *interfaceTemplateServiceSpy) UpdateInterfaceTemplate(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.UpdateInterfaceTemplateCommand,
) (*domaindcim.InterfaceTemplate, error) {
	spy.updateCalls++
	spy.updateCommand = command
	return spy.template, nil
}

func (spy *interfaceTemplateServiceSpy) DeleteInterfaceTemplate(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.DeleteInterfaceTemplateCommand,
) error {
	spy.deleteCalls++
	spy.deleteCommand = command
	return nil
}
