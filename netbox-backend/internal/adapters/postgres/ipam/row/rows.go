// Package row owns the private PostgreSQL representation of first-profile
// IPAM Managed Objects. The one DCIM dependency is the profile-declared
// IPAddress-to-Interface relationship.
package row

import dcimrow "netbox-go/internal/adapters/postgres/dcim/row"

// RowMetadata retains the common first-profile persistence metadata shape.
type RowMetadata = dcimrow.RowMetadata

type VRFRow struct {
	RowMetadata
	Name          string  `gorm:"size:100;not null;index" json:"name"`
	RD            *string `gorm:"size:21;uniqueIndex:uq_go_vrf_rd" json:"rd"`
	EnforceUnique bool    `gorm:"not null" json:"enforce_unique"`
	Description   string  `gorm:"size:200;not null;default:''" json:"description"`
	Comments      string  `gorm:"type:text;not null;default:''" json:"comments"`
}

func (VRFRow) TableName() string { return "go_ipam_vrfs" }

type PrefixRow struct {
	RowMetadata
	Prefix       string  `gorm:"size:64;not null;index:idx_go_prefix_value,priority:2" json:"prefix"`
	VRFID        *int64  `gorm:"index:idx_go_prefix_vrf;index:idx_go_prefix_value,priority:1" json:"vrf"`
	VRF          *VRFRow `gorm:"foreignKey:VRFID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	Status       string  `gorm:"size:50;not null;index" json:"status"`
	IsPool       bool    `gorm:"not null" json:"is_pool"`
	MarkUtilized bool    `gorm:"not null" json:"mark_utilized"`
	Description  string  `gorm:"size:200;not null;default:''" json:"description"`
	Comments     string  `gorm:"type:text;not null;default:''" json:"comments"`
}

func (PrefixRow) TableName() string { return "go_ipam_prefixes" }

type IPAddressRow struct {
	RowMetadata
	Address            string                `gorm:"size:64;not null;index:idx_go_ip_address_value,priority:2" json:"address"`
	VRFID              *int64                `gorm:"index:idx_go_ip_address_vrf;index:idx_go_ip_address_value,priority:1" json:"vrf"`
	VRF                *VRFRow               `gorm:"foreignKey:VRFID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	Status             string                `gorm:"size:50;not null;index" json:"status"`
	Role               *string               `gorm:"size:50;index" json:"role"`
	DNSName            string                `gorm:"size:255;not null;default:'';index" json:"dns_name"`
	Description        string                `gorm:"size:200;not null;default:''" json:"description"`
	Comments           string                `gorm:"type:text;not null;default:''" json:"comments"`
	AssignedObjectType *string               `gorm:"size:50;index:idx_go_ip_assignment,priority:1;check:chk_go_ip_assignment_pair,(assigned_object_type IS NULL AND assigned_object_id IS NULL) OR (assigned_object_type = 'dcim.interface' AND assigned_object_id IS NOT NULL)" json:"assigned_object_type"`
	AssignedObjectID   *int64                `gorm:"index:idx_go_ip_assignment,priority:2" json:"assigned_object_id"`
	AssignedInterface  *dcimrow.InterfaceRow `gorm:"foreignKey:AssignedObjectID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
}

func (IPAddressRow) TableName() string { return "go_ipam_ip_addresses" }

type Descriptor struct {
	Name         string
	Model        any
	Dependencies []string
}

// Descriptors returns IPAM tables in their existing FK-safe bootstrap order.
func Descriptors() []Descriptor {
	return []Descriptor{
		{Name: "go_ipam_vrfs", Model: &VRFRow{}},
		{Name: "go_ipam_prefixes", Model: &PrefixRow{}, Dependencies: []string{"go_ipam_vrfs"}},
		{Name: "go_ipam_ip_addresses", Model: &IPAddressRow{}, Dependencies: []string{"go_ipam_vrfs", "go_dcim_interfaces"}},
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
