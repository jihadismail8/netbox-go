package dcim

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	postgresTransaction "netbox-go/internal/adapters/postgres/transaction"
	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

const (
	interfaceTemplateTableAlias  = "interface_templates"
	interfaceTemplateDeviceAlias = "interface_template_device_types"
)

type InterfaceTemplateRepository struct{ db *gorm.DB }

var _ applicationdcim.InterfaceTemplateRepository = (*InterfaceTemplateRepository)(nil)

func NewInterfaceTemplateRepository(db *gorm.DB) *InterfaceTemplateRepository {
	if db == nil {
		panic("dcim interface template repository requires a database")
	}
	return &InterfaceTemplateRepository{db: db}
}

func (repository *InterfaceTemplateRepository) List(
	ctx context.Context,
	criteria applicationdcim.InterfaceTemplateListCriteria,
) (applicationdcim.InterfaceTemplatePage, error) {
	base := repository.filteredQuery(ctx, criteria)
	var count int64
	if err := base.Session(&gorm.Session{}).Count(&count).Error; err != nil {
		return applicationdcim.InterfaceTemplatePage{},
			translateInterfaceTemplateReadError(0, "count InterfaceTemplates", err)
	}
	if count < 0 {
		return applicationdcim.InterfaceTemplatePage{}, shared.NewError(
			shared.ErrorReasonInternal,
			"InterfaceTemplate count returned an invalid value.",
		)
	}
	query := applyInterfaceTemplateOrdering(
		selectInterfaceTemplateProjection(base), criteria.Ordering,
	)
	if !criteria.DeferPagination && criteria.Offset > 0 {
		query = query.Offset(int(criteria.Offset))
	}
	if !criteria.DeferPagination && criteria.Limit > 0 {
		query = query.Limit(int(criteria.Limit))
	}
	var rows []interfaceTemplateProjectionRow
	if err := query.Find(&rows).Error; err != nil {
		return applicationdcim.InterfaceTemplatePage{},
			translateInterfaceTemplateReadError(0, "list InterfaceTemplates", err)
	}
	results := make([]*domaindcim.InterfaceTemplate, 0, len(rows))
	for _, row := range rows {
		template, err := interfaceTemplateFromProjection(row)
		if err != nil {
			return applicationdcim.InterfaceTemplatePage{}, err
		}
		results = append(results, template)
	}
	return applicationdcim.InterfaceTemplatePage{
		Count: uint64(count), Results: results,
	}, nil
}

func (repository *InterfaceTemplateRepository) Get(
	ctx context.Context,
	id shared.ID,
) (*domaindcim.InterfaceTemplate, error) {
	return repository.get(ctx, id, false)
}

func (repository *InterfaceTemplateRepository) GetForUpdate(
	ctx context.Context,
	id shared.ID,
) (*domaindcim.InterfaceTemplate, error) {
	return repository.get(ctx, id, true)
}

func (repository *InterfaceTemplateRepository) get(
	ctx context.Context,
	id shared.ID,
	forUpdate bool,
) (*domaindcim.InterfaceTemplate, error) {
	query := repository.baseQuery(ctx).
		Where(interfaceTemplateTableAlias+".id = ?", id.Int64())
	query = selectInterfaceTemplateProjection(query)
	if forUpdate {
		query = query.Clauses(clause.Locking{
			Strength: clause.LockingStrengthUpdate,
			Table:    clause.Table{Name: interfaceTemplateTableAlias},
		})
	}
	var row interfaceTemplateProjectionRow
	if err := query.Take(&row).Error; err != nil {
		return nil, translateInterfaceTemplateReadError(id, "get InterfaceTemplate", err)
	}
	return interfaceTemplateFromProjection(row)
}

func (repository *InterfaceTemplateRepository) Create(
	ctx context.Context,
	template *domaindcim.InterfaceTemplate,
) error {
	if template == nil {
		return shared.NewError(
			shared.ErrorReasonInternal, "Cannot persist a nil InterfaceTemplate.",
		)
	}
	if template.ID().IsValid() {
		return shared.NewError(
			shared.ErrorReasonInternal,
			"Cannot create an already persisted InterfaceTemplate.",
		)
	}
	row := interfaceTemplateToRow(*template)
	if err := repository.database(ctx).Create(&row).Error; err != nil {
		return translateInterfaceTemplateMutationError("create InterfaceTemplate", err)
	}
	return template.AssignID(shared.ID(row.ID))
}

func (repository *InterfaceTemplateRepository) Update(
	ctx context.Context,
	template *domaindcim.InterfaceTemplate,
) error {
	if template == nil || !template.ID().IsValid() {
		return shared.NewError(
			shared.ErrorReasonInternal,
			"Cannot update an unpersisted InterfaceTemplate.",
		)
	}
	row := interfaceTemplateToRow(*template)
	result := repository.database(ctx).
		Model(&dcimrow.InterfaceTemplateRow{}).
		Where("id = ?", template.ID().Int64()).
		Select(
			"device_type_id", "name", "label", "type", "enabled",
			"mgmt_only", "description", "last_updated",
		).
		Updates(&row)
	if result.Error != nil {
		return translateInterfaceTemplateMutationError("update InterfaceTemplate", result.Error)
	}
	if result.RowsAffected == 0 {
		return shared.NotFound("InterfaceTemplate", template.ID())
	}
	return nil
}

func (repository *InterfaceTemplateRepository) Delete(
	ctx context.Context,
	template *domaindcim.InterfaceTemplate,
) error {
	if template == nil || !template.ID().IsValid() {
		return shared.NewError(
			shared.ErrorReasonInternal,
			"Cannot delete an unpersisted InterfaceTemplate.",
		)
	}
	result := repository.database(ctx).
		Where("id = ?", template.ID().Int64()).
		Delete(&dcimrow.InterfaceTemplateRow{})
	if result.Error != nil {
		return translateInterfaceTemplateMutationError("delete InterfaceTemplate", result.Error)
	}
	if result.RowsAffected == 0 {
		return shared.NotFound("InterfaceTemplate", template.ID())
	}
	return nil
}

func (repository *InterfaceTemplateRepository) filteredQuery(
	ctx context.Context,
	criteria applicationdcim.InterfaceTemplateListCriteria,
) *gorm.DB {
	query := repository.baseQuery(ctx)
	if len(criteria.IDs) > 0 {
		query = query.Where(interfaceTemplateTableAlias+".id IN ?", criteria.IDs)
	}
	if criteria.VisibilityConstrained {
		if len(criteria.VisibleObjectIDs) == 0 {
			query = query.Where("1 = 0")
		} else {
			query = query.Where(
				interfaceTemplateTableAlias+".id IN ?",
				primitiveIDs(criteria.VisibleObjectIDs),
			)
		}
	}
	if criteria.Query != "" {
		pattern := containsPattern(criteria.Query)
		query = query.Where(
			"(LOWER("+interfaceTemplateTableAlias+".name) LIKE ? ESCAPE '\\' OR "+
				"LOWER("+interfaceTemplateTableAlias+".description) LIKE ? ESCAPE '\\')",
			pattern, pattern,
		)
	}
	if len(criteria.DeviceTypeIDs) > 0 {
		query = query.Where(
			interfaceTemplateTableAlias+".device_type_id IN ?", criteria.DeviceTypeIDs,
		)
	}
	if len(criteria.Names) > 0 {
		query = query.Where(interfaceTemplateTableAlias+".name IN ?", criteria.Names)
	}
	if len(criteria.Types) > 0 {
		values := make([]string, len(criteria.Types))
		for index, value := range criteria.Types {
			values[index] = value.String()
		}
		query = query.Where(interfaceTemplateTableAlias+".type IN ?", values)
	}
	if criteria.Enabled != nil {
		query = query.Where(interfaceTemplateTableAlias+".enabled = ?", *criteria.Enabled)
	}
	if criteria.MgmtOnly != nil {
		query = query.Where(interfaceTemplateTableAlias+".mgmt_only = ?", *criteria.MgmtOnly)
	}
	return query
}

func (repository *InterfaceTemplateRepository) baseQuery(ctx context.Context) *gorm.DB {
	deviceTypeTable := (dcimrow.DeviceTypeRow{}).TableName()
	return repository.database(ctx).
		Table(interfaceTemplateTableExpression()).
		Joins(
			"JOIN " + deviceTypeTable + " AS " + interfaceTemplateDeviceAlias +
				" ON " + interfaceTemplateDeviceAlias + ".id = " +
				interfaceTemplateTableAlias + ".device_type_id",
		)
}

func (repository *InterfaceTemplateRepository) database(ctx context.Context) *gorm.DB {
	db := repository.db
	if tx, ok := postgresTransaction.FromContext(ctx); ok {
		db = tx
	}
	return db.WithContext(ctx)
}

type interfaceTemplateProjectionRow struct {
	dcimrow.InterfaceTemplateRow
	DeviceTypeModel string `gorm:"column:device_type_model"`
	DeviceTypeSlug  string `gorm:"column:device_type_slug"`
}

func selectInterfaceTemplateProjection(query *gorm.DB) *gorm.DB {
	return query.Select(
		interfaceTemplateTableAlias + ".*, " +
			interfaceTemplateDeviceAlias + ".model AS device_type_model, " +
			interfaceTemplateDeviceAlias + ".slug AS device_type_slug",
	)
}

func interfaceTemplateTableExpression() string {
	return (dcimrow.InterfaceTemplateRow{}).TableName() +
		" AS " + interfaceTemplateTableAlias
}

func applyInterfaceTemplateOrdering(
	query *gorm.DB,
	ordering []applicationdcim.InterfaceTemplateSort,
) *gorm.DB {
	hasUniqueOrdering := false
	for _, requested := range ordering {
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{
				Table: interfaceTemplateTableAlias,
				Name:  interfaceTemplateSortColumn(requested.Field),
			},
			Desc: requested.Descending,
		})
		if requested.Field == applicationdcim.InterfaceTemplateSortID {
			hasUniqueOrdering = true
		}
	}
	if !hasUniqueOrdering {
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Table: interfaceTemplateTableAlias, Name: "id"},
		})
	}
	return query
}

func interfaceTemplateSortColumn(
	field applicationdcim.InterfaceTemplateSortField,
) string {
	switch field {
	case applicationdcim.InterfaceTemplateSortID:
		return "id"
	case applicationdcim.InterfaceTemplateSortDeviceType:
		return "device_type_id"
	case applicationdcim.InterfaceTemplateSortName:
		return "name"
	case applicationdcim.InterfaceTemplateSortType:
		return "type"
	case applicationdcim.InterfaceTemplateSortCreated:
		return "created"
	case applicationdcim.InterfaceTemplateSortLastUpdated:
		return "last_updated"
	default:
		return "id"
	}
}

func interfaceTemplateToRow(
	template domaindcim.InterfaceTemplate,
) dcimrow.InterfaceTemplateRow {
	state := template.State()
	return dcimrow.InterfaceTemplateRow{
		RowMetadata: dcimrow.RowMetadata{
			ID: state.ID.Int64(), Created: state.Created.Time, LastUpdated: state.LastUpdated.Time,
		},
		DeviceTypeID: state.DeviceType.ID().Int64(), Name: state.Name, Label: state.Label,
		Type: state.Type, Enabled: state.Enabled, MgmtOnly: state.MgmtOnly,
		Description: state.Description,
	}
}

func interfaceTemplateFromProjection(
	row interfaceTemplateProjectionRow,
) (*domaindcim.InterfaceTemplate, error) {
	deviceType, err := domaindcim.NewDeviceTypeReference(
		shared.ID(row.DeviceTypeID), row.DeviceTypeModel, row.DeviceTypeSlug,
	)
	if err != nil {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal,
			"Could not restore persisted InterfaceTemplate relationship.",
			err,
		)
	}
	template, err := domaindcim.RestoreInterfaceTemplate(
		domaindcim.InterfaceTemplateState{
			ID: shared.ID(row.ID), DeviceType: deviceType, Name: row.Name, Label: row.Label,
			Type: row.Type, Enabled: row.Enabled, MgmtOnly: row.MgmtOnly,
			Description: row.Description, Created: shared.NewTimestamp(row.Created),
			LastUpdated: shared.NewTimestamp(row.LastUpdated),
		},
	)
	if err != nil {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal,
			"Could not restore persisted InterfaceTemplate state.",
			err,
		)
	}
	return template, nil
}

func translateInterfaceTemplateReadError(
	id shared.ID,
	operation string,
	err error,
) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return shared.NotFound("InterfaceTemplate", id)
	}
	return shared.WrapError(
		shared.ErrorReasonInternal, fmt.Sprintf("Could not %s.", operation), err,
	)
}

func translateInterfaceTemplateMutationError(operation string, err error) error {
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "uq_go_interface_template_name"),
		strings.Contains(message, "go_dcim_interface_templates.device_type_id") &&
			strings.Contains(message, "go_dcim_interface_templates.name"):
		const description = "The fields device_type, name must make a unique set."
		return shared.ConflictWithViolations(
			description, err,
			shared.FieldViolation{
				Field: "non_field_errors", Reason: "unique_together",
				Description: description,
			},
		)
	case duplicateConstraint(err):
		return shared.Conflict("A matching InterfaceTemplate already exists.", err)
	case foreignKeyConstraint(err):
		return shared.Conflict("The selected DeviceType no longer exists.", err)
	default:
		return shared.WrapError(
			shared.ErrorReasonInternal, fmt.Sprintf("Could not %s.", operation), err,
		)
	}
}
