---
sidebar_position: 8
---

# User Management

Course Lecturers and PROMPT Administrators can manage who has Lecturer or Editor access to a course directly from the PROMPT UI, without needing access to the Keycloak admin console.

Keycloak stays the single source of truth for group membership - the User Management page is just a convenient front end on top of the same Keycloak groups.

---

## 1. Open the User Management Page

Open any course you are a Lecturer of, then click **User Management** in the course sidebar (right next to **Settings**).

> ⚠️ **Note**: If you do not see **User Management**, you do not have the necessary permissions. Only `Course Lecturer` and `PROMPT_Admin` can open this page. Course Editors and Students cannot.

The page shows two tables:

- **Lecturers** - full course administration, including team management.
- **Editors** - can edit course content but cannot manage the team.

---

## 2. Add a Lecturer or Editor

1. Click **Add Lecturer** or **Add Editor** on the corresponding card.
2. In the dialog that opens, type at least 2 characters into the search field. PROMPT searches Keycloak for matching users by name, email, or username.
3. Click **Add** next to the user you want to assign the role to.

Notes:

- A row labelled **Already a Lecturer** (or **Already an Editor**) cannot be added again to the same role; the button is disabled.
- A row labelled **Currently an Editor** / **Currently a Lecturer** means the user already has the *other* role. Adding them to a second role is allowed and does not remove the first.
- The change takes effect in Keycloak immediately. The table refreshes on its own once the request completes.
- If the search returns more matches than fit on screen, a **More results available** hint appears below the list. Narrow your query to find the right person.

---

## 3. Remove a Lecturer or Editor

1. Click the trash icon at the end of the row.
2. Confirm in the dialog that appears.

The row is removed from the Keycloak group immediately and the table refreshes.

> ⚠️ **You cannot remove yourself.** The trash icon on your own row is disabled and the server rejects any direct API attempt with HTTP 400. This is intentional: it prevents accidentally locking a course out of management access. If you need to step down as the sole Lecturer, ask another Lecturer to remove you, or use the Keycloak admin console.

---

## 4. Permissions Summary

| Action | Course Lecturer | Course Editor | PROMPT Admin |
| :--- | :---: | :---: | :---: |
| Open the User Management page | ✅ | ❌ | ✅ |
| Add or remove Lecturers | ✅ | ❌ | ✅ |
| Add or remove Editors | ✅ | ❌ | ✅ |
| Remove themselves | ❌ | n/a | ❌ |

---

## 5. Troubleshooting

- **The page shows an error or no users at all.** The Keycloak service account may be misconfigured. A PROMPT Administrator can verify this on the **System Status** page (see the [System Status](../admin/keycloak-prod#9-user-management-service-account) section in the admin docs).
- **A user I expect to find is not in the search results.** Make sure the user exists in Keycloak. Users are only searchable after their Keycloak account has been created (e.g. after their first login through an institutional identity provider).
- **An add or remove "succeeded" but I do not see the change.** Refresh the page. If the change still does not appear, check the Keycloak admin console directly - the User Management page reflects whatever Keycloak reports.
