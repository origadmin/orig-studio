package dal

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/oklog/ulid/v2"

	"origadmin/application/origstudio/internal/data/entity/schema"
	"origadmin/application/origstudio/internal/enterprise/audit/dto"
)

type AuditService struct {
	db   *sql.DB
	ch   chan *dto.Entry
	done chan struct{}
	once sync.Once
	log  *log.Helper
}

func NewAuditService(db *sql.DB, logger log.Logger) *AuditService {
	svc := &AuditService{
		db:   db,
		ch:   make(chan *dto.Entry, 1024),
		done: make(chan struct{}),
		log:  log.NewHelper(log.With(logger, "module", "enterprise/audit")),
	}
	go svc.worker()
	return svc
}

func (s *AuditService) worker() {
	defer close(s.done)
	for entry := range s.ch {
		if err := s.writeEntry(context.Background(), entry); err != nil {
			s.log.Errorf("audit write failed: %v", err)
		}
	}
}

func (s *AuditService) writeEntry(ctx context.Context, entry *dto.Entry) error {
	id := ulid.Make().String()
	detailJSON := entry.Detail
	if detailJSON == "" {
		detailJSON = "{}"
	}
	query := fmt.Sprintf(
		`INSERT INTO %s (id, user_id, username, action, resource, resource_id, detail, ip, user_agent, result, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		schema.AuditLogTableName,
	)
	_, err := s.db.ExecContext(ctx, query, id, entry.UserID, entry.Username, entry.Action, entry.Resource, "", detailJSON, entry.IP, entry.UserAgent, entry.Result, time.Now())
	return err
}

func (s *AuditService) Log(ctx context.Context, entry *dto.Entry) error {
	entryCopy := *entry
	select {
	case s.ch <- &entryCopy:
		return nil
	default:
		go func() {
			s.ch <- &entryCopy
		}()
		return nil
	}
}

func (s *AuditService) LogSimple(ctx context.Context, userID, username, action, resource, ip, userAgent, result string) error {
	return s.Log(ctx, &dto.Entry{
		UserID:    userID,
		Username:  username,
		Action:    action,
		Resource:  resource,
		IP:        ip,
		UserAgent: userAgent,
		Result:    result,
	})
}

func (s *AuditService) LogSync(ctx context.Context, entry *dto.Entry) error {
	return s.writeEntry(ctx, entry)
}

func (s *AuditService) EnsureTable(ctx context.Context) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL DEFAULT '',
			username VARCHAR(150) NOT NULL DEFAULT '',
			action VARCHAR(50) NOT NULL DEFAULT '',
			resource VARCHAR(50) NOT NULL DEFAULT '',
			resource_id VARCHAR(36) NOT NULL DEFAULT '',
			detail TEXT NOT NULL DEFAULT '{}',
			ip VARCHAR(64) NOT NULL DEFAULT '',
			user_agent TEXT NOT NULL DEFAULT '',
			result VARCHAR(20) NOT NULL DEFAULT '',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`, schema.AuditLogTableName)
	_, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("ensure audit_logs table: %w", err)
	}

	indexQueries := []string{
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON %s (user_id)`, schema.AuditLogTableName),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON %s (action)`, schema.AuditLogTableName),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON %s (resource)`, schema.AuditLogTableName),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON %s (created_at)`, schema.AuditLogTableName),
	}
	for _, q := range indexQueries {
		if _, err := s.db.ExecContext(ctx, q); err != nil {
			s.log.Warnf("create audit index failed (non-fatal): %v", err)
		}
	}
	return nil
}

func (s *AuditService) Query(ctx context.Context, filter *dto.QueryFilter) (*dto.QueryResult, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	where := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if filter.UserID != "" {
		where += fmt.Sprintf(" AND user_id = $%d", argIdx)
		args = append(args, filter.UserID)
		argIdx++
	}
	if filter.Action != "" {
		where += fmt.Sprintf(" AND action = $%d", argIdx)
		args = append(args, filter.Action)
		argIdx++
	}
	if filter.Resource != "" {
		where += fmt.Sprintf(" AND resource = $%d", argIdx)
		args = append(args, filter.Resource)
		argIdx++
	}
	if filter.Result != "" {
		where += fmt.Sprintf(" AND result = $%d", argIdx)
		args = append(args, filter.Result)
		argIdx++
	}
	if filter.StartTime != "" {
		where += fmt.Sprintf(" AND created_at >= $%d", argIdx)
		args = append(args, filter.StartTime)
		argIdx++
	}
	if filter.EndTime != "" {
		where += fmt.Sprintf(" AND created_at <= $%d", argIdx)
		args = append(args, filter.EndTime)
		argIdx++
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", schema.AuditLogTableName, where)
	var total int64
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count audit logs: %w", err)
	}

	offset := (filter.Page - 1) * filter.PageSize
	selectQuery := fmt.Sprintf(
		"SELECT id, user_id, username, action, resource, resource_id, detail, ip, user_agent, result, created_at FROM %s %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		schema.AuditLogTableName, where, argIdx, argIdx+1,
	)
	args = append(args, filter.PageSize, offset)
	rows, err := s.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("query audit logs: %w", err)
	}
	defer rows.Close()

	var items []*dto.Entry
	for rows.Next() {
		var (
			id         string
			userID     string
			username   string
			action     string
			resource   string
			resourceID string
			detailStr  string
			ip         string
			userAgent  string
			result     string
			createdAt  time.Time
		)
		if err := rows.Scan(&id, &userID, &username, &action, &resource, &resourceID, &detailStr, &ip, &userAgent, &result, &createdAt); err != nil {
			return nil, fmt.Errorf("scan audit log: %w", err)
		}
		items = append(items, &dto.Entry{
			ID:         id,
			UserID:     userID,
			Username:   username,
			Action:     action,
			Resource:   resource,
			IP:         ip,
			UserAgent:  userAgent,
			Result:     result,
			Detail:     detailStr,
			CreateTime: createdAt,
		})
	}

	return &dto.QueryResult{
		Items:    items,
		Total:    int(total),
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

func (s *AuditService) Close() {
	s.once.Do(func() {
		close(s.ch)
		<-s.done
	})
}

func EntryDetailMap(e *dto.Entry) map[string]interface{} {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(e.Detail), &m); err != nil {
		return map[string]interface{}{}
	}
	return m
}

func EntrySetDetail(e *dto.Entry, v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		e.Detail = "{}"
		return
	}
	e.Detail = string(data)
}