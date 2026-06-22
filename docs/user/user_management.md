---
sidebar_position: 8
---

# User Management

Course Lecturers and PROMPT Administrators can manage who has Lecturer or Editor access to a course directly inside PROMPT, without asking an administrator to do it for them.

---

## 1. Open the User Management Page

Open any course you are a Lecturer of, then click **User Management** in the course sidebar (right next to **Settings**).

{/* TODO: screenshot of the course sidebar with the User Management entry highlighted */}


> ⚠️ **Note**: If you do not see **User Management**, you do not have the necessary permissions. Only `Course Lecturer` and `PROMPT_Admin` can open this page. Course Editors and Students cannot.

The page shows two tables:

- **Lecturers** - full course administration, including team management.
- **Editors** - can edit course content but cannot manage the team.

---

## 2. Add a Lecturer or Editor

1. Click **Add Lecturer** or **Add Editor** on the corresponding card.
2. Type at least 2 characters in the search field. PROMPT looks up matching users by name, email, or username.
3. Click **Add** next to the user you want to assign the role to.

{/* TODO: screenshot of the Add User dialog with search results visible */}

Notes:

- A row labelled **Already a Lecturer** (or **Already an Editor**) cannot be added again to the same role - the button is disabled.
- A row labelled **Currently an Editor** / **Currently a Lecturer** means the user already has the *other* role. Adding them to a second role is allowed and does not remove the first.
- The change takes effect immediately. The table refreshes on its own after a moment.
- If the search has more matches than fit on screen, a **More results available** hint appears below the list. Narrow your query to find the right person.

---

## 3. Remove a Lecturer or Editor

1. Click the trash icon at the end of the row.
2. Confirm in the dialog that appears.

The row is removed from the Keycloak group immediately and the table refreshes.

> ⚠️ **You cannot remove yourself.** The trash icon on your own row is disabled, and PROMPT will block any other attempt to remove you. This is intentional: it prevents you from accidentally locking your course out of all management access. If you need to step down as the sole Lecturer of a course, ask another Lecturer to remove you, or contact a PROMPT Administrator.

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

- **The page shows an error or no users at all.** This usually means PROMPT cannot reach the user database. Ask a PROMPT Administrator to check the configuration - they can verify it via the **System Status** page (see the [admin guide](../admin/keycloak-prod#9-user-management-service-account)).
- **A user I expect to find is not in the search results.** Make sure the user has logged into PROMPT (or the institutional login system) at least once. People who have never logged in do not have an account yet and cannot be found.
- **I added or removed a user but the table did not update.** Refresh the page. If the change still does not appear, the assignment may have failed - try again, or ask a PROMPT Administrator to check.
