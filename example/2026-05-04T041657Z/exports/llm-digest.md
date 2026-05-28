# AWS Cost Audit — LLM Digest

> This is the first file to read. It summarizes all findings from this audit run.

**Audit Run:** `aws-cost-audit-2026-04-01-2026-04-30-20260504T041657Z-default`  
**Period:** 2026-04-01 to 2026-04-30  
**Generated:** 2026-05-04 04:16:57 UTC  
**AWS Profile:** default  

## Executive Summary

| Metric | Value |
|--------|-------|
| Total Monthly Spend | $721279.16 |
| Total Estimated Savings | $26767.37/month |
| Total Findings | 17 |
| P0 (Urgent) | 0 |
| P1 (High Confidence Waste) | 13 |
| P2 (Optimization) | 4 |
| P3 (Governance) | 0 |

## Top Findings by Savings

| Priority | Title | Service | Account | Savings/mo | Confidence |
|----------|-------|---------|---------|------------|------------|
| P1 | Daily Cost Spike: AmazonCloudWatch on 2026-04-10 ($705 vs... | AmazonCloudWatch |  | $8191 | Medium |
| P1 | Daily Cost Spike: Amazon GuardDuty on 2026-04-01 ($309 vs... | Amazon GuardDuty |  | $5155 | Medium |
| P1 | Daily Cost Spike: Amazon GuardDuty on 2026-04-02 ($295 vs... | Amazon GuardDuty |  | $4733 | Medium |
| P1 | Daily Cost Spike: AWS CloudTrail on 2026-04-19 ($72 vs av... | AWS CloudTrail |  | $1296 | Medium |
| P1 | Daily Cost Spike: AWS CloudTrail on 2026-04-06 ($58 vs av... | AWS CloudTrail |  | $875 | Medium |
| P1 | Daily Cost Spike: AWS Config on 2026-04-12 ($31 vs avg $9) | AWS Config |  | $669 | Medium |
| P1 | Daily Cost Spike: AWS Security Hub on 2026-04-12 ($11 vs ... | AWS Security Hub |  | $183 | Medium |
| P1 | Daily Cost Spike: AWS Secrets Manager on 2026-04-19 ($13 ... | AWS Secrets Manager |  | $179 | Medium |
| P1 | Daily Cost Spike: AWS Security Hub on 2026-04-05 ($10 vs ... | AWS Security Hub |  | $143 | Medium |
| P1 | Daily Cost Spike: AWS Secrets Manager on 2026-04-06 ($10 ... | AWS Secrets Manager |  | $96 | Medium |

## P1 — High Confidence Waste

### FND-5530ca85579570b0: Daily Cost Spike: AmazonCloudWatch on 2026-04-10 ($705 vs avg $432)

Daily cost for AmazonCloudWatch spiked to $705.38 on 2026-04-10, which is 4.1 standard deviations above the mean ($432.35).

- **Evidence:** Date: 2026-04-10. Cost: $705.38. Mean: $432.35. Stddev: $65.94. Threshold (2.0×σ): $564.23.
- **Savings:** $8191.11/month
- **Action:** Investigate the root cause of the spike. Check CloudTrail for unusual API calls. Review new resources created on or before 2026-04-10.

### FND-8a468dd3fd4a146c: Daily Cost Spike: Amazon GuardDuty on 2026-04-01 ($309 vs avg $138)

Daily cost for Amazon GuardDuty spiked to $309.45 on 2026-04-01, which is 3.0 standard deviations above the mean ($137.62).

- **Evidence:** Date: 2026-04-01. Cost: $309.45. Mean: $137.62. Stddev: $56.34. Threshold (2.0×σ): $250.31.
- **Savings:** $5154.84/month
- **Action:** Investigate the root cause of the spike. Check CloudTrail for unusual API calls. Review new resources created on or before 2026-04-01.

### FND-3df8df824b76c5cd: Daily Cost Spike: Amazon GuardDuty on 2026-04-02 ($295 vs avg $138)

Daily cost for Amazon GuardDuty spiked to $295.38 on 2026-04-02, which is 2.8 standard deviations above the mean ($137.62).

- **Evidence:** Date: 2026-04-02. Cost: $295.38. Mean: $137.62. Stddev: $56.34. Threshold (2.0×σ): $250.31.
- **Savings:** $4732.74/month
- **Action:** Investigate the root cause of the spike. Check CloudTrail for unusual API calls. Review new resources created on or before 2026-04-02.

### FND-d0686663b09d2a06: Daily Cost Spike: AWS CloudTrail on 2026-04-19 ($72 vs avg $29)

Daily cost for AWS CloudTrail spiked to $72.01 on 2026-04-19, which is 3.9 standard deviations above the mean ($28.79).

- **Evidence:** Date: 2026-04-19. Cost: $72.01. Mean: $28.79. Stddev: $11.13. Threshold (2.0×σ): $51.06.
- **Savings:** $1296.45/month
- **Action:** Investigate the root cause of the spike. Check CloudTrail for unusual API calls. Review new resources created on or before 2026-04-19.

### FND-7dc376792cf6b08e: Daily Cost Spike: AWS CloudTrail on 2026-04-06 ($58 vs avg $29)

Daily cost for AWS CloudTrail spiked to $57.96 on 2026-04-06, which is 2.6 standard deviations above the mean ($28.79).

- **Evidence:** Date: 2026-04-06. Cost: $57.96. Mean: $28.79. Stddev: $11.13. Threshold (2.0×σ): $51.06.
- **Savings:** $875.07/month
- **Action:** Investigate the root cause of the spike. Check CloudTrail for unusual API calls. Review new resources created on or before 2026-04-06.

### FND-302555769e1268cf: Daily Cost Spike: AWS Config on 2026-04-12 ($31 vs avg $9)

Daily cost for AWS Config spiked to $31.30 on 2026-04-12, which is 4.0 standard deviations above the mean ($8.99).

- **Evidence:** Date: 2026-04-12. Cost: $31.30. Mean: $8.99. Stddev: $5.59. Threshold (2.0×σ): $20.18.
- **Savings:** $669.26/month
- **Action:** Investigate the root cause of the spike. Check CloudTrail for unusual API calls. Review new resources created on or before 2026-04-12.

### FND-5332304715a83fdd: Daily Cost Spike: AWS Security Hub on 2026-04-12 ($11 vs avg $5)

Daily cost for AWS Security Hub spiked to $11.12 on 2026-04-12, which is 3.4 standard deviations above the mean ($5.03).

- **Evidence:** Date: 2026-04-12. Cost: $11.12. Mean: $5.03. Stddev: $1.77. Threshold (2.0×σ): $8.57.
- **Savings:** $182.56/month
- **Action:** Investigate the root cause of the spike. Check CloudTrail for unusual API calls. Review new resources created on or before 2026-04-12.

### FND-11940b561f12033e: Daily Cost Spike: AWS Secrets Manager on 2026-04-19 ($13 vs avg $7)

Daily cost for AWS Secrets Manager spiked to $12.57 on 2026-04-19, which is 4.2 standard deviations above the mean ($6.61).

- **Evidence:** Date: 2026-04-19. Cost: $12.57. Mean: $6.61. Stddev: $1.42. Threshold (2.0×σ): $9.44.
- **Savings:** $178.93/month
- **Action:** Investigate the root cause of the spike. Check CloudTrail for unusual API calls. Review new resources created on or before 2026-04-19.

### FND-19357867cc00c26b: Daily Cost Spike: AWS Security Hub on 2026-04-05 ($10 vs avg $5)

Daily cost for AWS Security Hub spiked to $9.80 on 2026-04-05, which is 2.7 standard deviations above the mean ($5.03).

- **Evidence:** Date: 2026-04-05. Cost: $9.80. Mean: $5.03. Stddev: $1.77. Threshold (2.0×σ): $8.57.
- **Savings:** $142.86/month
- **Action:** Investigate the root cause of the spike. Check CloudTrail for unusual API calls. Review new resources created on or before 2026-04-05.

### FND-3748378674faf280: Daily Cost Spike: AWS Secrets Manager on 2026-04-06 ($10 vs avg $7)

Daily cost for AWS Secrets Manager spiked to $9.79 on 2026-04-06, which is 2.2 standard deviations above the mean ($6.61).

- **Evidence:** Date: 2026-04-06. Cost: $9.79. Mean: $6.61. Stddev: $1.42. Threshold (2.0×σ): $9.44.
- **Savings:** $95.57/month
- **Action:** Investigate the root cause of the spike. Check CloudTrail for unusual API calls. Review new resources created on or before 2026-04-06.

### FND-ffa65ff91fe0d888: Daily Cost Spike: Amazon Simple Email Service on 2026-04-14 ($7 vs avg $5)

Daily cost for Amazon Simple Email Service spiked to $7.04 on 2026-04-14, which is 2.5 standard deviations above the mean ($5.10).

- **Evidence:** Date: 2026-04-14. Cost: $7.04. Mean: $5.10. Stddev: $0.78. Threshold (2.0×σ): $6.66.
- **Savings:** $58.00/month
- **Action:** Investigate the root cause of the spike. Check CloudTrail for unusual API calls. Review new resources created on or before 2026-04-14.

### FND-148cd760cb02013f: Daily Cost Spike: Amazon ElastiCache on 2026-04-22 ($451 vs avg $451)

Daily cost for Amazon ElastiCache spiked to $450.77 on 2026-04-22, which is 3.6 standard deviations above the mean ($450.72).

- **Evidence:** Date: 2026-04-22. Cost: $450.77. Mean: $450.72. Stddev: $0.01. Threshold (2.0×σ): $450.75.
- **Savings:** $1.38/month
- **Action:** Investigate the root cause of the spike. Check CloudTrail for unusual API calls. Review new resources created on or before 2026-04-22.

### FND-be31d4c14fe6659e: Daily Cost Spike: Amazon ElastiCache on 2026-04-20 ($451 vs avg $451)

Daily cost for Amazon ElastiCache spiked to $450.76 on 2026-04-20, which is 3.2 standard deviations above the mean ($450.72).

- **Evidence:** Date: 2026-04-20. Cost: $450.76. Mean: $450.72. Stddev: $0.01. Threshold (2.0×σ): $450.75.
- **Savings:** $1.22/month
- **Action:** Investigate the root cause of the spike. Check CloudTrail for unusual API calls. Review new resources created on or before 2026-04-20.

## Savings by Service

| Service | Est. Savings/mo |
|---------|----------------|
| AWS Config | $669.26 |
| AWS Security Hub | $325.42 |
| Amazon Simple Email Service | $58.00 |
| Amazon ElastiCache | $2.60 |
| Amazon Virtual Private Cloud | $169.70 |
| AmazonCloudWatch | $8191.11 |
| Amazon GuardDuty | $9887.58 |
| AWS CloudTrail | $2171.52 |
| AWS Secrets Manager | $274.50 |
| Amazon CloudWatch | $5017.68 |

## Savings by Account

| Account | Est. Savings/mo |
|---------|----------------|
|  | $26767.37 |

## Known Limitations

- Savings estimates use static pricing tables and may not reflect your exact pricing tier.
- CloudWatch and NAT Gateway cost findings are based on aggregated service spend, not per-resource attribution.
- Stopped EC2 compute savings are not estimated (already stopped); attached EBS costs are captured separately.
- Rightsizing recommendations require Compute Optimizer data (not yet collected in MVP).
