---
title: "Production Setup"
---

# Production Setup

The general setup is described in the [Contributor Setup Guide](/contributor/setup). This guide focuses on the dockerized setup process for PROMPT. The PROMPT setup process is straightforward and requires only a few environment variables to adjust.

There are three crucial steps for setting up PROMPT:

1. **Keycloak Setup (required)**
2. **SMTP Server Setup (optional – recommended)**
3. **Container Setup**

---

## 1. Keycloak Setup

PROMPT uses Keycloak for authentication and user rights management. In particular, PROMPT requires the following Keycloak components:

### Clients

- **`prompt-client`**  
  Used for login on the client side.

- **`prompt-server`**  
  Used by the backend server to authenticate and manage user groups and roles.

### Roles and Groups

- **General Roles** (client-specific roles for `prompt-server`):
  - **`PROMPT_ADMIN`**: Grants access to all courses and functionality.
  - **`PROMPT_LECTURER`**: Allows the creation of new courses.

- **Course-Specific Roles**  
  When a course is created, PROMPT automatically creates the following client roles for `prompt-server`:
  - **`<SemesterTag>-<CourseName>`** (Group and Role)
    - **`<SemesterTag>-<CourseName>-Lecturer`**  
      Group: `Lecturer` (a subgroup of the course group). This role grants admin rights on the course. The course creator is automatically added to this role.
    - **`<SemesterTag>-<CourseName>-Editor`**  
      Group: `Editor` (a subgroup of the course group). This role has limited user rights. For more details, refer to the [Contributor Guide](/contributor/setup).

There are two options for setting up Keycloak:

### Option 1: Configure an Existing Keycloak Instance for PROMPT

Follow these steps:

#### 1. Create the Clients `prompt-client`

1. **Set URLs**  
   Add your client domain to the following fields. All URLs must be prefixed with `https`.
   - **Root URL**
   - **Home URL**
   - **Valid Redirect URLs** (must be postfixed with `/*`)
   - **Valid Logout URLs** (must be postfixed with `/*`)
   - **Web-Origin URLs**
     IMPORTANT: Only postfix Valid Redirect and Logout URLs.
2. **Add Mappers for User Attributes**  
   PROMPT authenticates students using their `matriculation_number` and `university_login`. To include these in the token:
   - Navigate to **Client Scopes** for `prompt-client`.
   - Select the `prompt-client-dedicated` scope.
   - Click on **Add Mapper by configuration**.
   - For each mapper, set:
     - **Mapper Type:** User Attribute
     - **User Attribute:** `university_login` or `matriculation_number`  
       (adjust if your system uses a different name, e.g., TUM ID)
     - **Token Claim Name:** `university_login` or `matriculation_number` (do **not** change)
     - **Add to Access Token:** Enabled

#### 2. Create the Clients `prompt-server`

1. **Set URLs**  
   Add your server domain to:
   - **Root URL**
   - **Home URL**
   - **Valid Redirect URI**
   - **Valid Post Logout Redirect URI**

   Each URL must start with `https` and ALL must be postfixed with `/*` (note that the postfixes might differ from those in `prompt-client`).

2. **Enable Service Account Access**  
   This is required to allow PROMPT to create course-specific roles:
   - Enable **Service Account Roles**.
   - Under the **Service Account Roles** tab, add the role `realm-admin`.
   - In the **Credentials** tab, select **Client Authenticator: Client ID and Secret**.
     - Save and generate a new secret.
     - Store this secret in your environment file under `KEYCLOAK_CLIENT_SECRET`.

3. **Add the required client roles**
   - Add the roles `PROMPT_ADMIN` and `PROMPT_LECTURER`
   - Optionally create groups for these roles for easier user role management.

---

### Option 2: Use a Standalone Keycloak Instance for PROMPT

For this option, refer to the [Contributor Setup Guide](/contributor/setup) on how to configure a standalone Keycloak instance and the `keycloakConfig.json`.

Keep in mind:

- The `docker-compose.extern.yml` includes a sample setup with a Docker container for Keycloak and a database instance to persist Keycloak data.
- **Domain Setup:**  
  Configure a separate domain for your Keycloak instance (this can be a subdomain but **not** a subpath).
  - **Valid Examples:**
    - `prompt.<yourDomain>.de` and `keycloak.prompt.<yourDomain>.de`
  - **Invalid Example:**
    - `prompt.<yourDomain>.de` and `prompt.<yourDomain>.de/keycloak`
- Follow the Contributor Guide closely and copy the `KEYCLOAK_CLIENT_SECRET` into your environment file as described.
- To support student login, your Keycloak instance must allow authentication with `matriculation_number` and `university_login` attributes. Without these, student-specific management console features will not be available (although this does not affect the application module).

---

## 2. SMTP Server Setup (Optional – Recommended)

PROMPT integrates with a mail service to, for example, send confirmation emails to students or enable instructors to send emails to all accepted/declined students. The AET Chair uses a Postfix container (refer to `docker-compose.prod.yml`).

You can use any SMTP server by adjusting your environment file with the corresponding SMTP settings (see the details in section [3.1 Adjust Environment Variables](#adjust-environment-variables)).

---

## 3. Container Setup

### 3.1 Adjust Environment Variables

An `.env.template` file is provided in the repository. This file includes all the runtime variables required for deployment. Below is an explanation of the variables:

#### Deployment Location Specific Variables

- **`ACME_EMAIL`**  
  Email used for HTTPS certificate generation and notifications.

- **`ENVIRONMENT`**  
  Set to `production`. Changing this will result in additional debug output.

- **`CORE_HOST`**  
  The deployment URL of your application.

- **`CHAIR_NAME_SHORT`**  
  The short name displayed in the website header.

- **`CHAIR_NAME_LONG`**  
  Used in the welcome text (can be the same as `CHAIR_NAME_SHORT`).

#### Dockerized Environment Variables

- **`SERVER_ADDRESS`**  
  Defaults to `0.0.0.0:8080`. Change only if you are not deploying in a dockerized environment.

- **`DB_HOST`**  
  Defaults to `db`. Do not change if you are using the dockerized environment.

- **`DB_PORT`**  
  Defaults to `5432`. Do not change if you are using the dockerized environment.

- **`DB_NAME`**  
  Defaults to `prompt`.

- **`DB_USER`**  
  Defaults to `prompt-postgres`.

- **`DB_PASSWORD`**  
  Defaults to `prompt`.

#### Keycloak Variables

Adjust these if deploying Keycloak on a different host:

- **`KEYCLOAK_DB_USER`**  
  Defaults to `keycloak`.

- **`KEYCLOAK_DB_PASSWORD`**  
  Defaults to `keycloak`.

- **`KEYCLOAK_HOST`**  
  Example: `keycloak.ase.in.tum.de`.

- **`KEYCLOAK_REALM_NAME`**  
  Defaults to `prompt`.

- **`KEYCLOAK_CLIENT_ID`**  
  Defaults to `prompt-server`.

- **`KEYCLOAK_CLIENT_SECRET`**  
  Generate this as described in the Keycloak setup section.

- **`KEYCLOAK_ID_OF_CLIENT`**  
  Example: `a584ca61-fa83-4e95-98b6-c5f3157ae4b4` (change if not using pre-configuration).

- **`KEYCLOAK_AUTHORIZED_PARTY`**  
  Defaults to `prompt-client`.

- **`KEYCLOAK_ADMIN`**  
  Defaults to `admin`.

- **`KEYCLOAK_ADMIN_PASSWORD`**  
  Defaults to `admin`.

#### Deployment Version Variables

- **`SERVER_IMAGE_TAG`**
- **`CORE_IMAGE_TAG`**
- **`TEMPLATE_IMAGE_TAG`**
- **`INTERVIEW_IMAGE_TAG`**
- **`MATCHING_IMAGE_TAG`**

Set these according to your deployment version (default is `main`).

#### SMTP Server Variables

- **`SENDER_EMAIL`**  
  Example: `prompt@ase.cit.tum.de`

- **`SENDER_NAME`**  
  Example: `AET Mailing Service`

- **`SMTP_HOST`**  
  Example: `postfix`

- **`SMTP_PORT`**  
  Defaults to `25`

- **`SMTP_USERNAME`** (Optional)  
  Username for SMTP authentication. Leave empty if your SMTP server doesn't require authentication.

- **`SMTP_PASSWORD`** (Optional)  
  Password for SMTP authentication. Leave empty if your SMTP server doesn't require authentication.

#### File Storage (S3-Compatible) Variables

PROMPT now stores uploaded files in an S3-compatible bucket (SeaweedFS S3 gateway, AWS S3, MinIO, etc.). The storage service uses presigned URLs, so you must configure both internal and public endpoints.

- **`S3_BUCKET`**  
  Bucket name for uploaded files (e.g., `prompt-files`).

- **`S3_REGION`**  
  Region for your S3-compatible backend. Use `us-east-1` for SeaweedFS.

- **`S3_ENDPOINT`**  
  Internal endpoint used by the server to reach the S3 API.  
  Example (SeaweedFS S3 gateway): `http://seaweedfs-s3:8333`.

- **`S3_PUBLIC_ENDPOINT`**  
  Public endpoint used in presigned URLs that clients access.  
  Example (production): `https://s3.<your-domain>`  
  Example (local): `http://localhost:8334`.

- **`S3_ACCESS_KEY`** / **`S3_SECRET_KEY`**  
  Credentials for the S3 API. Used by both the core server and the SeaweedFS S3 gateway. The access key acts as an identifier (e.g., `prompt-s3-user`), while the secret key should be a strong random value.

- **`S3_FORCE_PATH_STYLE`**  
  Set to `true` for SeaweedFS/MinIO. Set to `false` for AWS S3.

- **`S3_PRESIGN_UPLOAD_TTL_SECONDS`**  
  Presigned upload URL TTL in seconds (default: `60`).

- **`S3_PRESIGN_DOWNLOAD_TTL_SECONDS`**  
  Presigned download URL TTL in seconds (default: `30`).

- **`S3_PRESIGN_TTL_SECONDS`** (Legacy optional)  
  Fallback TTL used only if the specific upload/download TTL variables are not set.

- **`MAX_FILE_UPLOAD_SIZE_MB`**  
  Maximum allowed file size for uploads (default: `50`).

- **`ALLOWED_FILE_TYPES`**  
  Comma-separated list of allowed MIME types. Leave empty to allow all.

---

### 3.2 Select the Appropriate Docker Compose File

- **`docker-compose.prod.yml`**  
  Does **not** include a Keycloak container. Use this file if you have an already running Keycloak instance.

- **`docker-compose.extern.prod.yml`**  
  Includes a Keycloak container. Refer to the Keycloak configuration above for details.

### 3.3 File Bucket Notes

- The S3-compatible adapter checks for the bucket at startup and will attempt to create it if it does not exist (when the backend supports it).
- For production, ensure your DNS and TLS are set up for the `S3_PUBLIC_ENDPOINT` so presigned URLs are reachable by clients.

---

### 3.4 Start the Docker Containers

To start the docker containers, run the following command (adjust `<path to your env file>` as needed):

```bash
docker compose -f docker-compose.prod.yml --env-file=<path to your env file> up --pull=always -d
```
