# AWS Cost Audit Findings

**Audit Run:** aws-cost-audit-2026-04-01-2026-04-30-20260504T041657Z-default  
**Period:** 2026-04-01 to 2026-04-30  
**Generated:** 2026-05-04 04:16:57 UTC  

## P1 — High Confidence Waste (13 findings)

### FND-5530ca85579570b0: Daily Cost Spike: AmazonCloudWatch on 2026-04-10 ($705 vs avg $432)

| Field | Value |
|-------|-------|
| Priority | P1 |
| Category | Anomaly |
| Service | AmazonCloudWatch |
| Account |  () |
| Region |  |
| Resources |  |
| Est. Savings | $8191.11/month |
| Confidence | Medium |
| Risk | Low |
| Status | open |

**Summary:** Daily cost for AmazonCloudWatch spiked to $705.38 on 2026-04-10, which is 4.1 standard deviations above the mean ($432.35).

**Evidence:** Date: 2026-04-10. Cost: $705.38. Mean: $432.35. Stddev: $65.94. Threshold (2.0×σ): $564.23.

**Recommended Action:** Investigate the root cause of the spike. Check CloudTrail for unusual API calls. Review new resources created on or before 2026-04-10.

---

### FND-8a468dd3fd4a146c: Daily Cost Spike: Amazon GuardDuty on 2026-04-01 ($309 vs avg $138)

| Field | Value |
|-------|-------|
| Priority | P1 |
| Category | Anomaly |
| Service | Amazon GuardDuty |
| Account |  () |
| Region |  |
| Resources |  |
| Est. Savings | $5154.84/month |
| Confidence | Medium |
| Risk | Low |
| Status | open |

**Summary:** Daily cost for Amazon GuardDuty spiked to $309.45 on 2026-04-01, which is 3.0 standard deviations above the mean ($137.62).

**Evidence:** Date: 2026-04-01. Cost: $309.45. Mean: $137.62. Stddev: $56.34. Threshold (2.0×σ): $250.31.

**Recommended Action:** Investigate the root cause of the spike. Check CloudTrail for unusual API calls. Review new resources created on or before 2026-04-01.

---

### FND-3df8df824b76c5cd: Daily Cost Spike: Amazon GuardDuty on 2026-04-02 ($295 vs avg $138)

| Field | Value |
|-------|-------|
| Priority | P1 |
| Category | Anomaly |
| Service | Amazon GuardDuty |
| Account |  () |
| Region |  |
| Resources |  |
| Est. Savings | $4732.74/month |
| Confidence | Medium |
| Risk | Low |
| Status | open |

**Summary:** Daily cost for Amazon GuardDuty spiked to $295.38 on 2026-04-02, which is 2.8 standard deviations above the mean ($137.62).

**Evidence:** Date: 2026-04-02. Cost: $295.38. Mean: $137.62. Stddev: $56.34. Threshold (2.0×σ): $250.31.

**Recommended Action:** Investigate the root cause of the spike. Check CloudTrail for unusual API calls. Review new resources created on or before 2026-04-02.

---

### FND-d0686663b09d2a06: Daily Cost Spike: AWS CloudTrail on 2026-04-19 ($72 vs avg $29)

| Field | Value |
|-------|-------|
| Priority | P1 |
| Category | Anomaly |
| Service | AWS CloudTrail |
| Account |  () |
| Region |  |
| Resources |  |
| Est. Savings | $1296.45/month |
| Confidence | Medium |
| Risk | Low |
| Status | open |

**Summary:** Daily cost for AWS CloudTrail spiked to $72.01 on 2026-04-19, which is 3.9 standard deviations above the mean ($28.79).

**Evidence:** Date: 2026-04-19. Cost: $72.01. Mean: $28.79. Stddev: $11.13. Threshold (2.0×σ): $51.06.

**Recommended Action:** Investigate the root cause of the spike. Check CloudTrail for unusual API calls. Review new resources created on or before 2026-04-19.

---

### FND-7dc376792cf6b08e: Daily Cost Spike: AWS CloudTrail on 2026-04-06 ($58 vs avg $29)

| Field | Value |
|-------|-------|
| Priority | P1 |
| Category | Anomaly |
| Service | AWS CloudTrail |
| Account |  () |
| Region |  |
| Resources |  |
| Est. Savings | $875.07/month |
| Confidence | Medium |
| Risk | Low |
| Status | open |

**Summary:** Daily cost for AWS CloudTrail spiked to $57.96 on 2026-04-06, which is 2.6 standard deviations above the mean ($28.79).

**Evidence:** Date: 2026-04-06. Cost: $57.96. Mean: $28.79. Stddev: $11.13. Threshold (2.0×σ): $51.06.

**Recommended Action:** Investigate the root cause of the spike. Check CloudTrail for unusual API calls. Review new resources created on or before 2026-04-06.

---

### FND-302555769e1268cf: Daily Cost Spike: AWS Config on 2026-04-12 ($31 vs avg $9)

| Field | Value |
|-------|-------|
| Priority | P1 |
| Category | Anomaly |
| Service | AWS Config |
| Account |  () |
| Region |  |
| Resources |  |
| Est. Savings | $669.26/month |
| Confidence | Medium |
| Risk | Low |
| Status | open |

**Summary:** Daily cost for AWS Config spiked to $31.30 on 2026-04-12, which is 4.0 standard deviations above the mean ($8.99).

**Evidence:** Date: 2026-04-12. Cost: $31.30. Mean: $8.99. Stddev: $5.59. Threshold (2.0×σ): $20.18.

**Recommended Action:** Investigate the root cause of the spike. Check CloudTrail for unusual API calls. Review new resources created on or before 2026-04-12.

---

### FND-5332304715a83fdd: Daily Cost Spike: AWS Security Hub on 2026-04-12 ($11 vs avg $5)

| Field | Value |
|-------|-------|
| Priority | P1 |
| Category | Anomaly |
| Service | AWS Security Hub |
| Account |  () |
| Region |  |
| Resources |  |
| Est. Savings | $182.56/month |
| Confidence | Medium |
| Risk | Low |
| Status | open |

**Summary:** Daily cost for AWS Security Hub spiked to $11.12 on 2026-04-12, which is 3.4 standard deviations above the mean ($5.03).

**Evidence:** Date: 2026-04-12. Cost: $11.12. Mean: $5.03. Stddev: $1.77. Threshold (2.0×σ): $8.57.

**Recommended Action:** Investigate the root cause of the spike. Check CloudTrail for unusual API calls. Review new resources created on or before 2026-04-12.

---

### FND-11940b561f12033e: Daily Cost Spike: AWS Secrets Manager on 2026-04-19 ($13 vs avg $7)

| Field | Value |
|-------|-------|
| Priority | P1 |
| Category | Anomaly |
| Service | AWS Secrets Manager |
| Account |  () |
| Region |  |
| Resources |  |
| Est. Savings | $178.93/month |
| Confidence | Medium |
| Risk | Low |
| Status | open |

**Summary:** Daily cost for AWS Secrets Manager spiked to $12.57 on 2026-04-19, which is 4.2 standard deviations above the mean ($6.61).

**Evidence:** Date: 2026-04-19. Cost: $12.57. Mean: $6.61. Stddev: $1.42. Threshold (2.0×σ): $9.44.

**Recommended Action:** Investigate the root cause of the spike. Check CloudTrail for unusual API calls. Review new resources created on or before 2026-04-19.

---

### FND-19357867cc00c26b: Daily Cost Spike: AWS Security Hub on 2026-04-05 ($10 vs avg $5)

| Field | Value |
|-------|-------|
| Priority | P1 |
| Category | Anomaly |
| Service | AWS Security Hub |
| Account |  () |
| Region |  |
| Resources |  |
| Est. Savings | $142.86/month |
| Confidence | Medium |
| Risk | Low |
| Status | open |

**Summary:** Daily cost for AWS Security Hub spiked to $9.80 on 2026-04-05, which is 2.7 standard deviations above the mean ($5.03).

**Evidence:** Date: 2026-04-05. Cost: $9.80. Mean: $5.03. Stddev: $1.77. Threshold (2.0×σ): $8.57.

**Recommended Action:** Investigate the root cause of the spike. Check CloudTrail for unusual API calls. Review new resources created on or before 2026-04-05.

---

### FND-3748378674faf280: Daily Cost Spike: AWS Secrets Manager on 2026-04-06 ($10 vs avg $7)

| Field | Value |
|-------|-------|
| Priority | P1 |
| Category | Anomaly |
| Service | AWS Secrets Manager |
| Account |  () |
| Region |  |
| Resources |  |
| Est. Savings | $95.57/month |
| Confidence | Medium |
| Risk | Low |
| Status | open |

**Summary:** Daily cost for AWS Secrets Manager spiked to $9.79 on 2026-04-06, which is 2.2 standard deviations above the mean ($6.61).

**Evidence:** Date: 2026-04-06. Cost: $9.79. Mean: $6.61. Stddev: $1.42. Threshold (2.0×σ): $9.44.

**Recommended Action:** Investigate the root cause of the spike. Check CloudTrail for unusual API calls. Review new resources created on or before 2026-04-06.

---

### FND-ffa65ff91fe0d888: Daily Cost Spike: Amazon Simple Email Service on 2026-04-14 ($7 vs avg $5)

| Field | Value |
|-------|-------|
| Priority | P1 |
| Category | Anomaly |
| Service | Amazon Simple Email Service |
| Account |  () |
| Region |  |
| Resources |  |
| Est. Savings | $58.00/month |
| Confidence | Medium |
| Risk | Low |
| Status | open |

**Summary:** Daily cost for Amazon Simple Email Service spiked to $7.04 on 2026-04-14, which is 2.5 standard deviations above the mean ($5.10).

**Evidence:** Date: 2026-04-14. Cost: $7.04. Mean: $5.10. Stddev: $0.78. Threshold (2.0×σ): $6.66.

**Recommended Action:** Investigate the root cause of the spike. Check CloudTrail for unusual API calls. Review new resources created on or before 2026-04-14.

---

### FND-148cd760cb02013f: Daily Cost Spike: Amazon ElastiCache on 2026-04-22 ($451 vs avg $451)

| Field | Value |
|-------|-------|
| Priority | P1 |
| Category | Anomaly |
| Service | Amazon ElastiCache |
| Account |  () |
| Region |  |
| Resources |  |
| Est. Savings | $1.38/month |
| Confidence | Medium |
| Risk | Low |
| Status | open |

**Summary:** Daily cost for Amazon ElastiCache spiked to $450.77 on 2026-04-22, which is 3.6 standard deviations above the mean ($450.72).

**Evidence:** Date: 2026-04-22. Cost: $450.77. Mean: $450.72. Stddev: $0.01. Threshold (2.0×σ): $450.75.

**Recommended Action:** Investigate the root cause of the spike. Check CloudTrail for unusual API calls. Review new resources created on or before 2026-04-22.

---

### FND-be31d4c14fe6659e: Daily Cost Spike: Amazon ElastiCache on 2026-04-20 ($451 vs avg $451)

| Field | Value |
|-------|-------|
| Priority | P1 |
| Category | Anomaly |
| Service | Amazon ElastiCache |
| Account |  () |
| Region |  |
| Resources |  |
| Est. Savings | $1.22/month |
| Confidence | Medium |
| Risk | Low |
| Status | open |

**Summary:** Daily cost for Amazon ElastiCache spiked to $450.76 on 2026-04-20, which is 3.2 standard deviations above the mean ($450.72).

**Evidence:** Date: 2026-04-20. Cost: $450.76. Mean: $450.72. Stddev: $0.01. Threshold (2.0×σ): $450.75.

**Recommended Action:** Investigate the root cause of the spike. Check CloudTrail for unusual API calls. Review new resources created on or before 2026-04-20.

---

## P2 — Optimization Required (4 findings)

### FND-8159b6c4136da2fb: High CloudWatch Logs Cost ($12544/month) in 

| Field | Value |
|-------|-------|
| Priority | P2 |
| Category | Observability |
| Service | Amazon CloudWatch |
| Account |  () |
| Region |  |
| Resources |  |
| Est. Savings | $5017.68/month |
| Confidence | Medium |
| Risk | Low |
| Status | open |

**Summary:** CloudWatch costs in  total $12544.21/month, exceeding the $50 threshold.

**Evidence:** Aggregated CloudWatch service cost from cost_monthly: $12544.21. Region: .

**Recommended Action:** Review log groups without retention policies. Set retention (e.g. 30-90 days) to reduce storage costs. Reduce verbose logging where appropriate.

---

### FND-9d7c66dead7b1667: High NAT Gateway Cost ($566/month) in 

| Field | Value |
|-------|-------|
| Priority | P2 |
| Category | Architecture |
| Service | Amazon Virtual Private Cloud |
| Account |  () |
| Region |  |
| Resources |  |
| Est. Savings | $169.70/month |
| Confidence | Medium |
| Risk | Medium |
| Status | open |

**Summary:** NAT Gateway charges in  total $565.65/month, exceeding the $100 threshold.

**Evidence:** Aggregated VPC service cost from cost_monthly: $565.65. Region: . Account: .

**Recommended Action:** Review traffic routing. Consider VPC endpoints for S3/DynamoDB to eliminate NAT data processing charges. Review cross-AZ NAT Gateway usage.

---

### FND-6f2f534130e98ab8: Stopped EC2 Instance (i-07970379c328e00f1, t3a.large)

| Field | Value |
|-------|-------|
| Priority | P2 |
| Category | Waste |
| Service | Amazon EC2 |
| Account |  () |
| Region | ap-southeast-1 |
| Resources | i-07970379c328e00f1 |
| Est. Savings | $0.00/month |
| Confidence | Medium |
| Risk | Medium |
| Status | open |

**Summary:** Instance i-07970379c328e00f1 has been stopped. Attached EBS volumes continue to incur charges.

**Evidence:** Instance state: stopped. Type: t3a.large. Launch time: 2026-04-03T07:45:08+00:00. Age: 30 days. Threshold: 30 days.

**Recommended Action:** Confirm the instance is no longer needed. Terminate with: aws ec2 terminate-instances --instance-ids i-07970379c328e00f1

**Source Files:** aws-cost-audit/2026-05-04T041657Z/raw/inventory/ec2_describe-instances_ap-southeast-1.json

---

### FND-495ccbb14de4be86: Stopped EC2 Instance (i-0d16f29483c77b603, m6a.large)

| Field | Value |
|-------|-------|
| Priority | P2 |
| Category | Waste |
| Service | Amazon EC2 |
| Account |  () |
| Region | ap-southeast-1 |
| Resources | i-0d16f29483c77b603 |
| Est. Savings | $0.00/month |
| Confidence | Medium |
| Risk | Medium |
| Status | open |

**Summary:** Instance i-0d16f29483c77b603 has been stopped. Attached EBS volumes continue to incur charges.

**Evidence:** Instance state: stopped. Type: m6a.large. Launch time: 2025-11-13T10:51:43+00:00. Age: 171 days. Threshold: 30 days.

**Recommended Action:** Confirm the instance is no longer needed. Terminate with: aws ec2 terminate-instances --instance-ids i-0d16f29483c77b603

**Source Files:** aws-cost-audit/2026-05-04T041657Z/raw/inventory/ec2_describe-instances_ap-southeast-1.json

---

