// Package row owns the private PostgreSQL representation of first-profile
// DCIM Managed Objects. These types never cross the persistence adapter.
package row

import "time"

// RowMetadata is embedded by every private first-profile DCIM row.
type RowMetadata struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"-"`
	Created     time.Time `gorm:"not null" json:"-"`
	LastUpdated time.Time `gorm:"not null;index" json:"-"`
}

type SiteRow struct {
	RowMetadata
	Name        string `gorm:"size:100;not null;uniqueIndex:uq_go_site_name" json:"name"`
	Slug        string `gorm:"size:100;not null;uniqueIndex:uq_go_site_slug" json:"slug"`
	Status      string `gorm:"size:50;not null;index" json:"status"`
	Facility    string `gorm:"size:50;not null;default:''" json:"facility"`
	Description string `gorm:"size:200;not null;default:''" json:"description"`
	Comments    string `gorm:"type:text;not null;default:''" json:"comments"`
}

func (SiteRow) TableName() string { return "go_dcim_sites" }

type ManufacturerRow struct {
	RowMetadata
	Name        string `gorm:"size:100;not null;uniqueIndex:uq_go_manufacturer_name" json:"name"`
	Slug        string `gorm:"size:100;not null;uniqueIndex:uq_go_manufacturer_slug" json:"slug"`
	Description string `gorm:"size:200;not null;default:''" json:"description"`
}

func (ManufacturerRow) TableName() string { return "go_dcim_manufacturers" }

type RackRoleRow struct {
	RowMetadata
	Name        string `gorm:"size:100;not null;uniqueIndex:uq_go_rack_role_name" json:"name"`
	Slug        string `gorm:"size:100;not null;uniqueIndex:uq_go_rack_role_slug" json:"slug"`
	Color       string `gorm:"type:char(6);not null" json:"color"`
	Description string `gorm:"size:200;not null;default:''" json:"description"`
}

func (RackRoleRow) TableName() string { return "go_dcim_rack_roles" }

type RackTypeRow struct {
	RowMetadata
	ManufacturerID int64            `gorm:"not null;index:idx_go_rack_type_manufacturer;uniqueIndex:uq_go_rack_type_model,priority:1" json:"manufacturer"`
	Manufacturer   *ManufacturerRow `gorm:"foreignKey:ManufacturerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	Model          string           `gorm:"size:100;not null;uniqueIndex:uq_go_rack_type_model,priority:2" json:"model"`
	Slug           string           `gorm:"size:100;not null;uniqueIndex:uq_go_rack_type_slug" json:"slug"`
	FormFactor     string           `gorm:"size:50;not null;index" json:"form_factor"`
	Width          int64            `gorm:"type:smallint;not null" json:"width"`
	UHeight        int64            `gorm:"type:smallint;not null" json:"u_height"`
	StartingUnit   int64            `gorm:"type:smallint;not null" json:"starting_unit"`
	DescUnits      bool             `gorm:"not null" json:"desc_units"`
	Description    string           `gorm:"size:200;not null;default:''" json:"description"`
	Comments       string           `gorm:"type:text;not null;default:''" json:"comments"`
}

func (RackTypeRow) TableName() string { return "go_dcim_rack_types" }

type RackRow struct {
	RowMetadata
	SiteID       int64        `gorm:"not null;index:idx_go_rack_site" json:"site"`
	Site         *SiteRow     `gorm:"foreignKey:SiteID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	Name         string       `gorm:"size:100;not null;index" json:"name"`
	FacilityID   *string      `gorm:"size:50;index" json:"facility_id"`
	RackTypeID   *int64       `gorm:"index:idx_go_rack_type" json:"rack_type"`
	RackType     *RackTypeRow `gorm:"foreignKey:RackTypeID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	Status       string       `gorm:"size:50;not null;index" json:"status"`
	RoleID       *int64       `gorm:"index:idx_go_rack_role" json:"role"`
	Role         *RackRoleRow `gorm:"foreignKey:RoleID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	Serial       string       `gorm:"size:50;not null;default:''" json:"serial"`
	AssetTag     *string      `gorm:"size:50;uniqueIndex:uq_go_rack_asset_tag" json:"asset_tag"`
	FormFactor   *string      `gorm:"size:50" json:"form_factor"`
	Width        int64        `gorm:"type:smallint;not null" json:"width"`
	UHeight      int64        `gorm:"type:smallint;not null" json:"u_height"`
	StartingUnit int64        `gorm:"type:smallint;not null" json:"starting_unit"`
	DescUnits    bool         `gorm:"not null" json:"desc_units"`
	Airflow      *string      `gorm:"size:50" json:"airflow"`
	Description  string       `gorm:"size:200;not null;default:''" json:"description"`
	Comments     string       `gorm:"type:text;not null;default:''" json:"comments"`
}

func (RackRow) TableName() string { return "go_dcim_racks" }

type DeviceRoleRow struct {
	RowMetadata
	ParentID    *int64         `gorm:"index:idx_go_device_role_parent" json:"parent"`
	Parent      *DeviceRoleRow `gorm:"foreignKey:ParentID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	Name        string         `gorm:"size:100;not null;uniqueIndex:uq_go_device_role_parent_name,expression:(CASE WHEN parent_id IS NULL THEN '0' ELSE CAST(parent_id AS TEXT) END || ':' || name)" json:"name"`
	Slug        string         `gorm:"size:100;not null;uniqueIndex:uq_go_device_role_parent_slug,expression:(CASE WHEN parent_id IS NULL THEN '0' ELSE CAST(parent_id AS TEXT) END || ':' || slug)" json:"slug"`
	Color       string         `gorm:"type:char(6);not null" json:"color"`
	VMRole      bool           `gorm:"not null" json:"vm_role"`
	Description string         `gorm:"size:200;not null;default:''" json:"description"`
	Comments    string         `gorm:"type:text;not null;default:''" json:"comments"`
}

func (DeviceRoleRow) TableName() string { return "go_dcim_device_roles" }

type DeviceTypeRow struct {
	RowMetadata
	ManufacturerID         int64            `gorm:"not null;index:idx_go_device_type_manufacturer;uniqueIndex:uq_go_device_type_model,priority:1" json:"manufacturer"`
	Manufacturer           *ManufacturerRow `gorm:"foreignKey:ManufacturerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	Model                  string           `gorm:"size:100;not null;uniqueIndex:uq_go_device_type_model,priority:2" json:"model"`
	Slug                   string           `gorm:"size:100;not null;uniqueIndex:uq_go_device_type_slug,expression:(CAST(manufacturer_id AS TEXT) || ':' || slug)" json:"slug"`
	PartNumber             string           `gorm:"size:50;not null;default:''" json:"part_number"`
	UHeight                float64          `gorm:"type:numeric(4,1);not null" json:"u_height"`
	ExcludeFromUtilization bool             `gorm:"not null" json:"exclude_from_utilization"`
	IsFullDepth            bool             `gorm:"not null" json:"is_full_depth"`
	Airflow                *string          `gorm:"size:50" json:"airflow"`
	Description            string           `gorm:"size:200;not null;default:''" json:"description"`
	Comments               string           `gorm:"type:text;not null;default:''" json:"comments"`
}

func (DeviceTypeRow) TableName() string { return "go_dcim_device_types" }

type InterfaceTemplateRow struct {
	RowMetadata
	DeviceTypeID int64          `gorm:"not null;index:idx_go_interface_template_type;uniqueIndex:uq_go_interface_template_name,priority:1" json:"device_type"`
	DeviceType   *DeviceTypeRow `gorm:"foreignKey:DeviceTypeID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	Name         string         `gorm:"size:64;not null;uniqueIndex:uq_go_interface_template_name,priority:2" json:"name"`
	Label        string         `gorm:"size:64;not null;default:''" json:"label"`
	Type         string         `gorm:"size:50;not null;index" json:"type"`
	Enabled      bool           `gorm:"not null" json:"enabled"`
	MgmtOnly     bool           `gorm:"not null" json:"mgmt_only"`
	Description  string         `gorm:"size:200;not null;default:''" json:"description"`
}

func (InterfaceTemplateRow) TableName() string { return "go_dcim_interface_templates" }

type DeviceRow struct {
	RowMetadata
	DeviceTypeID int64          `gorm:"not null;index:idx_go_device_type" json:"device_type"`
	DeviceType   *DeviceTypeRow `gorm:"foreignKey:DeviceTypeID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	RoleID       int64          `gorm:"not null;index:idx_go_device_role" json:"role"`
	Role         *DeviceRoleRow `gorm:"foreignKey:RoleID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	Name         *string        `gorm:"size:64;uniqueIndex:uq_go_device_site_name_ci,priority:2,expression:lower(name),where:name IS NOT NULL" json:"name"`
	SiteID       int64          `gorm:"not null;index:idx_go_device_site;uniqueIndex:uq_go_device_site_name_ci,priority:1,where:name IS NOT NULL" json:"site"`
	Site         *SiteRow       `gorm:"foreignKey:SiteID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	RackID       *int64         `gorm:"index:idx_go_device_rack;uniqueIndex:uq_go_device_rack_position_face,priority:1,where:rack_id IS NOT NULL AND position IS NOT NULL" json:"rack"`
	Rack         *RackRow       `gorm:"foreignKey:RackID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	Position     *float64       `gorm:"type:numeric(4,1);index:idx_go_device_position;uniqueIndex:uq_go_device_rack_position_face,priority:2,where:rack_id IS NOT NULL AND position IS NOT NULL" json:"position"`
	Face         string         `gorm:"size:50;not null;default:'';uniqueIndex:uq_go_device_rack_position_face,priority:3,where:rack_id IS NOT NULL AND position IS NOT NULL" json:"face"`
	Status       string         `gorm:"size:50;not null;index" json:"status"`
	Serial       string         `gorm:"size:50;not null;default:''" json:"serial"`
	AssetTag     *string        `gorm:"size:50;uniqueIndex:uq_go_device_asset_tag" json:"asset_tag"`
	Airflow      *string        `gorm:"size:50" json:"airflow"`
	Description  string         `gorm:"size:200;not null;default:''" json:"description"`
	Comments     string         `gorm:"type:text;not null;default:''" json:"comments"`
}

func (DeviceRow) TableName() string { return "go_dcim_devices" }

type InterfaceRow struct {
	RowMetadata
	DeviceID    int64      `gorm:"not null;index:idx_go_interface_device;uniqueIndex:uq_go_interface_name,priority:1" json:"device"`
	Device      *DeviceRow `gorm:"foreignKey:DeviceID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	Name        string     `gorm:"size:64;not null;uniqueIndex:uq_go_interface_name,priority:2" json:"name"`
	Label       string     `gorm:"size:64;not null;default:''" json:"label"`
	Type        string     `gorm:"size:50;not null;index" json:"type"`
	Enabled     bool       `gorm:"not null" json:"enabled"`
	MgmtOnly    bool       `gorm:"not null" json:"mgmt_only"`
	MTU         *int64     `gorm:"type:integer" json:"mtu"`
	Speed       *int64     `gorm:"type:integer" json:"speed"`
	Duplex      *string    `gorm:"size:50" json:"duplex"`
	Description string     `gorm:"size:200;not null;default:''" json:"description"`
}

func (InterfaceRow) TableName() string { return "go_dcim_interfaces" }

type Descriptor struct {
	Name         string
	Model        any
	Dependencies []string
}

// Descriptors returns DCIM tables in their existing FK-safe bootstrap order.
func Descriptors() []Descriptor {
	return []Descriptor{
		{Name: "go_dcim_sites", Model: &SiteRow{}},
		{Name: "go_dcim_manufacturers", Model: &ManufacturerRow{}},
		{Name: "go_dcim_rack_roles", Model: &RackRoleRow{}},
		{Name: "go_dcim_rack_types", Model: &RackTypeRow{}, Dependencies: []string{"go_dcim_manufacturers"}},
		{Name: "go_dcim_racks", Model: &RackRow{}, Dependencies: []string{"go_dcim_sites", "go_dcim_rack_roles", "go_dcim_rack_types"}},
		{Name: "go_dcim_device_roles", Model: &DeviceRoleRow{}},
		{Name: "go_dcim_device_types", Model: &DeviceTypeRow{}, Dependencies: []string{"go_dcim_manufacturers"}},
		{Name: "go_dcim_interface_templates", Model: &InterfaceTemplateRow{}, Dependencies: []string{"go_dcim_device_types"}},
		{Name: "go_dcim_devices", Model: &DeviceRow{}, Dependencies: []string{"go_dcim_sites", "go_dcim_racks", "go_dcim_device_roles", "go_dcim_device_types"}},
		{Name: "go_dcim_interfaces", Model: &InterfaceRow{}, Dependencies: []string{"go_dcim_devices"}},
	}
}

func Models() []any {
	descriptors := Descriptors()
	models := make([]any, 0, len(descriptors))
	for _, descriptor := range descriptors {
		models = append(models, descriptor.Model)
	}
	return models
}
