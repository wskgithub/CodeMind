package repository

import (
	"codemind/internal/model"

	"gorm.io/gorm"
)

// UserRepository 用户数据访问层
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户 Repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create 创建用户
func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// FindByID 根据 ID 查找用户（包含部门信息）
func (r *UserRepository) FindByID(id int64) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Department").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByUsername 根据用户名查找用户
func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Department").Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail 根据邮箱查找用户
func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update 更新用户信息
func (r *UserRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

// UpdateFields 更新指定字段
func (r *UserRepository) UpdateFields(id int64, fields map[string]interface{}) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Updates(fields).Error
}

// Delete 软删除用户
func (r *UserRepository) Delete(id int64) error {
	return r.db.Delete(&model.User{}, id).Error
}

// List 分页查询用户列表
func (r *UserRepository) List(page, pageSize int, filters map[string]interface{}) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := r.db.Model(&model.User{}).Preload("Department")

	// 应用过滤条件
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

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// ListByDepartment 查询部门下的用户
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

// CountByDepartment 统计部门下的用户数
func (r *UserRepository) CountByDepartment(deptID int64) (int64, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("department_id = ?", deptID).Count(&count).Error
	return count, err
}

// CountAll 统计所有用户数
func (r *UserRepository) CountAll() (int64, error) {
	var count int64
	err := r.db.Model(&model.User{}).Count(&count).Error
	return count, err
}

// ExistsUsername 检查用户名是否已存在
func (r *UserRepository) ExistsUsername(username string) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

// ExistsEmail 检查邮箱是否已存在
func (r *UserRepository) ExistsEmail(email string, excludeUserID ...int64) (bool, error) {
	var count int64
	query := r.db.Model(&model.User{}).Where("email = ?", email)
	if len(excludeUserID) > 0 {
		query = query.Where("id != ?", excludeUserID[0])
	}
	err := query.Count(&count).Error
	return count > 0, err
}
