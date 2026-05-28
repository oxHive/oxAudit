# Action Plan - AWS Cost Optimization

## Phase 1: Identify and Terminate Unused Services

**Goal:** Reduce active service count before investing migration effort.

### Step 1 — Classify services by activity (all three signals required)

| Signal | Inactive threshold |
|--------|-------------------|
| Traffic (last 30 days) | 0 requests OR < 10 req/day average |
| Last git commit | > 90 days ago |
| Last pipeline run (dev + master) | > 90 days ago |

- **Archive candidate:** All three signals show inactive
- **Review candidate:** Two of three signals show inactive → manual decision
- **Keep:** One or fewer signals show inactive

### Step 2 — Execute terminations

1. Tag archive candidates with `status=decommission-pending`
2. Notify team, hold 2 weeks for objections
3. Export final snapshots / DB dumps to S3 (Glacier tier)
4. Terminate Elastic Beanstalk environments and associated RDS Table, ElastiCache, SQS resources

**Exit criteria for Phase 2:** Active service list finalized, zero unresolved terminations.

---

## Phase 2: Migrate Active Services — Elastic Beanstalk → ECS

**Timeline:** Estimate 2 sprints per service batch of 3 for each environment; adjust after Phase 1 count is known.

### Step 3 — Containerize each service

- Dockerfile + `.dockerignore` per service
- Health-check endpoint confirmed at `/api/health`
- Secrets migrated from EB env vars → AWS Secrets Manager / Parameter Store

### Step 4 — ECS task + service definitions

- Fargate launch type (eliminates EC2 instance management)
- Set `desiredCount=0` for all non-production (dev/staging) services
- Define autoscaling policy: scale up on CPU > 60% or request queue depth > 50

### Step 5 — Internal load balancer audit

For each ALB/NLB:

- If only one target service and traffic < 500 req/day → remove LB, use direct service discovery (AWS Cloud Map)
- If shared across services → keep, review listener rules for consolidation
- Document decision in `lb-audit.md`

---

## Phase 3: Validate Savings

- Pull Cost Explorer report 30 days after Phase 2 completes
- Compare against pre-audit baseline (tag: `cost-baseline-2026-05`)
- Target: ≥ 30% reduction in EC2 + ELB line items
- If target missed: review remaining EB environments and over-provisioned ECS task sizes
