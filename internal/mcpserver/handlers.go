package mcpserver

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

func textOK(v interface{}) (CallResult, *rpcError) {
	b, _ := json.MarshalIndent(v, "", "  ")
	return CallResult{Content: []Content{{Type: "text", Text: string(b)}}}, nil
}

func textErr(msg string) (CallResult, *rpcError) {
	return CallResult{Content: []Content{{Type: "text", Text: msg}}, IsError: true}, nil
}

func strArg(args map[string]interface{}, key string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func intArg(args map[string]interface{}, key string, def int) int {
	if v, ok := args[key]; ok {
		if f, ok := v.(float64); ok {
			return int(f)
		}
	}
	return def
}

func floatArg(args map[string]interface{}, key string, def float64) float64 {
	if v, ok := args[key]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return def
}

func (s *Server) handleGetSummary(ctx context.Context, _ map[string]interface{}) (CallResult, *rpcError) {
	s.mu.RLock()
	runID := s.auditRunID
	s.mu.RUnlock()

	var periodStart, periodEnd, executedAt string
	s.db.QueryRowContext(ctx,
		`SELECT period_start, period_end, executed_at FROM audit_run WHERE id = ?`, runID,
	).Scan(&periodStart, &periodEnd, &executedAt)

	var total, p0, p1, p2, p3 int
	var totalSavings float64
	s.db.QueryRowContext(ctx, `
		SELECT COUNT(*),
			COALESCE(SUM(CASE WHEN priority='P0' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN priority='P1' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN priority='P2' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN priority='P3' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(est_monthly_savings_usd), 0)
		FROM findings WHERE audit_run_id = ? AND status = 'open'`, runID,
	).Scan(&total, &p0, &p1, &p2, &p3, &totalSavings)

	var accounts, regions int
	s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM accounts WHERE audit_run_id = ?`, runID).Scan(&accounts)
	s.db.QueryRowContext(ctx,
		`SELECT COUNT(DISTINCT region) FROM resources WHERE audit_run_id = ?`, runID).Scan(&regions)

	return textOK(map[string]interface{}{
		"audit_run_id":              runID,
		"period_start":              periodStart,
		"period_end":                periodEnd,
		"executed_at":               executedAt,
		"accounts":                  accounts,
		"regions":                   regions,
		"total_findings":            total,
		"p0":                        p0,
		"p1":                        p1,
		"p2":                        p2,
		"p3":                        p3,
		"estimated_monthly_savings": totalSavings,
	})
}

func (s *Server) handleListFindings(ctx context.Context, args map[string]interface{}) (CallResult, *rpcError) {
	s.mu.RLock()
	runID := s.auditRunID
	s.mu.RUnlock()

	priority := strArg(args, "priority")
	service := strArg(args, "service")
	category := strArg(args, "category")
	status := strArg(args, "status")
	if status == "" {
		status = "open"
	}
	minSavings := floatArg(args, "min_savings_usd", 0)
	limit := intArg(args, "limit", 50)

	query := `SELECT id, title, priority, category, service, region, account_name,
	                 est_monthly_savings_usd, confidence, risk, status
	          FROM findings WHERE audit_run_id = ? AND status = ?`
	queryArgs := []interface{}{runID, status}

	if priority != "" {
		query += " AND priority = ?"
		queryArgs = append(queryArgs, priority)
	}
	if service != "" {
		query += " AND service = ?"
		queryArgs = append(queryArgs, service)
	}
	if category != "" {
		query += " AND category = ?"
		queryArgs = append(queryArgs, category)
	}
	if minSavings > 0 {
		query += " AND est_monthly_savings_usd >= ?"
		queryArgs = append(queryArgs, minSavings)
	}
	query += " ORDER BY priority, est_monthly_savings_usd DESC LIMIT ?"
	queryArgs = append(queryArgs, limit)

	rows, err := s.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return textErr("query error: " + err.Error())
	}
	defer rows.Close()

	type finding struct {
		ID         string  `json:"id"`
		Title      string  `json:"title"`
		Priority   string  `json:"priority"`
		Category   string  `json:"category"`
		Service    string  `json:"service"`
		Region     string  `json:"region"`
		Account    string  `json:"account_name"`
		Savings    float64 `json:"est_monthly_savings_usd"`
		Confidence string  `json:"confidence"`
		Risk       string  `json:"risk"`
		Status     string  `json:"status"`
	}

	var findings []finding
	for rows.Next() {
		var f finding
		rows.Scan(&f.ID, &f.Title, &f.Priority, &f.Category, &f.Service,
			&f.Region, &f.Account, &f.Savings, &f.Confidence, &f.Risk, &f.Status)
		findings = append(findings, f)
	}
	if findings == nil {
		findings = []finding{}
	}
	return textOK(findings)
}

func (s *Server) handleGetFinding(ctx context.Context, args map[string]interface{}) (CallResult, *rpcError) {
	s.mu.RLock()
	runID := s.auditRunID
	s.mu.RUnlock()

	id := strArg(args, "finding_id")
	if id == "" {
		return textErr("finding_id is required")
	}

	type fullFinding struct {
		ID                string  `json:"id"`
		Priority          string  `json:"priority"`
		Category          string  `json:"category"`
		Service           string  `json:"service"`
		AccountID         string  `json:"account_id"`
		AccountName       string  `json:"account_name"`
		Region            string  `json:"region"`
		Title             string  `json:"title"`
		Summary           string  `json:"summary"`
		Evidence          string  `json:"evidence"`
		RecommendedAction string  `json:"recommended_action"`
		Savings           float64 `json:"est_monthly_savings_usd"`
		Confidence        string  `json:"confidence"`
		Risk              string  `json:"risk"`
		Owner             string  `json:"owner"`
		Status            string  `json:"status"`
		ResourceIDs       string  `json:"resource_ids"`
		SourceFiles       string  `json:"source_files"`
		CreatedAt         string  `json:"created_at"`
	}

	var f fullFinding
	err := s.db.QueryRowContext(ctx, `
		SELECT id, priority, category, service, account_id, account_name, region,
		       title, summary, evidence, recommended_action, est_monthly_savings_usd,
		       confidence, risk, owner, status, resource_ids_json, source_files_json, created_at
		FROM findings WHERE id = ? AND audit_run_id = ?`, id, runID,
	).Scan(&f.ID, &f.Priority, &f.Category, &f.Service, &f.AccountID, &f.AccountName,
		&f.Region, &f.Title, &f.Summary, &f.Evidence, &f.RecommendedAction,
		&f.Savings, &f.Confidence, &f.Risk, &f.Owner, &f.Status,
		&f.ResourceIDs, &f.SourceFiles, &f.CreatedAt)

	if err == sql.ErrNoRows {
		return textErr(fmt.Sprintf("finding %s not found in latest run", id))
	}
	if err != nil {
		return textErr("query error: " + err.Error())
	}
	return textOK(f)
}

func (s *Server) handleGetCostBreakdown(_ context.Context, _ map[string]interface{}) (CallResult, *rpcError) {
	return textErr("not yet implemented")
}

func (s *Server) handleQueryResources(_ context.Context, _ map[string]interface{}) (CallResult, *rpcError) {
	return textErr("not yet implemented")
}

func (s *Server) handleRunAudit(_ context.Context, _ map[string]interface{}) (CallResult, *rpcError) {
	return textErr("not yet implemented")
}
