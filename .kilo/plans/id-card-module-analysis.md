# ID Card Module Analysis & Fix Plan

## Issues Identified

### 1. Infinite Loading Issue
**Location:** `sms-frontend/src/pages/IdCardPage.tsx:107-116`

The `useEffect` only handles `Parent` role:
```tsx
useEffect(() => {
  if (role !== 'Parent') return
  getMyChildren()
    .then((res) => { ... })
    .finally(() => setLoading(false))
}, [role])
```

**Problem:** For `Admin`, `Teacher`, and `Student` roles, the effect returns early without calling `setLoading(false)`, causing infinite loading.

**Fix:** Move `setLoading(false)` outside the conditional or add an else branch.

### 2. Console Errors
**Location:** `sms-frontend/src/pages/IdCardPage.tsx:120`

```tsx
if (!authUser || !role) return null
```

**Problem:** Returns `null` instead of showing a proper loading state or error. This can cause React warnings and doesn't handle the case where auth is still initializing.

**Fix:** Add proper loading check and error handling.

### 3. RBAC - Route Access
**Location:** `sms-frontend/src/App.tsx:66`

Current: `allowedRoles={['Admin', 'Teacher', 'Student', 'Parent']}`

**Requirement:** ID Card should only be accessible by Admin.

**Fix:** Change to `allowedRoles={['Admin']}`

### 4. RBAC - Sidebar Visibility
**Location:** `sms-frontend/src/components/layout/Sidebar.tsx:38`

Current: `roles: ['Admin', 'Teacher', 'Student', 'Parent']`

**Fix:** Change to `roles: ['Admin']`

### 5. Profile Image Issues
**Location:** `sms-frontend/src/pages/IdCardPage.tsx:24`

```tsx
const profileUrl = user.avatar_url ? `${API_BASE}${user.avatar_url}` : null
```

**Problem:** 
- `avatar_url` from backend may be empty string or null
- No upload option for missing images
- Image not mandatory

**Fix:** 
- Handle empty string case
- Add upload option when image is missing
- Make profile image mandatory (show placeholder or upload prompt)

### 6. Layout Issues
**Location:** `sms-frontend/src/pages/IdCardPage.tsx:47`

```tsx
<div className="flex justify-center -mt-10">
```

**Problem:** `-mt-10` may cause overlap with header.

**Fix:** Adjust margin value.

## Files to Modify

1. `sms-frontend/src/pages/IdCardPage.tsx`
   - Fix loading state management
   - Fix RBAC to Admin only
   - Fix profile image handling
   - Improve layout

2. `sms-frontend/src/App.tsx`
   - Change route `allowedRoles` for `/id-card` to `['Admin']`

3. `sms-frontend/src/components/layout/Sidebar.tsx`
   - Change sidebar nav item `roles` for `/id-card` to `['Admin']`

## Verification Checklist

- [ ] Loading spinner doesn't show infinitely
- [ ] No console errors on page load
- [ ] Only Admin can access `/id-card` route
- [ ] Sidebar hides ID Card link for non-Admin users
- [ ] Profile image displays correctly from `avatar_url`
- [ ] Profile image shows placeholder when missing
- [ ] Upload option available for missing images
- [ ] Card alignment is proper (no header overlap)
- [ ] Profile image circle positioned correctly