package repository

import (
	"codemind/internal/model"

	"gorm.io/gorm"
)

// DepartmentRepository handles department data access
type DepartmentRepository struct {
	db *gorm.DB
}

// NewDepartmentRepository creates a department repository
func NewDepartmentRepository(db *gorm.DB) *DepartmentRepository {
	return &DepartmentRepository{db: db}
}

// Create creates a new department
func (r *DepartmentRepository) Create(dept *model.Department) error {
	return r.db.Create(dept).Error
}

// FindByID finds a department by ID
func (r *DepartmentRepository) FindByID(id int64) (*model.Department, error) {
	var dept model.Department
	err := r.db.Preload("Manager").First(&dept, id).Error
	if err != nil {
		return nil, err
	}
	return &dept, nil
}

// FindByName finds a department by name
func (r *DepartmentRepository) FindByName(name string) (*model.Department, error) {
	var dept model.Department
	err := r.db.Where("name = ?", name).First(&dept).Error
	if err != nil {
		return nil, err
	}
	return &dept, nil
}

// Update updates department information
func (r *DepartmentRepository) Update(dept *model.Department) error {
	return r.db.Save(dept).Error
}

// UpdateFields updates specific fields
func (r *DepartmentRepository) UpdateFields(id int64, fields map[string]interface{}) error {
	return r.db.Model(&model.Department{}).Where("id = ?", id).Updates(fields).Error
}

// Delete deletes a department
func (r *DepartmentRepository) Delete(id int64) error {
	return r.db.Delete(&model.Department{}, id).Error
}

// ListAll returns all departments
func (r *DepartmentRepository) ListAll() ([]model.Department, error) {
	var depts []model.Department
	err := r.db.Preload("Manager").Order("id ASC").Find(&depts).Error
	return depts, err
}

// ListByParentID returns child departments
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

// CountAll counts all departments
func (r *DepartmentRepository) CountAll() (int64, error) {
	var count int64
	err := r.db.Model(&model.Department{}).Count(&count).Error
	return count, err
}

// ExistsName checks if a department name exists
func (r *DepartmentRepository) ExistsName(name string, excludeID ...int64) (bool, error) {
	var count int64
	query := r.db.Model(&model.Department{}).Where("name = ?", name)
	if len(excludeID) > 0 {
		query = query.Where("id != ?", excludeID[0])
	}
	err := query.Count(&count).Error
	return count > 0, err
}

// HasChildren checks if department has children
func (r *DepartmentRepository) HasChildren(id int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.Department{}).Where("parent_id = ?", id).Count(&count).Error
	return count > 0, err
}
