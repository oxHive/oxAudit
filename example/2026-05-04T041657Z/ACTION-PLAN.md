# AWS Cost Audit — Action Plan
**Audit Run:** `aws-cost-audit-2026-04-01-2026-04-30-20260504T041657Z-default`
**Period:** 2026-04-01 to 2026-04-30
**Generated:** 2026-05-04
**Region:** ap-southeast-1

---

## Conclusion

**Total monthly spend: $721K | Estimated recoverable savings: $26,767/month (3.7%)**

April is significantly better than February–March ($1.18M). The massive CloudWatch spikes from March ($2,321/day) have partially subsided — April's worst day was $705. However, **several patterns from the previous audit remain unfixed**, indicating the earlier action plan was not fully executed. Additionally, two new concrete waste findings emerged: real EC2 instances sitting stopped in ap-southeast-1, one for 171 days.

### Comparison to Previous Audit (Feb–Apr)

| Metric | Feb–Apr | Apr Only | Change |
|--------|---------|----------|--------|
| Monthly Spend | $1,184,979 | $721,279 | ↓ 39% |
| Est. Savings | $236,082 | $26,767 | ↓ 89% |
| Total Findings | 43 | 17 | ↓ 60% |
| GuardDuty month-start spike | ✅ Found | ✅ Still happening | **NOT FIXED** |
| CloudTrail data event spike | ✅ Found | ✅ Still happening | **NOT FIXED** |
| ElastiCache extra node ($450/day) | ✅ Found | ✅ Still at $450/day | **NOT FIXED** |
| Stopped EC2 instances | — | ✅ 2 found (30d, 171d) | **NEW** |
| CloudWatch log costs | — | $12,544/month structural | **NEW** |

---

## Priority Order

| # | Action | Est. Savings | Effort | Do by |
|---|--------|-------------|--------|-------|
| 1 | Terminate `i-0d16f29483c77b603` (m6a.large, 171 days stopped) | EBS cost savings | 15 min | **Today** |
| 2 | Investigate CloudWatch spike on Apr 10 ($705 vs avg $432) | $8,191/mo | 1 hr | This week |
| 3 | Fix GuardDuty month-start spikes — scope S3 data events *(carried over)* | $9,887/mo | 1 hr | This week |
| 4 | Fix CloudTrail data events *(carried over from March)* | $2,171/mo | 30 min | This week |
| 5 | Investigate Apr 12 Config + Security Hub correlated spike | $852/mo | 30 min | This week |
| 6 | Investigate Apr 19 CloudTrail + Secrets Manager correlated spike | $1,475/mo | 30 min | This week |
| 7 | Set retention policies on all CloudWatch log groups | $5,017/mo | 2 hrs | This week |
| 8 | Terminate `i-07970379c328e00f1` (t3a.large, 30 days stopped) | EBS cost savings | 15 min | This week |
| 9 | Remove ElastiCache extra replica node *(carried over)* | ~$1,980/mo | 30 min | This week |
| 10 | Add VPC endpoints for S3/DynamoDB to reduce NAT cost | $169/mo | 2 hrs | Next week |

---

## Detailed Action Plan

### 1. Stopped EC2 Instance — URGENT: 171 Days (i-0d16f29483c77b603, m6a.large)

An `m6a.large` instance has been stopped since **November 2025** — 171 days. It is still accruing EBS volume charges every day it exists.

```bash
# Confirm instance details before terminating
aws ec2 describe-instances --region ap-southeast-1 \
  --instance-ids i-0d16f29483c77b603 \
  --query 'Reservations[].Instances[].[InstanceId,InstanceType,State.Name,LaunchTime,Tags]'

# Check what EBS volumes are attached
aws ec2 describe-volumes --region ap-southeast-1 \
  --filters Name=attachment.instance-id,Values=i-0d16f29483c77b603 \
  --query 'Volumes[].[VolumeId,Size,VolumeType,State]'

# Snapshot attached volumes before terminating (safety step)
aws ec2 create-snapshot --region ap-southeast-1 \
  --volume-id <volume-id> --description "pre-termination-backup-i-0d16f29483c77b603"

# Terminate after owner confirmation
aws ec2 terminate-instances --region ap-southeast-1 \
  --instance-ids i-0d16f29483c77b603
```

> **Note:** Stopped EC2 doesn't incur compute charges, but attached EBS volumes do. An m6a.large typically carries 50–200 GiB of EBS — that's $4–16/month in storage alone, plus any provisioned IOPS.

---

### 2. CloudWatch Spike — Apr 10 ($8,191/month)

CloudWatch cost hit $705.38 on April 10, which is 4.1 standard deviations above the April baseline of $432/day. This is a smaller spike than March's $2,321 but still statistically extreme and suggests the root cause from March was not fully resolved.

**Root cause hypothesis:** A deployment or scheduled job on Apr 10 again enabled high-volume logging. The baseline itself ($432/day) is already elevated compared to what a normal environment should cost — indicating there may be ongoing verbose logging even on "normal" days.

```bash
# Find log groups with highest ingestion around Apr 10
aws cloudwatch get-metric-statistics \
  --namespace AWS/Logs --metric-name IncomingBytes \
  --start-time 2026-04-09T00:00:00Z --end-time 2026-04-11T00:00:00Z \
  --period 3600 --statistics Sum --region ap-southeast-1

# Check which log groups had spikes
aws logs describe-log-groups --region ap-southeast-1 \
  --query 'sort_by(logGroups, &storedBytes)[-10:].[logGroupName,storedBytes,retentionInDays]' \
  --output table

# Find what was deployed on Apr 10
aws cloudtrail lookup-events \
  --start-time 2026-04-09T00:00:00 --end-time 2026-04-11T00:00:00 \
  --lookup-attributes AttributeKey=EventName,AttributeValue=UpdateFunctionConfiguration \
  --region ap-southeast-1
```

**Fix:** The $432/day baseline itself suggests ongoing verbose logging. Both the spike and the baseline need to be addressed — set log levels to INFO/WARN in all production services and apply retention policies (see item 7).

---

### 3. GuardDuty — Month-Start Spikes STILL HAPPENING ($9,887/month) ⚠️ Carried Over

GuardDuty spiked again on Apr 1 ($309, 3.0σ) and Apr 2 ($295, 2.8σ). **This exact pattern was identified and actioned in the February–April audit and was not fixed.** The same recurring billing-cycle surge is costing approximately $9,887/month in excess charges.

```bash
# Check which data sources are enabled and their usage
aws guardduty list-detectors --region ap-southeast-1

aws guardduty get-usage-statistics \
  --detector-id <detector-id> \
  --usage-statistic-type SUM_BY_DATA_SOURCE \
  --usage-criteria '{"DataSources":["S3_LOGS","FLOW_LOGS","CLOUD_TRAIL","DNS_LOGS"]}' \
  --region ap-southeast-1

# Check if S3 data events protection is enabled on all buckets
aws guardduty get-detector --detector-id <detector-id> \
  --query 'DataSources.S3Logs'
```

**Fix:** Scope GuardDuty S3 protection to only sensitive buckets (PII, financial data, secrets). Disable blanket S3 data event analysis.

```bash
# Disable S3 logs globally, then re-enable only for sensitive buckets
aws guardduty update-detector --detector-id <detector-id> \
  --data-sources '{"S3Logs":{"Enable":false}}' \
  --region ap-southeast-1
```

---

### 4. CloudTrail Data Events — STILL HAPPENING ($2,171/month) ⚠️ Carried Over

CloudTrail spiked on Apr 6 ($57.96, 2.6σ) and Apr 19 ($72.01, 3.9σ) vs. a $29 baseline. The Mar 26 data event issue identified in the previous audit was not removed. CloudTrail's April baseline ($29/day) is already nearly double what it was before March ($32 was a spike then; now it's the mean).

```bash
# List all trails and check data event configuration
aws cloudtrail describe-trails --region ap-southeast-1

aws cloudtrail get-event-selectors --trail-name <trail-name>

# Advanced selectors (newer trail configs)
aws cloudtrail get-advanced-event-selectors --trail-name <trail-name>
```

**Fix:** Remove data event selectors that were enabled around Mar 26. Keep only management events unless specific S3 buckets or Lambda functions require audit-level data event logging.

```bash
# Reset to management events only
aws cloudtrail put-event-selectors --trail-name <trail-name> \
  --event-selectors '[{
    "ReadWriteType": "All",
    "IncludeManagementEvents": true,
    "DataResources": []
  }]' --region ap-southeast-1
```

---

### 5. AWS Config + Security Hub — Apr 12 Correlated Spike ($852/month)

Config spiked to $31.30 on Apr 12 (4.0σ) and Security Hub spiked to $11.12 on the same day (3.4σ). This same-day correlation means a Config evaluation run triggered a large number of Security Hub findings, which Security Hub charges per finding.

**Root cause hypothesis:** A new Config rule was enabled around Apr 11–12, or a scheduled evaluation ran for the first time across all resources in the account.

```bash
# List Config rules and their evaluation modes
aws configservice describe-config-rules --region ap-southeast-1 \
  --query 'ConfigRules[].[ConfigRuleName,ConfigRuleState,Source.SourceIdentifier]'

# Check Config rule evaluation results for Apr 12
aws configservice get-compliance-details-by-config-rule \
  --config-rule-name <rule-name> --region ap-southeast-1

# Check Security Hub findings count on Apr 12
aws securityhub get-findings --region ap-southeast-1 \
  --filters '{"CreatedAt":[{"Start":"2026-04-12T00:00:00Z","End":"2026-04-13T00:00:00Z","DateRange":{"Value":1,"Unit":"DAYS"}}]}' \
  --query 'Findings | length(@)'
```

**Fix:** Review newly added Config rules. If a broad managed rule (e.g. `REQUIRED_TAGS`) was applied to all resources for the first time, it will generate a spike on first evaluation. Disable rules that are not providing actionable findings, or scope them to specific resource types.

---

### 6. CloudTrail + Secrets Manager — Apr 19 Correlated Spike ($1,475/month combined)

CloudTrail hit $72 on Apr 19 (its highest April spike, 3.9σ) and Secrets Manager spiked to $12.57 on the same day (4.2σ above its mean of $6.61). This co-occurrence on Apr 19 — and again on Apr 6 for both services — points to a scheduled automation that rotates or accesses a large number of secrets and generates corresponding CloudTrail data events.

```bash
# Check Secrets Manager API activity on Apr 19
aws cloudtrail lookup-events \
  --start-time 2026-04-19T00:00:00 --end-time 2026-04-20T00:00:00 \
  --lookup-attributes AttributeKey=EventSource,AttributeValue=secretsmanager.amazonaws.com \
  --region ap-southeast-1

# List secrets to identify high-rotation candidates
aws secretsmanager list-secrets --region ap-southeast-1 \
  --query 'SecretList[].[Name,RotationEnabled,LastRotatedDate,RotationRules.AutomaticallyAfterDays]'
```

**Fix:** Identify the scheduled rotation job. If secrets are being rotated unnecessarily frequently (e.g. every few days instead of 30 days), extend the rotation interval. Also check if CloudTrail data events are tracking GetSecretValue calls — these generate one event per API call and can be very high volume.

---

### 7. CloudWatch Logs — Structural $12,544/Month Cost ($5,017/month savings)

Unlike the spike findings, this is a baseline cost issue. CloudWatch Logs total spend in April was $12,544 — $432/day as the new "normal." Log groups without retention policies are accumulating data indefinitely at $0.03/GB/month storage plus $0.50/GB ingestion.

```bash
# Find all log groups without retention, by stored size (largest first)
aws logs describe-log-groups --region ap-southeast-1 \
  --query 'sort_by(logGroups[?!retentionInDays], &storedBytes)[-20:].[logGroupName,storedBytes]' \
  --output table

# Find all log groups regardless of retention, by stored size
aws logs describe-log-groups --region ap-southeast-1 \
  --query 'sort_by(logGroups, &storedBytes)[-10:].[logGroupName,storedBytes,retentionInDays]' \
  --output table
```

**Fix — apply tiered retention:**

```bash
# Apply 30-day retention to all log groups with no retention policy
aws logs describe-log-groups --region ap-southeast-1 \
  --query 'logGroups[?!retentionInDays].logGroupName' --output text | \
  tr '\t' '\n' | while read group; do
    echo "Setting retention on: $group"
    aws logs put-retention-policy \
      --region ap-southeast-1 \
      --log-group-name "$group" \
      --retention-in-days 30
  done
```

**Recommended retention tiers:**
| Log Group Type | Retention |
|---|---|
| Application/debug logs | 14–30 days |
| Access logs | 90 days |
| Security/audit logs | 365 days |
| Compliance-required logs | Export to S3 Glacier, then delete |

---

### 8. Stopped EC2 Instance — 30 Days (i-07970379c328e00f1, t3a.large)

A `t3a.large` instance has been stopped for 30 days (launched Apr 3, 2026). It just crossed the alerting threshold. This is still within the window where the team likely knows its purpose — confirm with the owner before terminating.

```bash
# Check instance details and tags (find the owner)
aws ec2 describe-instances --region ap-southeast-1 \
  --instance-ids i-07970379c328e00f1 \
  --query 'Reservations[].Instances[].[Tags,InstanceType,State.Name]'

# Check attached volumes
aws ec2 describe-volumes --region ap-southeast-1 \
  --filters Name=attachment.instance-id,Values=i-07970379c328e00f1 \
  --query 'Volumes[].[VolumeId,Size,VolumeType]'

# Terminate if confirmed unused
aws ec2 terminate-instances --region ap-southeast-1 \
  --instance-ids i-07970379c328e00f1
```

---

### 9. ElastiCache Extra Replica — STILL RUNNING ($1,980/month) ⚠️ Carried Over

The ElastiCache cluster is still running at $450.77/day — the same elevated level identified in the previous audit. The extra replica node added around Mar 26 has not been removed. April data confirms the $450/day spend is now the permanent new baseline rather than a temporary spike.

```bash
# Check current replica count vs what it was before Mar 26
aws elasticache describe-replication-groups --region ap-southeast-1 \
  --query 'ReplicationGroups[].[ReplicationGroupId,MemberClusters,Description]'

aws elasticache describe-cache-clusters --region ap-southeast-1 \
  --query 'CacheClusters[].[CacheClusterId,CacheNodeType,CacheClusterCreateTime,CacheClusterStatus]' \
  --output table
```

**Fix:** If the replica was added temporarily and is no longer needed:

```bash
aws elasticache decrease-replica-count --region ap-southeast-1 \
  --replication-group-id <group-id> \
  --new-replica-count <current-minus-1> \
  --apply-immediately
```

---

### 10. NAT Gateway — VPC Endpoints ($169/month)

NAT Gateway spend is $565/month, exceeding the $100 threshold. Adding VPC Gateway Endpoints for S3 and DynamoDB (both free) eliminates data processing charges for traffic to those services.

```bash
# Check what VPC endpoints already exist
aws ec2 describe-vpc-endpoints --region ap-southeast-1 \
  --query 'VpcEndpoints[].[ServiceName,State,VpcId]' --output table

# Get VPC and route table IDs
aws ec2 describe-vpcs --region ap-southeast-1 \
  --query 'Vpcs[].[VpcId,Tags[?Key==`Name`].Value|[0]]' --output table

# Create free S3 gateway endpoint
aws ec2 create-vpc-endpoint --region ap-southeast-1 \
  --vpc-id <vpc-id> \
  --service-name com.amazonaws.ap-southeast-1.s3 \
  --route-table-ids <route-table-id>

# Create free DynamoDB gateway endpoint
aws ec2 create-vpc-endpoint --region ap-southeast-1 \
  --vpc-id <vpc-id> \
  --service-name com.amazonaws.ap-southeast-1.dynamodb \
  --route-table-ids <route-table-id>
```

---

## Summary of Expected Savings

| Action | Est. Monthly Savings | Confidence | Effort | Status |
|--------|---------------------|------------|--------|--------|
| GuardDuty S3 data source scoping | $9,887 | Medium | 1 hr | ⚠️ Carried over |
| CloudWatch spike investigation (Apr 10) | $8,191 | Medium | 1 hr | New |
| CloudWatch log retention policies | $5,017 | High | 2 hrs | New |
| CloudTrail data events removal | $2,171 | High | 30 min | ⚠️ Carried over |
| ElastiCache replica removal | ~$1,980 | High | 30 min | ⚠️ Carried over |
| CloudTrail + Secrets Manager (Apr 19) | $1,475 | Medium | 30 min | New |
| Config + Security Hub (Apr 12) | $852 | Medium | 30 min | New |
| Stopped EC2 — 171 days (m6a.large) | EBS TBD | High | 15 min | New |
| Stopped EC2 — 30 days (t3a.large) | EBS TBD | Medium | 15 min | New |
| NAT Gateway VPC endpoints | $169 | Medium | 2 hrs | New |
| **Total** | **~$29,742+/month** | | **~9 hrs** | |

---

## Items From Previous Audit Still Open

The following items were identified in the Feb–Apr audit (`ACTION-PLAN.md` in `2026-05-04T014515Z/`) and have **not been resolved**:

| Item | Previous Finding | April Status |
|------|-----------------|--------------|
| GuardDuty S3 scoping | $17K/mo opportunity | Still spiking monthly |
| CloudTrail data events | Enabled Mar 26, $9K/mo | Still elevated baseline |
| ElastiCache replica | Added ~Mar 26, $1,980/mo | Still running |
| CloudWatch log retention | $10K/mo opportunity | $12,544 baseline in April |

These should be treated as **overdue** and resolved before the next audit run.
