package db

// migrations is a list of DDL batches applied in order.
// Index 0 = version 1. Each entry runs as a single transaction.
var migrations = []string{
	// Version 1 — initial schema
	`
CREATE TABLE IF NOT EXISTS audit_run (
    id              TEXT PRIMARY KEY,
    executed_at     TEXT NOT NULL,
    generated_at    TEXT NOT NULL,
    period_start    TEXT NOT NULL,
    period_end      TEXT NOT NULL,
    aws_profile     TEXT NOT NULL,
    billing_region  TEXT NOT NULL DEFAULT '',
    run_folder      TEXT NOT NULL DEFAULT '',
    config_json     TEXT,
    notes           TEXT,
    status          TEXT NOT NULL DEFAULT 'running'
);

CREATE TABLE IF NOT EXISTS raw_file (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    audit_run_id    TEXT NOT NULL REFERENCES audit_run(id),
    command_name    TEXT NOT NULL,
    service         TEXT NOT NULL DEFAULT '',
    region          TEXT NOT NULL DEFAULT '',
    file_path       TEXT NOT NULL,
    checksum        TEXT,
    bytes_written   INTEGER NOT NULL DEFAULT 0,
    collected_at    TEXT NOT NULL,
    duration_ms     INTEGER NOT NULL DEFAULT 0,
    status          TEXT NOT NULL DEFAULT 'ok',
    error_msg       TEXT NOT NULL DEFAULT '',
    required        INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS accounts (
    account_id      TEXT NOT NULL,
    account_name    TEXT NOT NULL DEFAULT '',
    email           TEXT NOT NULL DEFAULT '',
    environment     TEXT NOT NULL DEFAULT '',
    owner           TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL DEFAULT 'ACTIVE',
    audit_run_id    TEXT NOT NULL REFERENCES audit_run(id),
    PRIMARY KEY (account_id, audit_run_id)
);

CREATE TABLE IF NOT EXISTS cost_monthly (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    audit_run_id    TEXT NOT NULL REFERENCES audit_run(id),
    month           TEXT NOT NULL,
    account_id      TEXT NOT NULL DEFAULT '',
    account_name    TEXT NOT NULL DEFAULT '',
    service         TEXT NOT NULL DEFAULT '',
    region          TEXT NOT NULL DEFAULT '',
    usage_type      TEXT NOT NULL DEFAULT '',
    operation       TEXT NOT NULL DEFAULT '',
    tag_key         TEXT NOT NULL DEFAULT '',
    tag_value       TEXT NOT NULL DEFAULT '',
    unblended_cost  REAL NOT NULL DEFAULT 0,
    amortized_cost  REAL NOT NULL DEFAULT 0,
    usage_quantity  REAL NOT NULL DEFAULT 0,
    unit            TEXT NOT NULL DEFAULT 'USD',
    source_file     TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_cost_monthly_run_service
    ON cost_monthly(audit_run_id, service, month);
CREATE INDEX IF NOT EXISTS idx_cost_monthly_run_account
    ON cost_monthly(audit_run_id, account_id, month);

CREATE TABLE IF NOT EXISTS cost_daily (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    audit_run_id    TEXT NOT NULL REFERENCES audit_run(id),
    date            TEXT NOT NULL,
    account_id      TEXT NOT NULL DEFAULT '',
    service         TEXT NOT NULL DEFAULT '',
    region          TEXT NOT NULL DEFAULT '',
    unblended_cost  REAL NOT NULL DEFAULT 0,
    amortized_cost  REAL NOT NULL DEFAULT 0,
    source_file     TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_cost_daily_run_date
    ON cost_daily(audit_run_id, date, service);

CREATE TABLE IF NOT EXISTS resources (
    pk               INTEGER PRIMARY KEY AUTOINCREMENT,
    audit_run_id     TEXT NOT NULL REFERENCES audit_run(id),
    resource_id      TEXT NOT NULL,
    resource_type    TEXT NOT NULL,
    account_id       TEXT NOT NULL DEFAULT '',
    account_name     TEXT NOT NULL DEFAULT '',
    region           TEXT NOT NULL DEFAULT '',
    service          TEXT NOT NULL DEFAULT '',
    state            TEXT NOT NULL DEFAULT '',
    name             TEXT NOT NULL DEFAULT '',
    arn              TEXT NOT NULL DEFAULT '',
    created_at       TEXT,
    size_gib         REAL,
    instance_type    TEXT NOT NULL DEFAULT '',
    volume_type      TEXT NOT NULL DEFAULT '',
    iops             INTEGER,
    est_monthly_cost REAL,
    tags_json        TEXT NOT NULL DEFAULT '{}',
    raw_json         TEXT NOT NULL DEFAULT '{}',
    source_file      TEXT NOT NULL DEFAULT '',
    discovered_at    TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_resources_run_id_type
    ON resources(audit_run_id, resource_id, resource_type);
CREATE INDEX IF NOT EXISTS idx_resources_run_type_state
    ON resources(audit_run_id, resource_type, state);
CREATE INDEX IF NOT EXISTS idx_resources_run_account
    ON resources(audit_run_id, account_id);

CREATE TABLE IF NOT EXISTS recommendations (
    id                        TEXT PRIMARY KEY,
    audit_run_id              TEXT NOT NULL REFERENCES audit_run(id),
    rec_source                TEXT NOT NULL DEFAULT '',
    service                   TEXT NOT NULL DEFAULT '',
    account_id                TEXT NOT NULL DEFAULT '',
    region                    TEXT NOT NULL DEFAULT '',
    resource_id               TEXT NOT NULL DEFAULT '',
    current_resource_type     TEXT NOT NULL DEFAULT '',
    recommended_resource_type TEXT NOT NULL DEFAULT '',
    finding_text              TEXT NOT NULL DEFAULT '',
    est_monthly_savings_usd   REAL NOT NULL DEFAULT 0,
    confidence                TEXT NOT NULL DEFAULT '',
    raw_json                  TEXT NOT NULL DEFAULT '{}',
    source_file               TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS findings (
    id                      TEXT PRIMARY KEY,
    audit_run_id            TEXT NOT NULL REFERENCES audit_run(id),
    priority                TEXT NOT NULL,
    category                TEXT NOT NULL DEFAULT '',
    service                 TEXT NOT NULL DEFAULT '',
    account_id              TEXT NOT NULL DEFAULT '',
    account_name            TEXT NOT NULL DEFAULT '',
    region                  TEXT NOT NULL DEFAULT '',
    title                   TEXT NOT NULL DEFAULT '',
    summary                 TEXT NOT NULL DEFAULT '',
    evidence                TEXT NOT NULL DEFAULT '',
    recommended_action      TEXT NOT NULL DEFAULT '',
    est_monthly_savings_usd REAL NOT NULL DEFAULT 0,
    confidence              TEXT NOT NULL DEFAULT '',
    risk                    TEXT NOT NULL DEFAULT '',
    owner                   TEXT NOT NULL DEFAULT '',
    status                  TEXT NOT NULL DEFAULT 'open',
    resource_ids_json       TEXT NOT NULL DEFAULT '[]',
    tags_json               TEXT NOT NULL DEFAULT '{}',
    source_files_json       TEXT NOT NULL DEFAULT '[]',
    created_at              TEXT NOT NULL,
    updated_at              TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_findings_run_priority
    ON findings(audit_run_id, priority);
CREATE INDEX IF NOT EXISTS idx_findings_run_service
    ON findings(audit_run_id, service);

CREATE TABLE IF NOT EXISTS finding_resources (
    finding_id  TEXT NOT NULL REFERENCES findings(id),
    resource_pk INTEGER NOT NULL REFERENCES resources(pk),
    resource_id TEXT NOT NULL DEFAULT '',
    account_id  TEXT NOT NULL DEFAULT '',
    region      TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (finding_id, resource_pk)
);
`,
}
