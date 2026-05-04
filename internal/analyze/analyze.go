package analyze

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/graditya/oxaudit/internal/analyze/rules"
	"github.com/graditya/oxaudit/internal/config"
	"github.com/graditya/oxaudit/internal/findutil"
)

// Engine runs all registered rules and upserts findings into the database.
type Engine struct {
	Rules []Rule
}

// New builds an Engine with the default MVP rule set.
func New(cfg *config.Config) *Engine {
	return &Engine{
		Rules: []Rule{
			rules.NewUnattachedEBSRule(cfg),
			rules.NewOldSnapshotsRule(cfg),
			rules.NewUnassociatedEIPRule(cfg),
			rules.NewStoppedEC2Rule(cfg),
			rules.NewMissingTagsRule(cfg),
			rules.NewHighNATCostRule(cfg),
			rules.NewHighCWCostRule(cfg),
			rules.NewCostAnomalyRule(cfg),
		},
	}
}

// Run iterates all rules, generates findings, and writes them to the database.
// Rule failures are logged but do not abort the run.
func (e *Engine) Run(ctx context.Context, db *sql.DB, cfg *config.Config, auditRunID string, onRule func(ruleID, ruleName string, count int, err error)) error {
	for _, rule := range e.Rules {
		findings, err := rule.Run(ctx, db, cfg, auditRunID)
		if onRule != nil {
			onRule(rule.ID(), rule.Name(), len(findings), err)
		}
		if err != nil {
			log.Printf("rule %s (%s) failed: %v", rule.ID(), rule.Name(), err)
			continue
		}
		for _, f := range findings {
			if err := findutil.UpsertFinding(ctx, db, f); err != nil {
				return fmt.Errorf("saving finding from rule %s: %w", rule.ID(), err)
			}
		}
	}
	return nil
}
