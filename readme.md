# PROMPT 2.0: A Modular and Scalable Management Platform for Project-Based Teaching

[![Discord](https://img.shields.io/discord/1345423854160445580?style=for-the-badge&logo=discord&logoColor=white&label=Join%20our%20Community)](https://discord.gg/eybNUqD8gf)

## What is PROMPT?

**PROMPT (Project-Oriented Modular Platform for Teaching)** is a course management tool built to simplify the administration of project-based university courses.

Originally developed for the **iPraktikum** at the **Technical University of Munich**, PROMPT is now a **modular and extensible platform** that adapts to the needs of instructors and scales across different teaching formats.

### Key Features

#### Core Features

- **Course Configuration:** Build custom course flows using reusable course phases.
- **Student Management:** Manage participant data and applications.
- **Application Phase:** Streamlined workflows for student applications.

#### Dynamically Loaded Course Phases

PROMPT supports custom course phases as independently deployable modules:

- 🗓 **Interview Phase** – Schedule and manage student interviews.
- 🧑‍🤝‍🧑 **Team Phase** – Allocate students to projects and teams.
- 📄 **TUM Matching Export** – Export data in TUM-compatible format.
- 🧩 **Custom Phases** – Easily extend PROMPT with your own logic.

---

## Guide to PROMPT

Explore the live system here 👉 [https://prompt.aet.cit.tum.de/](https://prompt.aet.cit.tum.de/)

PROMPT is built for teaching teams who run complex project-based university courses. This guide gives you a first impression of the product, its modules, and typical workflows.

### Build and Configure Courses

PROMPT lets you define courses by combining reusable “phases”:

- Add application, interview, matching, and many more phases
- Define prerequisites and outcomes for each phase

![Screenshot: Course Configuration](docs/static/img/course-configurator.png)

---

### Manage Students and Teams

PROMPT provides powerful tools for handling student participation throughout your course:

- Track incoming applications and monitor their progress
- Schedule and conduct interviews with applicants
- Form teams and assign projects using manual or semi-automated workflows
- Evaluate student performance on a **per-phase basis**, enabling you to decide who passes or fails at each stage of the course independently

![Screenshot: Application Administration](docs/static/img/application-administration.png)

![Screenshot: Matching](docs/static/img/matching.png)

---

### Extend PROMPT with Custom Phases

Each course phase can be developed and plugged in independently—ideal for institutions with evolving or specialized workflows.

---

## Development Guide

This section is for **developers and contributors** looking to run PROMPT locally or contribute to its development.

### Project Structure

#### Clients

- Built with **React**, **TypeScript**, and **Webpack Module Federation**
- Micro-frontends for each course phase
- Shared design system using [`shadcn/ui`](https://ui.shadcn.com/)

#### Server

- Developed in **Go**
- Core service + modular service architecture for each course phase
- **PostgreSQL** for data storage
- **Keycloak** for authentication

---

### Getting Started

#### Prerequisites

| Tool | Version | Installation |
| ---- | ------- | ------------ |
| **Go** | 1.26+ | [go.dev/doc/install](https://go.dev/doc/install) |
| **Node.js** | 22.10.0 - 22.x | [nodejs.org](https://nodejs.org/) |
| **Docker** | Latest | [docker.com](https://www.docker.com/) |
| **golang-migrate** | Latest | [github.com/golang-migrate/migrate](https://github.com/golang-migrate/migrate) |
| **sqlc** | Latest | [docs.sqlc.dev](https://docs.sqlc.dev/en/latest/overview/install.html) |

#### 1. Enable Corepack (for Yarn 4)

```bash
# macOS with Homebrew (Homebrew strips corepack from Node)
brew install corepack

# Other platforms (included with Node 16.9+)
npm install -g corepack

# Enable corepack
corepack enable
```

#### 2. Clone & Configure Environment

```bash
git clone https://github.com/ls1intum/prompt2.git
cd prompt2

# Create environment files from templates
cp .env.template .env
cp .env.dev.template .env.dev  # Local development overrides
```

The `.env` file contains Docker/production configuration. The `.env.dev` file contains local development overrides (localhost instead of Docker hostnames) and is loaded by `make server`.

#### 3. Start Database & Keycloak

```bash
docker compose up -d db keycloak
```

#### 4. Configure Keycloak (first time only)

1. Navigate to [http://localhost:8081](http://localhost:8081)
2. Login with username `admin` and password `admin`
3. In the top-left dropdown, select **Create Realm** and upload `keycloakConfig.json`
4. In the new Prompt realm, create a user:
   - Go to **Users** → **Add user**
   - Fill in **all required fields** (these are needed for token claims):
     - Username
     - Email (required)
     - First name (required)
     - Last name (required)
   - Click **Join Groups** and select `PROMPT-Admins`
   - Go to **Credentials** tab and set a password
   - Go to **Attributes** tab and add:
     - `university_login` → `ab12cde`
     - `matriculation_number` → `01234567`
5. Configure client mappers (to include user attributes in tokens):
   - Go to **Clients** → `prompt-server` → **Client scopes** tab
   - Click `prompt-server-dedicated`
   - Click **Add mapper** → **By configuration** → **User Attribute**
   - Create mappers for each attribute:

     | Name | User Attribute | Token Claim Name | Claim JSON Type |
     | ---- | -------------- | ---------------- | --------------- |
     | university_login | university_login | university_login | String |
     | matriculation_number | matriculation_number | matriculation_number | String |

   - Ensure **Add to ID token** and **Add to access token** are enabled
6. Generate a client secret:
   - Go to **Clients** → `prompt-server` → **Credentials**
   - Click **Regenerate** and copy the secret
   - Paste the secret into your `.env.dev` and `.env` files as `KEYCLOAK_CLIENT_SECRET`

#### 5. Start the Server

The `.env` file is not automatically loaded by Go. Use one of these methods:

1. **Use Make** (recommended, works in any shell):

   ```bash
   make server
   ```
2. **Manual export** (bash/zsh only):

   ```bash
   set -a && source .env.dev && set +a
   cd servers/core && go run main.go
   ```

#### 6. Start the Clients

```bash
make clients
# or manually:
cd clients && yarn install && yarn run dev
```

This launches all micro-frontends simultaneously using Lerna. The app runs at [http://localhost:3000](http://localhost:3000).

To start a specific micro-frontend only, navigate to its subdirectory and run `yarn run dev` there.

---

### Useful Commands

```bash
# Generate sqlc code (after changing SQL queries)
cd servers/<service> && sqlc generate

# Run Go tests
cd servers/<service> && go test ./...

# Lint clients
cd clients && yarn eslint "core" --config "core/eslint.config.mjs"

# Add shadcn/ui component
cd clients/shared_library && yarn dlx shadcn add <component-name>
```

### API Spec Generation (Swagger)

We generate and commit swagger specs for the Go servers. To avoid forgetting this, install the repo-managed git hooks:

```bash
./scripts/install-githooks.sh
```

When you commit changes under `servers/core/` or `servers/template_server/`, the pre-commit hook regenerates and stages the swagger docs. Ensure `swag` is available on your PATH:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

### Ports Reference

| Service | Port |
| ------- | ---- |
| Core Client | 3000 |
| Keycloak | 8081 |
| Core Server | 8080 |
| PostgreSQL | 5432 |

---

## 💬 Community & Support

We are building a community around PROMPT! Whether you are an instructor using the platform or a developer looking to contribute, we'd love to have you.

- **Discord**: Join our [Discord Server](https://discord.gg/eybNUqD8gf) for real-time discussion and support.
- **GitHub Issues**: Report bugs or request features on our [Issue Tracker](https://github.com/ls1intum/prompt2/issues).
- **Documentation**: Visit our [Documentation Portal](https://ls1intum.github.io/prompt2/) for detailed guides.

## 🤝 Contributing

We welcome contributions of all kinds! Please see our [Contributing Guidelines](./CONTRIBUTING.md) to get started.

---

## Configuration

PROMPT can be customized for your course by composing it from different modular phases. These configurations are handled at the course level and dynamically loaded on demand.
