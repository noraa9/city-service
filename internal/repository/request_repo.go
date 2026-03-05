package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"city-service/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type RequestRepo struct {
	db *sqlx.DB
}

func NewRequestRepo(db *sqlx.DB) *RequestRepo {
	return &RequestRepo{db: db}
}

// dbRequestRow is a "flattened" row that includes joined data.
// We then map it into nested domain structs (Request.Category, Request.User, Request.Contractor).
type dbRequestRow struct {
	ID            uuid.UUID      `db:"id"`
	RequestNumber string         `db:"request_number"`
	Title         string         `db:"title"`
	CategoryID    sql.NullInt64  `db:"category_id"`
	Description   string         `db:"description"`
	Urgency       sql.NullString `db:"urgency"`
	Deadline      sql.NullTime   `db:"deadline"`
	Location      string         `db:"location"`
	PhotoURL      sql.NullString `db:"photo_url"`
	Status        string         `db:"status"`
	UserID        uuid.UUID      `db:"user_id"`
	ContractorID  uuid.NullUUID  `db:"contractor_id"`
	TakenAt       sql.NullTime   `db:"taken_at"`
	CreatedAt     time.Time      `db:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at"`

	// Category join (nullable because category_id is nullable)
	CategoryName sql.NullString `db:"category_name"`
	CategorySlug sql.NullString `db:"category_slug"`

	// Monitor (creator) join
	MonitorFullName sql.NullString `db:"monitor_full_name"`
	MonitorEmail    sql.NullString `db:"monitor_email"`
	MonitorPhone    sql.NullString `db:"monitor_phone"`
	MonitorRole     sql.NullString `db:"monitor_role"`

	// Contractor join (nullable)
	ContractorFullName          sql.NullString `db:"contractor_full_name"`
	ContractorEmail             sql.NullString `db:"contractor_email"`
	ContractorPhone             sql.NullString `db:"contractor_phone"`
	ContractorRole              sql.NullString `db:"contractor_role"`
	ContractorCompanyName       sql.NullString `db:"contractor_company_name"`
	ContractorResponsiblePerson sql.NullString `db:"contractor_responsible_person"`
	ContractorCompanyPhone      sql.NullString `db:"contractor_company_phone"`
}

const baseRequestSelect = `
	SELECT
		r.id, r.request_number, r.title, r.category_id, r.description, r.urgency,
		r.deadline, r.location, r.photo_url, r.status, r.user_id, r.contractor_id,
		r.taken_at, r.created_at, r.updated_at,

		c.name AS category_name,
		c.slug AS category_slug,

		m.full_name AS monitor_full_name,
		m.email     AS monitor_email,
		m.phone     AS monitor_phone,
		m.role      AS monitor_role,

		co.full_name          AS contractor_full_name,
		co.email              AS contractor_email,
		co.phone              AS contractor_phone,
		co.role               AS contractor_role,
		co.company_name       AS contractor_company_name,
		co.responsible_person AS contractor_responsible_person,
		co.company_phone      AS contractor_company_phone
	FROM requests r
	LEFT JOIN categories c ON c.id = r.category_id
	INNER JOIN users m ON m.id = r.user_id
	LEFT JOIN users co ON co.id = r.contractor_id
`

func (r *RequestRepo) CountAll(ctx context.Context) (int, error) {
	var c int
	if err := r.db.GetContext(ctx, &c, `SELECT COUNT(*) FROM requests`); err != nil {
		return 0, fmt.Errorf("count requests: %w", err)
	}
	return c, nil
}

func (r *RequestRepo) Create(ctx context.Context, req domain.Request) (domain.Request, error) {
	// We insert minimal fields; joined objects are loaded via GetByID below.
	q := `
		INSERT INTO requests (
			request_number, title, category_id, description, urgency, deadline,
			location, photo_url, status, user_id
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id, created_at, updated_at
	`

	var out struct {
		ID        uuid.UUID `db:"id"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}

	var deadline any = nil
	if req.Deadline != nil {
		deadline = *req.Deadline
	}

	var categoryID any = nil
	if req.CategoryID != 0 {
		categoryID = req.CategoryID
	}

	if err := r.db.GetContext(ctx, &out, q,
		req.RequestNumber,
		req.Title,
		categoryID,
		req.Description,
		nullString(req.Urgency),
		deadline,
		req.Location,
		nullStringPtr(req.PhotoURL),
		req.Status,
		req.UserID,
	); err != nil {
		return domain.Request{}, fmt.Errorf("insert request: %w", err)
	}

	// Return a fully populated request with joins for consistent API responses.
	return r.GetByID(ctx, out.ID)
}

func (r *RequestRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Request, error) {
	var row dbRequestRow
	q := baseRequestSelect + ` WHERE r.id = $1`
	if err := r.db.GetContext(ctx, &row, q, id); err != nil {
		if err == sql.ErrNoRows {
			return domain.Request{}, ErrNotFound
		}
		return domain.Request{}, fmt.Errorf("select request: %w", err)
	}
	return toDomainRequest(row), nil
}

func (r *RequestRepo) ListAll(ctx context.Context, f RequestFilters) ([]domain.Request, error) {
	q, args := buildRequestListQuery(baseRequestSelect, f, false, nil, nil)
	var rows []dbRequestRow
	if err := r.db.SelectContext(ctx, &rows, q, args...); err != nil {
		return nil, fmt.Errorf("list requests: %w", err)
	}
	return mapRequests(rows), nil
}

func (r *RequestRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Request, error) {
	// "My requests" is only by user_id; no other filters in spec.
	q := baseRequestSelect + ` WHERE r.user_id = $1 ORDER BY r.created_at DESC`
	var rows []dbRequestRow
	if err := r.db.SelectContext(ctx, &rows, q, userID); err != nil {
		return nil, fmt.Errorf("list my requests: %w", err)
	}
	return mapRequests(rows), nil
}

func (r *RequestRepo) ListNew(ctx context.Context, f RequestFilters) ([]domain.Request, error) {
	// Contractors list only "new" requests.
	status := domain.StatusNew
	f.Status = &status

	q, args := buildRequestListQuery(baseRequestSelect, f, false, nil, nil)
	var rows []dbRequestRow
	if err := r.db.SelectContext(ctx, &rows, q, args...); err != nil {
		return nil, fmt.Errorf("list new requests: %w", err)
	}
	return mapRequests(rows), nil
}

func (r *RequestRepo) ListByContractor(ctx context.Context, contractorID uuid.UUID) ([]domain.Request, error) {
	q := baseRequestSelect + ` WHERE r.contractor_id = $1 ORDER BY r.created_at DESC`
	var rows []dbRequestRow
	if err := r.db.SelectContext(ctx, &rows, q, contractorID); err != nil {
		return nil, fmt.Errorf("list contractor requests: %w", err)
	}
	return mapRequests(rows), nil
}

func (r *RequestRepo) Cancel(ctx context.Context, requestID uuid.UUID) error {
	// Business rule checks happen in service; repository only performs SQL.
	q := `UPDATE requests SET status = $1 WHERE id = $2`
	if _, err := r.db.ExecContext(ctx, q, domain.StatusCancelled, requestID); err != nil {
		return fmt.Errorf("cancel request: %w", err)
	}
	return nil
}

func (r *RequestRepo) AssignContractor(ctx context.Context, requestID uuid.UUID, contractorID uuid.UUID, takenAt time.Time) error {
	q := `
		UPDATE requests
		SET status = $1, contractor_id = $2, taken_at = $3
		WHERE id = $4
	`
	if _, err := r.db.ExecContext(ctx, q, domain.StatusInProgress, contractorID, takenAt, requestID); err != nil {
		return fmt.Errorf("assign contractor: %w", err)
	}
	return nil
}

func (r *RequestRepo) MarkDone(ctx context.Context, requestID uuid.UUID) error {
	q := `UPDATE requests SET status = $1 WHERE id = $2`
	if _, err := r.db.ExecContext(ctx, q, domain.StatusDone, requestID); err != nil {
		return fmt.Errorf("mark done: %w", err)
	}
	return nil
}

func mapRequests(rows []dbRequestRow) []domain.Request {
	out := make([]domain.Request, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainRequest(row))
	}
	return out
}

func toDomainRequest(r dbRequestRow) domain.Request {
	req := domain.Request{
		ID:            r.ID,
		RequestNumber: r.RequestNumber,
		Title:         r.Title,
		CategoryID:    int(r.CategoryID.Int64),
		Description:   r.Description,
		Urgency:       r.Urgency.String,
		Location:      r.Location,
		Status:        r.Status,
		UserID:        r.UserID,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}

	if r.CategoryID.Valid {
		req.Category = &domain.Category{
			ID:   int(r.CategoryID.Int64),
			Name: r.CategoryName.String,
			Slug: r.CategorySlug.String,
		}
	}

	if r.Deadline.Valid {
		t := r.Deadline.Time
		req.Deadline = &t
	}
	if r.PhotoURL.Valid {
		v := r.PhotoURL.String
		req.PhotoURL = &v
	}
	if r.ContractorID.Valid {
		id := r.ContractorID.UUID
		req.ContractorID = &id
	}
	if r.TakenAt.Valid {
		t := r.TakenAt.Time
		req.TakenAt = &t
	}

	// Monitor (creator) is always present (INNER JOIN).
	req.User = &domain.User{
		ID:       r.UserID,
		FullName: r.MonitorFullName.String,
		Email:    r.MonitorEmail.String,
		Phone:    r.MonitorPhone.String,
		Role:     r.MonitorRole.String,
	}

	// Contractor is optional (LEFT JOIN).
	if r.ContractorID.Valid {
		req.Contractor = &domain.User{
			ID:                r.ContractorID.UUID,
			FullName:          r.ContractorFullName.String,
			Email:             r.ContractorEmail.String,
			Phone:             r.ContractorPhone.String,
			Role:              r.ContractorRole.String,
			CompanyName:       stringPtr(r.ContractorCompanyName),
			ResponsiblePerson: stringPtr(r.ContractorResponsiblePerson),
			CompanyPhone:      stringPtr(r.ContractorCompanyPhone),
		}
	}

	return req
}

func buildRequestListQuery(base string, f RequestFilters, includeOrder bool, extraWhere []string, extraArgs []any) (string, []any) {
	where := make([]string, 0, 8)
	args := make([]any, 0, 8)

	// We build "WHERE ..." dynamically because filters are optional.
	add := func(cond string, val any) {
		args = append(args, val)
		where = append(where, fmt.Sprintf(cond, len(args)))
	}

	if f.Status != nil && *f.Status != "" {
		add("r.status = $%d", *f.Status)
	}
	if f.CategoryID != nil && *f.CategoryID != 0 {
		add("r.category_id = $%d", *f.CategoryID)
	}
	if f.Urgency != nil && *f.Urgency != "" {
		add("r.urgency = $%d", *f.Urgency)
	}
	if f.ContractorID != nil {
		add("r.contractor_id = $%d", *f.ContractorID)
	}
	if f.DateFrom != nil {
		add("r.created_at >= $%d", *f.DateFrom)
	}
	if f.DateTo != nil {
		add("r.created_at <= $%d", *f.DateTo)
	}

	where = append(where, extraWhere...)
	args = append(args, extraArgs...)

	sb := strings.Builder{}
	sb.WriteString(base)
	if len(where) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(where, " AND "))
	}
	sb.WriteString(" ORDER BY r.created_at DESC")

	return sb.String(), args
}

