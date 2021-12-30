// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package query

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"intery/server/model"

	"gorm.io/gen"
	"gorm.io/gen/field"
)

func newDeployment(db *gorm.DB) deployment {
	_deployment := deployment{}

	_deployment.deploymentDo.UseDB(db)
	_deployment.deploymentDo.UseModel(&model.Deployment{})

	tableName := _deployment.deploymentDo.TableName()
	_deployment.ALL = field.NewField(tableName, "*")
	_deployment.ID = field.NewUint(tableName, "id")
	_deployment.CreatedAt = field.NewTime(tableName, "created_at")
	_deployment.UpdatedAt = field.NewTime(tableName, "updated_at")
	_deployment.DeletedAt = field.NewField(tableName, "deleted_at")
	_deployment.ContainerId = field.NewUint(tableName, "container_id")
	_deployment.ProjectId = field.NewUint(tableName, "project_id")

	_deployment.fillFieldMap()

	return _deployment
}

type deployment struct {
	deploymentDo deploymentDo

	ALL         field.Field
	ID          field.Uint
	CreatedAt   field.Time
	UpdatedAt   field.Time
	DeletedAt   field.Field
	ContainerId field.Uint
	ProjectId   field.Uint

	fieldMap map[string]field.Expr
}

func (d deployment) As(alias string) *deployment {
	d.deploymentDo.DO = *(d.deploymentDo.As(alias).(*gen.DO))

	d.ALL = field.NewField(alias, "*")
	d.ID = field.NewUint(alias, "id")
	d.CreatedAt = field.NewTime(alias, "created_at")
	d.UpdatedAt = field.NewTime(alias, "updated_at")
	d.DeletedAt = field.NewField(alias, "deleted_at")
	d.ContainerId = field.NewUint(alias, "container_id")
	d.ProjectId = field.NewUint(alias, "project_id")

	d.fillFieldMap()

	return &d
}

func (d *deployment) WithContext(ctx context.Context) *deploymentDo {
	return d.deploymentDo.WithContext(ctx)
}

func (d deployment) TableName() string { return d.deploymentDo.TableName() }

func (d *deployment) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := d.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	return _f.(field.OrderExpr), true
}

func (d *deployment) fillFieldMap() {
	d.fieldMap = make(map[string]field.Expr, 6)
	d.fieldMap["id"] = d.ID
	d.fieldMap["created_at"] = d.CreatedAt
	d.fieldMap["updated_at"] = d.UpdatedAt
	d.fieldMap["deleted_at"] = d.DeletedAt
	d.fieldMap["container_id"] = d.ContainerId
	d.fieldMap["project_id"] = d.ProjectId
}

func (d deployment) clone(db *gorm.DB) deployment {
	d.deploymentDo.ReplaceDB(db)
	return d
}

type deploymentDo struct{ gen.DO }

func (d deploymentDo) Debug() *deploymentDo {
	return d.withDO(d.DO.Debug())
}

func (d deploymentDo) WithContext(ctx context.Context) *deploymentDo {
	return d.withDO(d.DO.WithContext(ctx))
}

func (d deploymentDo) Clauses(conds ...clause.Expression) *deploymentDo {
	return d.withDO(d.DO.Clauses(conds...))
}

func (d deploymentDo) Not(conds ...gen.Condition) *deploymentDo {
	return d.withDO(d.DO.Not(conds...))
}

func (d deploymentDo) Or(conds ...gen.Condition) *deploymentDo {
	return d.withDO(d.DO.Or(conds...))
}

func (d deploymentDo) Select(conds ...field.Expr) *deploymentDo {
	return d.withDO(d.DO.Select(conds...))
}

func (d deploymentDo) Where(conds ...gen.Condition) *deploymentDo {
	return d.withDO(d.DO.Where(conds...))
}

func (d deploymentDo) Order(conds ...field.Expr) *deploymentDo {
	return d.withDO(d.DO.Order(conds...))
}

func (d deploymentDo) Distinct(cols ...field.Expr) *deploymentDo {
	return d.withDO(d.DO.Distinct(cols...))
}

func (d deploymentDo) Omit(cols ...field.Expr) *deploymentDo {
	return d.withDO(d.DO.Omit(cols...))
}

func (d deploymentDo) Join(table schema.Tabler, on ...field.Expr) *deploymentDo {
	return d.withDO(d.DO.Join(table, on...))
}

func (d deploymentDo) LeftJoin(table schema.Tabler, on ...field.Expr) *deploymentDo {
	return d.withDO(d.DO.LeftJoin(table, on...))
}

func (d deploymentDo) RightJoin(table schema.Tabler, on ...field.Expr) *deploymentDo {
	return d.withDO(d.DO.RightJoin(table, on...))
}

func (d deploymentDo) Group(cols ...field.Expr) *deploymentDo {
	return d.withDO(d.DO.Group(cols...))
}

func (d deploymentDo) Having(conds ...gen.Condition) *deploymentDo {
	return d.withDO(d.DO.Having(conds...))
}

func (d deploymentDo) Limit(limit int) *deploymentDo {
	return d.withDO(d.DO.Limit(limit))
}

func (d deploymentDo) Offset(offset int) *deploymentDo {
	return d.withDO(d.DO.Offset(offset))
}

func (d deploymentDo) Scopes(funcs ...func(gen.Dao) gen.Dao) *deploymentDo {
	return d.withDO(d.DO.Scopes(funcs...))
}

func (d deploymentDo) Unscoped() *deploymentDo {
	return d.withDO(d.DO.Unscoped())
}

func (d deploymentDo) Create(values ...*model.Deployment) error {
	if len(values) == 0 {
		return nil
	}
	return d.DO.Create(values)
}

func (d deploymentDo) CreateInBatches(values []*model.Deployment, batchSize int) error {
	return d.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (d deploymentDo) Save(values ...*model.Deployment) error {
	if len(values) == 0 {
		return nil
	}
	return d.DO.Save(values)
}

func (d deploymentDo) First() (*model.Deployment, error) {
	if result, err := d.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*model.Deployment), nil
	}
}

func (d deploymentDo) Take() (*model.Deployment, error) {
	if result, err := d.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*model.Deployment), nil
	}
}

func (d deploymentDo) Last() (*model.Deployment, error) {
	if result, err := d.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*model.Deployment), nil
	}
}

func (d deploymentDo) Find() ([]*model.Deployment, error) {
	result, err := d.DO.Find()
	return result.([]*model.Deployment), err
}

func (d deploymentDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.Deployment, err error) {
	buf := make([]*model.Deployment, 0, batchSize)
	err = d.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (d deploymentDo) FindInBatches(result *[]*model.Deployment, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return d.DO.FindInBatches(result, batchSize, fc)
}

func (d deploymentDo) Attrs(attrs ...field.AssignExpr) *deploymentDo {
	return d.withDO(d.DO.Attrs(attrs...))
}

func (d deploymentDo) Assign(attrs ...field.AssignExpr) *deploymentDo {
	return d.withDO(d.DO.Assign(attrs...))
}

func (d deploymentDo) Joins(field field.RelationField) *deploymentDo {
	return d.withDO(d.DO.Joins(field))
}

func (d deploymentDo) Preload(field field.RelationField) *deploymentDo {
	return d.withDO(d.DO.Preload(field))
}

func (d deploymentDo) FirstOrInit() (*model.Deployment, error) {
	if result, err := d.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*model.Deployment), nil
	}
}

func (d deploymentDo) FirstOrCreate() (*model.Deployment, error) {
	if result, err := d.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*model.Deployment), nil
	}
}

func (d deploymentDo) FindByPage(offset int, limit int) (result []*model.Deployment, count int64, err error) {
	count, err = d.Count()
	if err != nil {
		return
	}

	if limit <= 0 {
		return
	}

	result, err = d.Offset(offset).Limit(limit).Find()
	return
}

func (d deploymentDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = d.Count()
	if err != nil {
		return
	}

	err = d.Offset(offset).Limit(limit).Scan(result)
	return
}

func (d *deploymentDo) withDO(do gen.Dao) *deploymentDo {
	d.DO = *do.(*gen.DO)
	return d
}
