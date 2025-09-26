package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/kirimku/smartseller-backend/internal/domain/errors"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/tenant"
)

// QueryBuilder handles dynamic query construction based on tenant type
type QueryBuilder interface {
	Select(columns ...string) QueryBuilder
	From(table string) QueryBuilder
	Where(condition string, args ...interface{}) QueryBuilder
	TenantWhere(storefrontID uuid.UUID) QueryBuilder
	Join(joinType, table, condition string) QueryBuilder
	LeftJoin(table, condition string) QueryBuilder
	InnerJoin(table, condition string) QueryBuilder
	OrderBy(column, direction string) QueryBuilder
	GroupBy(columns ...string) QueryBuilder
	Having(condition string, args ...interface{}) QueryBuilder
	Limit(limit int) QueryBuilder
	Offset(offset int) QueryBuilder
	Build() (query string, args []interface{})
	BuildCount() (query string, args []interface{})
	Reset() QueryBuilder
}

// queryBuilder is the concrete implementation of QueryBuilder
type queryBuilder struct {
	tenantCtx   *tenant.TenantContext
	selectCols  []string
	fromTable   string
	whereConds  []whereCondition
	joins       []joinClause
	orderByCols []orderClause
	groupByCols []string
	havingConds []whereCondition
	limitVal    *int
	offsetVal   *int
	argCounter  int
	args        []interface{}
}

type whereCondition struct {
	condition string
	args      []interface{}
}

type joinClause struct {
	joinType  string
	table     string
	condition string
}

type orderClause struct {
	column    string
	direction string
}

// NewQueryBuilder creates a new query builder with tenant context
func NewQueryBuilder(tenantCtx *tenant.TenantContext) QueryBuilder {
	return &queryBuilder{
		tenantCtx:   tenantCtx,
		selectCols:  make([]string, 0),
		whereConds:  make([]whereCondition, 0),
		joins:       make([]joinClause, 0),
		orderByCols: make([]orderClause, 0),
		groupByCols: make([]string, 0),
		havingConds: make([]whereCondition, 0),
		args:        make([]interface{}, 0),
		argCounter:  0,
	}
}

func (qb *queryBuilder) Select(columns ...string) QueryBuilder {
	qb.selectCols = append(qb.selectCols, columns...)
	return qb
}

func (qb *queryBuilder) From(table string) QueryBuilder {
	// Add schema prefix for schema-based tenancy
	switch qb.tenantCtx.TenantType {
	case tenant.TenantTypeSchema:
		qb.fromTable = fmt.Sprintf("tenant_%s.%s", qb.tenantCtx.StorefrontID.String(), table)
	default:
		qb.fromTable = table
	}
	return qb
}

func (qb *queryBuilder) Where(condition string, args ...interface{}) QueryBuilder {
	adjustedCondition := qb.adjustPlaceholders(condition, len(args))
	qb.whereConds = append(qb.whereConds, whereCondition{
		condition: adjustedCondition,
		args:      args,
	})
	qb.args = append(qb.args, args...)
	return qb
}

func (qb *queryBuilder) TenantWhere(storefrontID uuid.UUID) QueryBuilder {
	// Only add storefront_id filter for shared database
	if qb.tenantCtx.TenantType == tenant.TenantTypeShared {
		qb.argCounter++
		placeholder := fmt.Sprintf("$%d", qb.argCounter)
		qb.whereConds = append(qb.whereConds, whereCondition{
			condition: "storefront_id = " + placeholder,
			args:      []interface{}{},
		})
		qb.args = append(qb.args, storefrontID)
	}
	return qb
}

func (qb *queryBuilder) Join(joinType, table, condition string) QueryBuilder {
	// Add schema prefix for joins too
	switch qb.tenantCtx.TenantType {
	case tenant.TenantTypeSchema:
		table = fmt.Sprintf("tenant_%s.%s", qb.tenantCtx.StorefrontID.String(), table)
	}

	qb.joins = append(qb.joins, joinClause{
		joinType:  joinType,
		table:     table,
		condition: condition,
	})
	return qb
}

func (qb *queryBuilder) LeftJoin(table, condition string) QueryBuilder {
	return qb.Join("LEFT", table, condition)
}

func (qb *queryBuilder) InnerJoin(table, condition string) QueryBuilder {
	return qb.Join("INNER", table, condition)
}

func (qb *queryBuilder) OrderBy(column, direction string) QueryBuilder {
	if direction != "ASC" && direction != "DESC" {
		direction = "ASC"
	}
	qb.orderByCols = append(qb.orderByCols, orderClause{
		column:    column,
		direction: direction,
	})
	return qb
}

func (qb *queryBuilder) GroupBy(columns ...string) QueryBuilder {
	qb.groupByCols = append(qb.groupByCols, columns...)
	return qb
}

func (qb *queryBuilder) Having(condition string, args ...interface{}) QueryBuilder {
	adjustedCondition := qb.adjustPlaceholders(condition, len(args))
	qb.havingConds = append(qb.havingConds, whereCondition{
		condition: adjustedCondition,
		args:      args,
	})
	qb.args = append(qb.args, args...)
	return qb
}

func (qb *queryBuilder) Limit(limit int) QueryBuilder {
	if limit > 0 {
		qb.limitVal = &limit
	}
	return qb
}

func (qb *queryBuilder) Offset(offset int) QueryBuilder {
	if offset >= 0 {
		qb.offsetVal = &offset
	}
	return qb
}

func (qb *queryBuilder) Build() (string, []interface{}) {
	var query strings.Builder

	// SELECT clause
	query.WriteString("SELECT ")
	if len(qb.selectCols) > 0 {
		query.WriteString(strings.Join(qb.selectCols, ", "))
	} else {
		query.WriteString("*")
	}

	// FROM clause
	if qb.fromTable != "" {
		query.WriteString(" FROM ")
		query.WriteString(qb.fromTable)
	}

	// JOIN clauses
	for _, join := range qb.joins {
		query.WriteString(fmt.Sprintf(" %s JOIN %s ON %s", join.joinType, join.table, join.condition))
	}

	// WHERE clause
	if len(qb.whereConds) > 0 {
		query.WriteString(" WHERE ")
		conditions := make([]string, len(qb.whereConds))
		for i, cond := range qb.whereConds {
			conditions[i] = cond.condition
		}
		query.WriteString(strings.Join(conditions, " AND "))
	}

	// GROUP BY clause
	if len(qb.groupByCols) > 0 {
		query.WriteString(" GROUP BY ")
		query.WriteString(strings.Join(qb.groupByCols, ", "))
	}

	// HAVING clause
	if len(qb.havingConds) > 0 {
		query.WriteString(" HAVING ")
		conditions := make([]string, len(qb.havingConds))
		for i, cond := range qb.havingConds {
			conditions[i] = cond.condition
		}
		query.WriteString(strings.Join(conditions, " AND "))
	}

	// ORDER BY clause
	if len(qb.orderByCols) > 0 {
		query.WriteString(" ORDER BY ")
		orderClauses := make([]string, len(qb.orderByCols))
		for i, order := range qb.orderByCols {
			orderClauses[i] = fmt.Sprintf("%s %s", order.column, order.direction)
		}
		query.WriteString(strings.Join(orderClauses, ", "))
	}

	// LIMIT clause
	if qb.limitVal != nil {
		query.WriteString(fmt.Sprintf(" LIMIT %d", *qb.limitVal))
	}

	// OFFSET clause
	if qb.offsetVal != nil {
		query.WriteString(fmt.Sprintf(" OFFSET %d", *qb.offsetVal))
	}

	return query.String(), qb.args
}

func (qb *queryBuilder) BuildCount() (string, []interface{}) {
	var query strings.Builder

	// SELECT COUNT(*) clause
	query.WriteString("SELECT COUNT(*)")

	// FROM clause
	if qb.fromTable != "" {
		query.WriteString(" FROM ")
		query.WriteString(qb.fromTable)
	}

	// JOIN clauses
	for _, join := range qb.joins {
		query.WriteString(fmt.Sprintf(" %s JOIN %s ON %s", join.joinType, join.table, join.condition))
	}

	// WHERE clause
	if len(qb.whereConds) > 0 {
		query.WriteString(" WHERE ")
		conditions := make([]string, len(qb.whereConds))
		for i, cond := range qb.whereConds {
			conditions[i] = cond.condition
		}
		query.WriteString(strings.Join(conditions, " AND "))
	}

	// GROUP BY clause (for count with group by, we need different handling)
	if len(qb.groupByCols) > 0 {
		// Wrap the query with group by in a subquery
		groupQuery := strings.Builder{}
		groupQuery.WriteString("SELECT COUNT(*) FROM (SELECT 1")

		if qb.fromTable != "" {
			groupQuery.WriteString(" FROM ")
			groupQuery.WriteString(qb.fromTable)
		}

		for _, join := range qb.joins {
			groupQuery.WriteString(fmt.Sprintf(" %s JOIN %s ON %s", join.joinType, join.table, join.condition))
		}

		if len(qb.whereConds) > 0 {
			groupQuery.WriteString(" WHERE ")
			conditions := make([]string, len(qb.whereConds))
			for i, cond := range qb.whereConds {
				conditions[i] = cond.condition
			}
			groupQuery.WriteString(strings.Join(conditions, " AND "))
		}

		groupQuery.WriteString(" GROUP BY ")
		groupQuery.WriteString(strings.Join(qb.groupByCols, ", "))

		// HAVING clause
		if len(qb.havingConds) > 0 {
			groupQuery.WriteString(" HAVING ")
			conditions := make([]string, len(qb.havingConds))
			for i, cond := range qb.havingConds {
				conditions[i] = cond.condition
			}
			groupQuery.WriteString(strings.Join(conditions, " AND "))
		}

		groupQuery.WriteString(") AS grouped_query")
		return groupQuery.String(), qb.args
	}

	// HAVING clause (only if there's GROUP BY)
	if len(qb.havingConds) > 0 && len(qb.groupByCols) > 0 {
		query.WriteString(" HAVING ")
		conditions := make([]string, len(qb.havingConds))
		for i, cond := range qb.havingConds {
			conditions[i] = cond.condition
		}
		query.WriteString(strings.Join(conditions, " AND "))
	}

	return query.String(), qb.args
}

func (qb *queryBuilder) Reset() QueryBuilder {
	qb.selectCols = make([]string, 0)
	qb.fromTable = ""
	qb.whereConds = make([]whereCondition, 0)
	qb.joins = make([]joinClause, 0)
	qb.orderByCols = make([]orderClause, 0)
	qb.groupByCols = make([]string, 0)
	qb.havingConds = make([]whereCondition, 0)
	qb.limitVal = nil
	qb.offsetVal = nil
	qb.args = make([]interface{}, 0)
	qb.argCounter = 0
	return qb
}

func (qb *queryBuilder) adjustPlaceholders(condition string, argCount int) string {
	result := condition
	for i := 1; i <= argCount; i++ {
		oldPlaceholder := fmt.Sprintf("$%d", i)
		newPlaceholder := fmt.Sprintf("$%d", qb.argCounter+i)
		result = strings.ReplaceAll(result, oldPlaceholder, newPlaceholder)
	}
	qb.argCounter += argCount
	return result
}

// BaseRepository provides common functionality for all tenant-aware repositories
type BaseRepository struct {
	db             *sqlx.DB
	tenantResolver tenant.TenantResolver
}

// NewBaseRepository creates a new base repository instance
func NewBaseRepository(db *sqlx.DB, tenantResolver tenant.TenantResolver) *BaseRepository {
	return &BaseRepository{
		db:             db,
		tenantResolver: tenantResolver,
	}
}

// GetDB returns the appropriate database connection for the tenant
func (br *BaseRepository) GetDB(ctx context.Context, storefrontID uuid.UUID) (*sqlx.DB, error) {
	db, err := br.tenantResolver.GetDatabaseConnection(ctx, storefrontID)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	// Convert *sql.DB to *sqlx.DB
	return sqlx.NewDb(db, "postgres"), nil
}

// NewQueryBuilder creates a new query builder with tenant context
func (br *BaseRepository) NewQueryBuilder(ctx context.Context, storefrontID uuid.UUID) (QueryBuilder, error) {
	// Get storefront to create tenant context
	storefront, err := br.tenantResolver.GetStorefrontBySlug(ctx, "")
	if err != nil {
		// If we can't get storefront by slug, create a minimal context
		tenantType, err := br.tenantResolver.GetTenantType(ctx, storefrontID)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve tenant type: %w", err)
		}

		tenantCtx := &tenant.TenantContext{
			StorefrontID: storefrontID,
			TenantType:   tenantType,
		}
		return NewQueryBuilder(tenantCtx), nil
	}

	tenantCtx := br.tenantResolver.CreateTenantContext(storefront)
	return NewQueryBuilder(tenantCtx), nil
}

// GetTenantContext retrieves tenant context from the resolver
func (br *BaseRepository) GetTenantContext(ctx context.Context, storefrontID uuid.UUID) (*tenant.TenantContext, error) {
	// Try to get from context first (if set by middleware)
	if tenantCtx := getTenantContextFromRequest(ctx); tenantCtx != nil && tenantCtx.StorefrontID == storefrontID {
		return tenantCtx, nil
	}

	// Resolve tenant type
	tenantType, err := br.tenantResolver.GetTenantType(ctx, storefrontID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve tenant type: %w", err)
	}

	return &tenant.TenantContext{
		StorefrontID: storefrontID,
		TenantType:   tenantType,
	}, nil
}

// ExecuteInTransaction executes a function within a database transaction
func (br *BaseRepository) ExecuteInTransaction(ctx context.Context, storefrontID uuid.UUID, fn func(*sqlx.Tx) error) error {
	db, err := br.GetDB(ctx, storefrontID)
	if err != nil {
		return err
	}

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}

// ValidateStorefrontAccess checks if the storefront exists and is accessible
func (br *BaseRepository) ValidateStorefrontAccess(ctx context.Context, storefrontID uuid.UUID) error {
	// This could be cached for performance
	query := `SELECT id FROM storefronts WHERE id = $1 AND deleted_at IS NULL`

	var id uuid.UUID
	err := br.db.GetContext(ctx, &id, query, storefrontID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.ErrStorefrontNotFound
		}
		return fmt.Errorf("failed to validate storefront access: %w", err)
	}

	return nil
}

// Helper functions

// getTenantContextFromRequest retrieves tenant context from request context (set by middleware)
func getTenantContextFromRequest(ctx context.Context) *tenant.TenantContext {
	if tenantCtx, ok := ctx.Value("tenant_context").(*tenant.TenantContext); ok {
		return tenantCtx
	}
	return nil
}

// PaginationHelper provides utilities for pagination
type PaginationHelper struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// GetOffset calculates the offset for SQL queries
func (p *PaginationHelper) GetOffset() int {
	if p.Page <= 1 {
		return 0
	}
	return (p.Page - 1) * p.PageSize
}

// GetLimit returns the limit for SQL queries
func (p *PaginationHelper) GetLimit() int {
	if p.PageSize <= 0 {
		return 20 // Default page size
	}
	if p.PageSize > 100 {
		return 100 // Max page size
	}
	return p.PageSize
}

// CalculateTotalPages calculates total pages from total count
func (p *PaginationHelper) CalculateTotalPages(totalCount int) int {
	if totalCount == 0 || p.PageSize <= 0 {
		return 0
	}
	return (totalCount + p.PageSize - 1) / p.PageSize
}

// ValidatePagination validates and normalizes pagination parameters
func ValidatePagination(page, pageSize int) *PaginationHelper {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return &PaginationHelper{
		Page:     page,
		PageSize: pageSize,
	}
}

// MetricsCollector interface for collecting repository metrics
type MetricsCollector interface {
	RecordQuery(operation string, table string, duration time.Duration, err error)
	RecordTenantAccess(storefrontID uuid.UUID, tenantType tenant.TenantType)
}

// NoOpMetricsCollector is a no-op implementation for when metrics are disabled
type NoOpMetricsCollector struct{}

func (n *NoOpMetricsCollector) RecordQuery(operation string, table string, duration time.Duration, err error) {
}
func (n *NoOpMetricsCollector) RecordTenantAccess(storefrontID uuid.UUID, tenantType tenant.TenantType) {
}

// WithMetrics wraps repository operations with metrics collection
func WithMetrics(collector MetricsCollector, operation, table string, fn func() error) error {
	start := time.Now()
	err := fn()
	duration := time.Since(start)
	collector.RecordQuery(operation, table, duration, err)
	return err
}
