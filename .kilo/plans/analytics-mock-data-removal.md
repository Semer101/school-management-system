# Analytics Module - Mock Data Removal Plan

## Executive Summary

The analytics module displays real database data for admin dashboards but contains **hardcoded fallback values** in the frontend dashboard for teachers and parents. The backend analytics endpoints query actual database records correctly.

## Root Causes Found

### 1. Frontend Hardcoded Values (DashboardPage.tsx)

**File:** `sms-frontend/src/pages/dashboard/DashboardPage.tsx`

| Line | Issue | Impact |
|------|-------|--------|
| 195 | `attendanceAvg: 95.0` hardcoded for Parent KPIs | Parents see fake 95% attendance |
| 196 | `gradeAvg: 'A'` hardcoded for Parent KPIs | Parents see fake grade average |
| 241 | `'95.0%'` fallback when `teacherKpis?.attendance_rate` is null/undefined | Teachers see fake 95% if no attendance records exist |

### 2. Backend Fallback Value (academic_ctrl.go)

**File:** `sms-backend/controllers/academic_ctrl.go:2475`

```go
var attendanceRate float64 = 95.0 // default/fallback
```

This returns 95.0% for teachers when no attendance records exist. Should return 0.

### 3. Missing Parent Dashboard KPIs Endpoint

**Issue:** No backend endpoint exists to calculate parent-specific KPIs (children count, average attendance, average grade).

**Current Frontend Approach (lines 182-200):**
- Fetches `getMyChildren()` and `getParentTransactions()`
- Hardcodes `attendanceAvg: 95.0` and `gradeAvg: 'A'`
- Does NOT fetch actual attendance/grades for children

## Files to Modify

### Backend Changes

1. **`sms-backend/controllers/academic_ctrl.go`**
   - Remove hardcoded `95.0` fallback for attendance rate (line 2475)
   - Return `0.0` when no data exists

2. **`sms-backend/controllers/analytics_ctrl.go`**
   - Add `GetParentDashboardKPIs` endpoint for parents
   - Calculate average attendance and grade across all children

### Frontend Changes

1. **`sms-frontend/src/pages/dashboard/DashboardPage.tsx`**
   - Remove hardcoded `attendanceAvg: 95.0` and `gradeAvg: 'A'`
   - Fetch actual data for children's attendance and grades
   - Handle empty states properly

2. **`sms-frontend/src/api/parent.ts`**
   - Add `getParentDashboardKPIs()` endpoint call

## Validation Results

### Charts & Graphs (AnalyticsPage.tsx)
| Chart | Source | Status |
|-------|--------|--------|
| Students by Grade | `GetAnalyticsSummary` DB query | ✅ Real |
| Stream Distribution | `GetAnalyticsSummary` DB query | ✅ Real |
| Average Grades by Subject | `GetAnalyticsSummary` DB query | ✅ Real |
| Attendance Breakdown | `GetAnalyticsSummary` DB query | ✅ Real |
| Monthly Attendance Trend | `GetAnalyticsSummary` DB query | ✅ Real |
| Promotion Status | `GetAnalyticsSummary` DB query | ✅ Real |

### KPIs
| KPI | Source | Status |
|-----|--------|--------|
| Admin: Students | `GetDashboardKPIs` DB count | ✅ Real |
| Admin: Teachers | `GetDashboardKPIs` DB count | ✅ Real |
| Admin: Classes | `GetDashboardKPIs` DB count | ✅ Real |
| Admin: Subjects | `GetDashboardKPIs` DB count | ✅ Real |
| Admin: Present Today | `GetDashboardKPIs` DB query | ✅ Real |
| Admin: Absent Today | `GetDashboardKPIs` DB query | ✅ Real |
| Admin: Pending Transactions | `GetDashboardKPIs` DB query | ✅ Real |
| Teacher: Attendance Rate | `GetTeacherKPIs` with 95% fallback | ⚠️ Fake fallback |
| Parent: Attendance Avg | Hardcoded `95.0` | ❌ Fake |
| Parent: Grade Avg | Hardcoded `'A'` | ❌ Fake |
| Finance | `GetAnalyticsSummary` DB query | ✅ Real |

## Implementation Plan

### Phase 1: Backend Fixes
1. Fix `GetTeacherKPIs` to return `0.0` instead of `95.0` when no attendance data
2. Add `GetParentDashboardKPIs` endpoint

### Phase 2: Frontend Fixes
1. Update `DashboardPage.tsx` to fetch real parent KPIs
2. Remove hardcoded fallback values
3. Add proper empty state handling

### Phase 3: Testing
1. Verify with empty database (should show 0/null values)
2. Verify with seeded data (should show real calculated values)