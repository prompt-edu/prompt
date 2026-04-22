# Keycloak Local Development Guide

## 1. Local Setup (Step-by-Step)
1. **Launch Containers:** Use `docker-compose up -d keycloak keycloak-db`. 
2. **Auto-Import:** The system is configured to import `keycloakConfig.json` automatically on startup.
3. **Verify:** Access the UI at `http://localhost:8081`.

## 2. Pre-configured Test Users

The following accounts are included in `keycloakConfig.json` to facilitate testing of various permission levels within the PROMPT platform.

| Username | Password | University Login | Matriculation Number | Client Role (`prompt-server`) |
| :--- | :--- | :--- | :--- | :--- |
| **admin** | `admin` | `ad12min` | `00000001` | `PROMPT_Admin` |
| **lecturer** | `lecturer` | `le50ctu` | `00000002` | `PROMPT_Lecturer` |
| **course-lecturer** | `course-lecturer` | `co67lec` | `00000003` | `PROMPT_Course_Lecturer` |
| **course-editor** | `course-editor` | `co69edt` | `00000004` | `PROMPT_Course_Editor` |
| **student** | `student` | `no42tum` | `00000005` | `PROMPT_Student` |
| **student-passkey** | `student-passkey` | `no43tum` | `00000006` | `PROMPT_Student` |

## 3. Modifying Users and Roles
- **Adding Users:** Go to `Users` -> `Add user`. Required fields: Username, Email, First/Last Name.
- **Assigning Roles:** Go to `Users` -> pick user -> go to `Role Mapping` tab and `Assign role`.

## 4. Manual Client Mapper Configuration
If you need to manually add mappers for `prompt-server`:
1. Go to `Clients` -> `prompt-server` -> `Client scopes`.
2. Click `prompt-server-dedicated` -> `Add mapper` -> `By configuration`.
3. Choose `User Attribute`.
4. Example (University Login):
   - Name: `university_login`
   - User Attribute: `university_login`
   - Token Claim Name: `university_login`

## 5. Troubleshooting & Reset
- **Reset to Default:** Run `docker-compose down -v` and `rm -rf keycloak_postgres_data`. This wipes the DB and re-imports the JSON on next Keycloak start.
- **401 Unauthorized:** Check if `KEYCLOAK_CLIENT_SECRET` in `.env.dev` matches the one in Keycloak UI.
- **Passkey Issues:** Ensure `Resident Key` is `Required` in WebAuthn Passwordless Policy.