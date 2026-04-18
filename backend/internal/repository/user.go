package repository

import (
	"time"

	"codemind/internal/model"

	"gorm.io/gorm"
)

// UserRepository handles user data access.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a user repository.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user.
func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// FindByID finds a user by ID with department preloaded.
func (r *UserRepository) FindByID(id int64) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Department").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByUsername finds a user by username.
func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Department").Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail finds a user by email.
func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update updates user information.
func (r *UserRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

// UpdateFields updates specific fields.
func (r *UserRepository) UpdateFields(id int64, fields map[string]interface{}) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Updates(fields).Error
}

// Delete soft-deletes a user.
func (r *UserRepository) Delete(id int64) error {
	return r.db.Delete(&model.User{}, id).Error
}

// List returns paginated user list with filters.
func (r *UserRepository) List(page, pageSize int, filters map[string]interface{}) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := r.db.Model(&model.User{}).Preload("Department")

	if keyword, ok := filters["keyword"].(string); ok && keyword != "" {
		likePattern := "%" + keyword + "%"
		query = query.Where("username LIKE ? OR display_name LIKE ? OR email LIKE ?",
			likePattern, likePattern, likePattern)
	}
	if deptID, ok := filters["department_id"].(*int64); ok && deptID != nil {
		query = query.Where("department_id = ?", *deptID)
	}
	if role, ok := filters["role"].(string); ok && role != "" {
		query = query.Where("role = ?", role)
	}
	if status, ok := filters["status"].(*int16); ok && status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// ListByDepartment returns users in a department.
func (r *UserRepository) ListByDepartment(deptID int64, page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := r.db.Model(&model.User{}).Where("department_id = ?", deptID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Department").Order("id DESC").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// CountByDepartment counts users in a department.
func (r *UserRepository) CountByDepartment(deptID int64) (int64, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("department_id = ?", deptID).Count(&count).Error
	return count, err
}

// CountByDepartmentBatch counts users per department in a single query.
func (r *UserRepository) CountByDepartmentBatch() (map[int64]int, error) {
	type DeptCount struct {
		DepartmentID int64 `gorm:"column:department_id"`
		Count        int   `gorm:"column:count"`
	}
	var rows []DeptCount
	err := r.db.Model(&model.User{}).
		Select("department_id, COUNT(*) as count").
		Where("department_id IS NOT NULL AND deleted_at IS NULL").
		Group("department_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64]int, len(rows))
	for _, r := range rows {
		result[r.DepartmentID] = r.Count
	}
	return result, nil
}

// CountAll counts all users.
func (r *UserRepository) CountAll() (int64, error) {
	var count int64
	err := r.db.Model(&model.User{}).Count(&count).Error
	return count, err
}

// ExistsUsername checks if username exists (excludes soft-deleted).
func (r *UserRepository) ExistsUsername(username string) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

// ExistsUsernameIncludingDeleted checks if username exists including soft-deleted.
func (r *UserRepository) ExistsUsernameIncludingDeleted(username string) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Unscoped().Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

// HardDeleteSoftDeletedUser permanently deletes a soft-deleted user by username.
func (r *UserRepository) HardDeleteSoftDeletedUser(username string) error {
	return r.db.Unscoped().Where("username = ? AND deleted_at IS NOT NULL", username).Delete(&model.User{}).Error
}

// ExistsEmail checks if email exists (excludes soft-deleted).
func (r *UserRepository) ExistsEmail(email string, excludeUserID ...int64) (bool, error) {
	var count int64
	query := r.db.Model(&model.User{}).Where("email = ?", email)
	if len(excludeUserID) > 0 {
		query = query.Where("id != ?", excludeUserID[0])
	}
	err := query.Count(&count).Error
	return count > 0, err
}

// ExistsEmailIncludingDeleted checks if email exists including soft-deleted.
func (r *UserRepository) ExistsEmailIncludingDeleted(email string) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Unscoped().Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// HardDeleteSoftDeletedUserByEmail permanently deletes a soft-deleted user by email.
func (r *UserRepository) HardDeleteSoftDeletedUserByEmail(email string) error {
	return r.db.Unscoped().Where("email = ? AND deleted_at IS NOT NULL", email).Delete(&model.User{}).Error
}

// IncrementLoginFailCount increments login fail count and returns updated user.
func (r *UserRepository) IncrementLoginFailCount(id int64) (*model.User, error) {
	now := time.Now()
	updates := map[string]interface{}{
		"login_fail_count":   gorm.Expr("login_fail_count + 1"),
		"last_login_fail_at": now,
	}

	if err := r.db.Model(&model.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}

	return r.FindByID(id)
}

// ClearLoginFailCount resets login fail count and unlock status.
func (r *UserRepository) ClearLoginFailCount(id int64) error {
	updates := map[string]interface{}{
		"login_fail_count":   0,
		"locked_until":       nil,
		"last_login_fail_at": nil,
	}
	return r.db.Model(&model.User{}).Where("id = ?", id).Updates(updates).Error
}

// LockAccount locks the account until a specified time.
func (r *UserRepository) LockAccount(id int64, lockedUntil time.Time) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Update("locked_until", lockedUntil).Error
}
