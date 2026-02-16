package repository

import (
	"codemind/internal/model"

	"gorm.io/gorm"
)

// DepartmentRepository 部门数据访问层
type DepartmentRepository struct {
	db *gorm.DB
}

// NewDepartmentRepository 创建部门 Repository
func NewDepartmentRepository(db *gorm.DB) *DepartmentRepository {
	return &DepartmentRepository{db: db}
}

// Create 创建部门
func (r *DepartmentRepository) Create(dept *model.Department) error {
	return r.db.Create(dept).Error
}

// FindByID 根据 ID 查找部门
func (r *DepartmentRepository) FindByID(id int64) (*model.Department, error) {
	var dept model.Department
	err := r.db.Preload("Manager").First(&dept, id).Error
	if err != nil {
		return nil, err
	}
	return &dept, nil
}

// FindByName 根据名称查找部门
func (r *DepartmentRepository) FindByName(name string) (*model.Department, error) {
	var dept model.Department
	err := r.db.Where("name = ?", name).First(&dept).Error
	if err != nil {
		return nil, err
	}
	return &dept, nil
}

// Update 更新部门信息
func (r *DepartmentRepository) Update(dept *model.Department) error {
	return r.db.Save(dept).Error
}

// UpdateFields 更新指定字段
func (r *DepartmentRepository) UpdateFields(id int64, fields map[string]interface{}) error {
	return r.db.Model(&model.Department{}).Where("id = ?", id).Updates(fields).Error
}

// Delete 删除部门
func (r *DepartmentRepository) Delete(id int64) error {
	return r.db.Delete(&model.Department{}, id).Error
}

// ListAll 查询所有部门（树形结构构建在 Service 层完成）
func (r *DepartmentRepository) ListAll() ([]model.Department, error) {
	var depts []model.Department
	err := r.db.Preload("Manager").Order("id ASC").Find(&depts).Error
	return depts, err
}

// ListByParentID 查询子部门
func (r *DepartmentRepository) ListByParentID(parentID *int64) ([]model.Department, error) {
	var depts []model.Department
	query := r.db.Preload("Manager")
	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}
	err := query.Order("id ASC").Find(&depts).Error
	return depts, err
}

// CountAll 统计所有部门数
func (r *DepartmentRepository) CountAll() (int64, error) {
	var count int64
	err := r.db.Model(&model.Department{}).Count(&count).Error
	return count, err
}

// ExistsName 检查部门名称是否已存在
func (r *DepartmentRepository) ExistsName(name string, excludeID ...int64) (bool, error) {
	var count int64
	query := r.db.Model(&model.Department{}).Where("name = ?", name)
	if len(excludeID) > 0 {
		query = query.Where("id != ?", excludeID[0])
	}
	err := query.Count(&count).Error
	return count > 0, err
}

// HasChildren 检查部门是否有子部门
func (r *DepartmentRepository) HasChildren(id int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.Department{}).Where("parent_id = ?", id).Count(&count).Error
	return count > 0, err
}
