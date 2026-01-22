# Migration Plan: Django -> Go Backend + New Frontend

## Goal
Replicate the existing Django app behavior and UX exactly, while migrating to a Go backend and a new frontend (technology TBD). A migration tool will move data into a new schema with a new encryption scheme for secret fields.

## Current App Parity Targets

### Routes and behaviors
Source: `server/urls.py`, `server/views.py`, `fvserver/urls.py`

- `/` home
- `/ajax/` DataTables JSON
- `/new/computer/` add computer
- `/new/secret/<id>/` add secret for computer
- `/info/secret/<id>/` secret details and request status
- `/info/<id or serial>/` computer details
- `/request/<id>/` request secret retrieval
- `/retrieve/<id>/` retrieve secret if approved
- `/approve/<id>/` approve or deny request
- `/manage-requests/` list outstanding requests
- `/verify/<serial>/<secret_type>/` escrow verification
- `/checkin/` client escrow endpoint
- `/login/`, `/logout/`, password change routes via Django auth

### Data model
Source: `server/models.py`

- `Computer`
  - `serial` (unique)
  - `username`
  - `computername`
  - `last_checkin`
- `Secret`
  - `computer` (FK)
  - `secret` (encrypted)
  - `secret_type` (`recovery_key`, `password`, `unlock_pin`)
  - `date_escrowed`
  - `rotation_required`
- `Request`
  - `secret` (FK)
  - `requesting_user` (FK -> User)
  - `approved` (null/true/false)
  - `auth_user` (approver)
  - `reason_for_request`
  - `reason_for_approval`
  - `date_requested`
  - `date_approved`
  - `current` (bool)
- Uses Django `User`, groups, and `can_approve` permission

### Workflow behavior
Source: `server/views.py`, `fvserver/system_settings.py`

- Request approval and denial flow; pending/approved/denied states.
- Self-approval optional gating (`APPROVE_OWN`).
- Global approver permission option (`ALL_APPROVE`).
- Cleanup: requests older than 7 days after approval are set `current=false`.
- Secret rotation signaling on retrieval (`ROTATE_VIEWED_SECRETS`).
- Emails for requests and approvals (if `SEND_EMAIL`).
- `HOST_NAME` used for link generation.

### UI/UX parity
Source: `server/templates/server/*.html`, `templates/*.html`, `site_static/*`

- Server-rendered pages using Bootstrap + DataTables.
- DataTables search/sort/pagination for the home list.
- Request, approve, retrieve, and manage screens.
- Login + password change flows (for local users).
- CSRF protections on all user input
- UI that allows admin users to create, edit, delete users and reset passwords.
- Utility in the main app binary to create the first admin user
- UI should look exactly the same as the existing django app
- For SAML users, isStaff or can approve permissions should be able to be set via saml attributes

## Migration Plan

### 1) Parity Spec and Contract Definition
- Enumerate and document every endpoint, request payload, response shape, and status code.
- Capture UI flows + required forms from templates.
- Freeze feature flags and configuration semantics:
  - `APPROVE_OWN`, `ALL_APPROVE`, `ROTATE_VIEWED_SECRETS`
  - Email behavior (`SEND_EMAIL`, `EMAIL_*`)
  - `HOST_NAME`

### 2) Go Backend Architecture (Design Only)
- **Auth**: prefer SAML-first with local reset (avoid importing Django hashes).
  - Data flags: `local_login_enabled` (tenant/user), `must_reset_password`, `password_hash` nullable, optional `auth_source` (`saml`, `local`, `hybrid`).
  - Migration: import identity fields only; set `password_hash=null`, `must_reset_password=true`; default `local_login_enabled=false` for SAML-only tenants.
  - UX: show SAML by default; local login only if enabled; if local login and `must_reset_password=true`, force reset flow before password auth.
  - Reset flow: admin-only resets set password + clear `must_reset_password`; no email sent; enforce strong password policy.
  - Admin actions: log admin-driven password resets and forced reset toggles (who/when/target user/IP).
  - Provide an admin UI to view these audit logs.
  - Add `must_reset_password` support for next-login resets (used during local-user migration).
- **Permissions**: replicate `can_approve` semantics and group assignment logic.
- **Endpoints**: provide exact behavior parity (including `checkin`/`verify` JSON).
- **Cleanup job**: scheduled cleanup for expired requests.
- **Email notifications**: remove entirely (no request/approval emails).
- **Audit/logging**: maintain approval/request metadata and event logs.
  - Admin audit events: password resets and forced reset toggles with fields `actor`, `timestamp`, `target_user`, `ip_address`, `action`, and `reason` (if provided).
  - Admin UI: add an audit log view for these events (read-only).

### 3) Data Model Mapping + Migration Tool
- **Extract**: dump from Django DB; decrypt `Secret.secret` with existing `FIELD_ENCRYPTION_KEY`.
- **Transform**: map old schema to new schema, preserving IDs and foreign keys where possible.
- **Re-encrypt**: apply new encryption scheme for secret fields.
- **Users and permissions**: migrate users, group membership, and `can_approve`.
  - Optional: migration tool accepts a password mapping file to set initial local passwords (when provided).
    - Proposed format: CSV with `username_or_email,password,must_reset_password` (last column optional).
    - Behavior: if an entry exists, set the password; `must_reset_password` defaults to `false` for mapped entries unless explicitly set.
    - Users not in the file default to `must_reset_password=true`.
- **Validation**:
  - Record counts by table.
  - Referential integrity checks.
  - Spot-check decrypt -> re-encrypt correctness.
  - Validate `/verify/` and `/checkin/` behavior on migrated data.


  Previous requqests and approvals should also be migrated.

### 4) Frontend Plan (Tech TBD)
- Option A: **Server-rendered HTML** with Go templates, reusing existing HTML structure and CSS/JS assets for perfect parity.

- Ensure that all approval/pending/retrieve states match current UI semantics.

### 5) Cutover and Rollback Strategy
- **Phase 1**: Freeze or dual-write, export data.
- **Phase 2**: Staging validation against the parity spec.
- **Phase 3**: Production cutover with snapshot and rollback plan.

## Open Questions / Decisions Needed
- Frontend direction (server-rendered vs SPA)?
    - server rendered
- Auth strategy (import Django hashes vs reset or SSO)?
    - answered above
- Target database for Go (Postgres vs other)?
    - postgres
    - sqlite
- Email delivery provider requirements?
    - scrap email completely
- Any changes desired in request cleanup timing or rotation semantics?
    - keep as-is
