package identity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	application "netbox-go/internal/application/identity"
	domain "netbox-go/internal/domain/identity"
)

type Store struct{ db *gorm.DB }

const (
	// "NBXG" / "IDEN". The transaction-scoped lock makes the empty-store
	// administrator bootstrap genuinely one-time across concurrent processes.
	identityLockNamespace int32 = 0x4e425847
	identityLockID        int32 = 0x4944454e
)

func NewStore(db *gorm.DB) *Store {
	if db == nil {
		panic("identity postgres store requires a database")
	}
	return &Store{db: db}
}
func (s *Store) Transaction(ctx context.Context, fn func(application.Store) error) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if tx.Name() == "postgres" {
			if err := tx.Exec("SELECT pg_advisory_xact_lock(?, ?)", identityLockNamespace, identityLockID).Error; err != nil {
				return fmt.Errorf("lock identity lifecycle: %w", err)
			}
		}
		return fn(&Store{db: tx})
	})
}

func (s *Store) CreateUser(ctx context.Context, user domain.User, passwordHash string) (domain.User, error) {
	permissions, err := json.Marshal(user.Permissions)
	if err != nil {
		return domain.User{}, err
	}
	row := UserRow{Username: user.Username, Email: user.Email, FirstName: user.FirstName, LastName: user.LastName, PasswordHash: passwordHash, IsStaff: user.IsStaff, IsSuperuser: user.IsSuperuser, IsActive: user.IsActive, Permissions: permissions, Created: user.Created, Updated: user.Updated}
	if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.User{}, fmt.Errorf("create identity user: %w", err)
	}
	return s.userFromRow(ctx, row)
}
func (s *Store) CreateGroup(ctx context.Context, group domain.Group) (domain.Group, error) {
	row := GroupRow{Name: group.Name}
	if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.Group{}, fmt.Errorf("create identity group: %w", err)
	}
	return domain.Group{ID: row.ID, Name: row.Name}, nil
}
func (s *Store) AddGroupMember(ctx context.Context, userID, groupID int64) error {
	return s.db.WithContext(ctx).Create(&GroupMembershipRow{UserID: userID, GroupID: groupID}).Error
}
func (s *Store) CreatePermissionGrant(ctx context.Context, permission domain.PermissionGrant) (domain.PermissionGrant, error) {
	row := PermissionGrantRow{Name: permission.Name, AppLabel: permission.AppLabel, Action: permission.Action, Model: permission.Model, ObjectID: permission.ObjectID}
	if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.PermissionGrant{}, fmt.Errorf("create identity permission grant: %w", err)
	}
	permission.ID = row.ID
	return permission, nil
}
func (s *Store) GrantPermissionToUser(ctx context.Context, userID, permissionID int64) error {
	return s.db.WithContext(ctx).Create(&UserPermissionGrantRow{UserID: userID, PermissionID: permissionID}).Error
}
func (s *Store) GrantPermissionToGroup(ctx context.Context, groupID, permissionID int64) error {
	return s.db.WithContext(ctx).Create(&GroupPermissionGrantRow{GroupID: groupID, PermissionID: permissionID}).Error
}
func (s *Store) UserByUsername(ctx context.Context, username string) (domain.User, string, error) {
	var row UserRow
	err := s.db.WithContext(ctx).Where("LOWER(username) = LOWER(?)", username).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.User{}, "", application.ErrNotFound
	}
	if err != nil {
		return domain.User{}, "", err
	}
	user, err := s.userFromRow(ctx, row)
	return user, row.PasswordHash, err
}
func (s *Store) UserByID(ctx context.Context, id int64) (domain.User, string, error) {
	var row UserRow
	err := s.db.WithContext(ctx).First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.User{}, "", application.ErrNotFound
	}
	if err != nil {
		return domain.User{}, "", err
	}
	user, err := s.userFromRow(ctx, row)
	return user, row.PasswordHash, err
}
func (s *Store) UpdatePassword(ctx context.Context, id int64, hash string) error {
	result := s.db.WithContext(ctx).Model(&UserRow{}).Where("id = ?", id).Updates(map[string]any{"password_hash": hash, "updated": time.Now().UTC()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return application.ErrNotFound
	}
	return nil
}

func (s *Store) CreateSession(ctx context.Context, record application.SessionRecord) error {
	row := SessionRow{SecretHash: record.SecretHash, CSRFHash: record.CSRFHash, UserID: record.UserID, Created: record.Created, Expires: record.Expires, LastSeen: record.LastSeen}
	return s.db.WithContext(ctx).Create(&row).Error
}
func (s *Store) SessionByHash(ctx context.Context, hash []byte) (application.SessionRecord, error) {
	var row SessionRow
	err := s.db.WithContext(ctx).Where("secret_hash = ?", hash).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return application.SessionRecord{}, application.ErrNotFound
	}
	if err != nil {
		return application.SessionRecord{}, err
	}
	return application.SessionRecord{SecretHash: row.SecretHash, CSRFHash: row.CSRFHash, UserID: row.UserID, Created: row.Created, Expires: row.Expires, LastSeen: row.LastSeen}, nil
}
func (s *Store) DeleteSession(ctx context.Context, hash []byte) error {
	result := s.db.WithContext(ctx).Where("secret_hash = ?", hash).Delete(&SessionRow{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return application.ErrNotFound
	}
	return nil
}
func (s *Store) DeleteSessionsForUser(ctx context.Context, userID int64) error {
	return s.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&SessionRow{}).Error
}

func (s *Store) CreateToken(ctx context.Context, record application.TokenRecord) (domain.APIToken, error) {
	allowed, err := json.Marshal(record.Token.AllowedIPs)
	if err != nil {
		return domain.APIToken{}, err
	}
	row := TokenRow{UserID: record.Token.UserID, Display: record.Token.Display, SecretHash: record.SecretHash, Description: record.Token.Description, WriteEnabled: record.Token.WriteEnabled, AllowedIPs: allowed, Created: record.Token.Created, Expires: record.Token.Expires}
	if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.APIToken{}, err
	}
	token := record.Token
	token.ID = row.ID
	return token, nil
}
func (s *Store) TokenByHash(ctx context.Context, hash []byte) (application.TokenRecord, domain.User, error) {
	var row TokenRow
	err := s.db.WithContext(ctx).Where("secret_hash = ?", hash).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return application.TokenRecord{}, domain.User{}, application.ErrNotFound
	}
	if err != nil {
		return application.TokenRecord{}, domain.User{}, err
	}
	user, _, err := s.UserByID(ctx, row.UserID)
	if err != nil {
		return application.TokenRecord{}, domain.User{}, err
	}
	token, err := tokenFromRow(row)
	if err != nil {
		return application.TokenRecord{}, domain.User{}, err
	}
	return application.TokenRecord{Token: token, SecretHash: row.SecretHash, RevokedAt: row.RevokedAt}, user, nil
}
func (s *Store) ListTokens(ctx context.Context, userID int64, limit, offset int) ([]domain.APIToken, int64, error) {
	var count int64
	query := s.db.WithContext(ctx).Model(&TokenRow{}).Where("user_id = ? AND revoked_at IS NULL", userID)
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	var rows []TokenRow
	if err := query.Order("id").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	tokens := make([]domain.APIToken, 0, len(rows))
	for _, row := range rows {
		token, err := tokenFromRow(row)
		if err != nil {
			return nil, 0, err
		}
		tokens = append(tokens, token)
	}
	return tokens, count, nil
}
func (s *Store) RevokeToken(ctx context.Context, userID, id int64, at time.Time) error {
	result := s.db.WithContext(ctx).Model(&TokenRow{}).Where("id = ? AND user_id = ? AND revoked_at IS NULL", id, userID).Update("revoked_at", at)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return application.ErrNotFound
	}
	return nil
}
func (s *Store) TouchToken(ctx context.Context, id int64, at time.Time) error {
	return s.db.WithContext(ctx).Model(&TokenRow{}).Where("id = ? AND (last_used IS NULL OR last_used <= ?)", id, at.Add(-time.Minute)).Update("last_used", at).Error
}
func (s *Store) CountUsers(ctx context.Context) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&UserRow{}).Count(&count).Error
	return count, err
}

func (s *Store) userFromRow(ctx context.Context, row UserRow) (domain.User, error) {
	permissions := []string{}
	if len(row.Permissions) > 0 {
		if err := json.Unmarshal(row.Permissions, &permissions); err != nil {
			return domain.User{}, err
		}
	}
	permissions, visibility, err := s.effectivePermissions(ctx, row.ID, permissions)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{ID: row.ID, Username: row.Username, Email: row.Email, FirstName: row.FirstName, LastName: row.LastName, IsStaff: row.IsStaff, IsSuperuser: row.IsSuperuser, IsActive: row.IsActive, Permissions: permissions, ObjectVisibility: visibility, Created: row.Created, Updated: row.Updated}, nil
}

func (s *Store) effectivePermissions(ctx context.Context, userID int64, direct []string) ([]string, map[string]map[int64]struct{}, error) {
	var userGrants []PermissionGrantRow
	if err := s.db.WithContext(ctx).
		Table("go_identity_permission_grants AS permission").
		Select("permission.*").
		Joins("JOIN go_identity_user_permission_grants AS assignment ON assignment.permission_id = permission.id").
		Where("assignment.user_id = ?", userID).
		Scan(&userGrants).Error; err != nil {
		return nil, nil, fmt.Errorf("load direct identity grants: %w", err)
	}
	var groupGrants []PermissionGrantRow
	if err := s.db.WithContext(ctx).
		Table("go_identity_permission_grants AS permission").
		Select("permission.*").
		Joins("JOIN go_identity_group_permission_grants AS assignment ON assignment.permission_id = permission.id").
		Joins("JOIN go_identity_group_memberships AS membership ON membership.group_id = assignment.group_id").
		Where("membership.user_id = ?", userID).
		Scan(&groupGrants).Error; err != nil {
		return nil, nil, fmt.Errorf("load group identity grants: %w", err)
	}

	effective := make(map[string]struct{}, len(direct)+len(userGrants)+len(groupGrants))
	global := make(map[string]struct{}, len(direct))
	visibility := make(map[string]map[int64]struct{})
	for _, raw := range direct {
		permission := strings.ToLower(strings.TrimSpace(raw))
		if permission == "" {
			continue
		}
		effective[permission] = struct{}{}
		global[permission] = struct{}{}
	}
	apply := func(row PermissionGrantRow) {
		grant := domain.PermissionGrant{AppLabel: row.AppLabel, Action: row.Action, Model: row.Model, ObjectID: row.ObjectID}
		permission := grant.Codename()
		effective[permission] = struct{}{}
		if row.ObjectID == nil {
			global[permission] = struct{}{}
			delete(visibility, permission)
			return
		}
		if _, globallyGranted := global[permission]; globallyGranted {
			return
		}
		if visibility[permission] == nil {
			visibility[permission] = make(map[int64]struct{})
		}
		visibility[permission][*row.ObjectID] = struct{}{}
	}
	for _, row := range userGrants {
		apply(row)
	}
	for _, row := range groupGrants {
		apply(row)
	}
	result := make([]string, 0, len(effective))
	for permission := range effective {
		result = append(result, permission)
	}
	sort.Strings(result)
	return result, visibility, nil
}
func tokenFromRow(row TokenRow) (domain.APIToken, error) {
	allowed := []string{}
	if len(row.AllowedIPs) > 0 {
		if err := json.Unmarshal(row.AllowedIPs, &allowed); err != nil {
			return domain.APIToken{}, err
		}
	}
	return domain.APIToken{ID: row.ID, UserID: row.UserID, Display: row.Display, Description: row.Description, WriteEnabled: row.WriteEnabled, AllowedIPs: allowed, Created: row.Created, Expires: row.Expires, LastUsed: row.LastUsed}, nil
}
