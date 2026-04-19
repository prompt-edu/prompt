This document provides instructions for deploying, securing, and maintaining Keycloak in a production environment for the **PROMPT 2.0** platform.

---

## 1. Production Configuration Guide
When moving from development to production, Keycloak must be optimized for the Quarkus-based distribution.

### Core Environment Variables
| Variable | Value | Description |
| :--- | :--- | :--- |
| `KC_PROXY_HEADERS` | `xforwarded` | Trust `X-Forwarded-*` headers from a TLS-terminating reverse proxy (Nginx/Traefik). |
| `KC_HOSTNAME` | `sso.yourdomain.com` | Strict hostname for security and redirect validation. |
| `KC_DB` | `postgres` | Database vendor. |
| `KC_HTTP_ENABLED` | `true` | Keep HTTP enabled behind the proxy; TLS is terminated at the reverse proxy. |

---

## 2. Security Best Practices
Protecting the Identity Provider is the foundation of PROMPT's security.

### Hardening Checklist
* **Change Default Credentials:** Immediately rotate the `KEYCLOAK_ADMIN_PASSWORD` and `KEYCLOAK_CLIENT_SECRET`.
* **SSL/TLS Setup:** Ensure the reverse proxy uses TLS 1.3 and strong cipher suites. Keycloak should never be exposed directly to the internet via HTTP.
* **Brute Force Protection:** * Navigate to **Realm Settings** -> **Security Defenses**.
    * Enable **Brute Force Detection**.
    * Set **Max Login Failures** (e.g., 5) and **Wait Increment** to prevent automated password guessing.
* **Master Realm Isolation:** Use the `master` realm *only* for managing Keycloak itself. All PROMPT users and clients must reside in the `prompt` realm.

---

## 3. Role and Permission Management

PROMPT uses a hybrid Role-Based Access Control (RBAC) model consisting of static global roles and dynamic course-level roles.

### A. Global System Roles
These roles are static and grant high-level permissions across the entire platform. Assign these directly in Keycloak:

| Role Name | Description |
| :--- | :--- |
| `PROMPT_Admin` | Full system orchestration, global settings, and user management. |
| `PROMPT_Lecturer` | Global teaching authority; allowed to create new courses and manage departments. |

### B. Course-Specific Roles
Permissions for specific courses are validated using a dynamic naming convention. The backend expects roles to be formatted exactly as:
`<semester>-<course_name>-<role>`

**Examples:**
- `Winter2024-Algorithms101-Lecturer`
- `Summer2025-MobileAppDev-Editor`
- `Winter2024-Algorithms101-Student`

> [!CAUTION]
> **Do not use the `PROMPT_` prefix for course-level roles.** If a course is named "IntroToAI" in the "Summer2026" semester, the student role **must** be `Summer2026-IntroToAI-Student`. Any other format will result in an authorization failure.

### C. Role Assignment Strategy
1. **Global Roles:** Assign to permanent staff via **Role Mapping**.
2. **Course Roles:** These are typically created and assigned dynamically during the course enrollment process. If manual setup is required, ensure the string matches the `<semester>-<course_name>-<role>` pattern defined in the backend configuration.
---

## 4. User Provisioning and Group Management
### Provisioning Strategies
1. **Manual:** Via the Admin Console (**Users** -> **Add user**).
2. **Federated:** Automatic syncing from LDAP/AD (recommended for universities).
3. **Self-Registration:** Can be enabled in **Realm Settings** -> **Login**, but usually discouraged for managed courses.

### Group Management
Use **Groups** to apply roles in bulk. For example, create a group `PROMPT-Admins` and map the `PROMPT_Admin` role to it. Every user added to this group inherits the role automatically.

---

## 5. Client Mapper Configuration
The Go backend requires specific user attributes to be present in the JWT token.

### Required Custom Mappers
Configure these in **Clients** -> **prompt-server** -> **Client scopes** -> **prompt-server-dedicated**:

| Name | Mapper Type | User Attribute | Token Claim Name |
| :--- | :--- | :--- | :--- |
| `university_login` | User Attribute | `university_login` | `university_login` |
| `matriculation_number` | User Attribute | `matriculation_number` | `matriculation_number` |

> **Note:** Without these mappers, the backend will return `401 Unauthorized` or fail to link the user to the database.

---

## 6. Realm Configuration Options
* **Themes:** Set custom themes for Login, Account, and Admin pages in **Realm Settings** -> **Themes**.
* **Email:** Configure a valid SMTP server in **Realm Settings** -> **Email** to support "Forgot Password" and email verification flows.
* **Tokens:** Adjust **Access Token Lifespan** (recommended: 5-15 minutes) and **SSO Session Max** to balance security and user experience.

---

## 7. Integration with External Identity Providers (IDP)
PROMPT is designed to integrate with institutional authentication systems.

### LDAP / Active Directory
1. Go to **User Federation** -> **Add provider** -> **ldap**.
2. Map institutional `uid` or `sAMAccountName` to Keycloak's `university_login`.
3. Set **Sync Registrations** to keep Keycloak in sync with the university directory.

### SAML 2.0 / OIDC
1. Go to **Identity Providers** -> **Add SAML v2.0 provider**.
2. Import metadata from your institution (e.g., Shibboleth).
3. Use **Mappers** to extract the user's login and roles from the SAML assertion.

---

## 8. Backup and Restore Procedures
### Database Backup (PostgreSQL)
Run a daily cron job to dump the Keycloak database:
```bash
docker exec prompt-keycloak-db \
  pg_dump -U "${KEYCLOAK_DB_USER}" "${KEYCLOAK_DB_NAME}" \
  > backup_$(date +%F).sql
 ```

### Database Restore (PostgreSQL)
Stop Keycloak, drop and recreate the database, then restore:
```bash
docker compose stop keycloak
docker exec -i prompt-keycloak-db \
  psql -U "${KEYCLOAK_DB_USER}" -d postgres \
  -c "DROP DATABASE IF EXISTS ${KEYCLOAK_DB_NAME};" \
  -c "CREATE DATABASE ${KEYCLOAK_DB_NAME};"
docker exec -i prompt-keycloak-db \
  psql -U "${KEYCLOAK_DB_USER}" -d "${KEYCLOAK_DB_NAME}" < backup_YYYY-MM-DD.sql
docker compose start keycloak
```