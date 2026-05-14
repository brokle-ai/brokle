---
date: 2026-04-18
status: enacted
tags: [scaffolded-code, deletion, services]
---

# Scaffolded `BillingService` deletion

Re-applied the scaffolded-code lesson at the service level — the entire `BillingService` struct (RecordUsage, CalculateBill, ProcessPayment, CheckUsageQuotas, CreateBillingRecord, GetBillingHistory) was never wired. Zero callers for a year. Deleted 9 files, dropped 2 Postgres tables (`usage_records`, `usage_quotas`), kept the live `BillableUsage*` ClickHouse pipeline untouched. Signal for the next audit: `grep -l "NewXService"` callers count = 0 means delete, not defer. An orphan service actively misleads reviewers who assume it's load-bearing.
