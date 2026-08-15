# This is the base dockerfile. Here the base image is pulled and the base setup is done for the project.
# Make sure to include the base setup for lerna here.
#
# Layering: everything yarn needs to resolve and install is copied first, so
# `yarn install` only re-runs when a manifest or the lockfile changes. A source
# edit then only invalidates the final `COPY . ./`.
# Requires BuildKit (Docker 23+) for the cache mount below.
FROM node:24.19.0 AS base

RUN corepack enable
RUN yarn config set httpTimeout 300000 --home

# Yarn 4 defaults to enableGlobalCache, which puts downloads in
# ${globalFolder}/cache. Pin globalFolder so the cache mount below provably
# targets the directory yarn actually writes to.
ENV YARN_GLOBAL_FOLDER=/root/.yarn/berry

WORKDIR /app

# --- Dependency layer: manifests and lockfile only ---------------------------
COPY package.json yarn.lock .yarnrc.yml ./

# One line per entry in the `workspaces.packages` array of package.json. These
# cannot drift silently: that array is an explicit list rather than a glob, so a
# workspace added there but not here makes `yarn install` fail.
COPY core/package.json                           core/package.json
COPY interview_component/package.json            interview_component/package.json
COPY matching_component/package.json             matching_component/package.json
COPY example_component/package.json              example_component/package.json
COPY assessment_component/package.json           assessment_component/package.json
COPY team_allocation_component/package.json      team_allocation_component/package.json
COPY self_team_allocation_component/package.json self_team_allocation_component/package.json
COPY certificate_component/package.json          certificate_component/package.json

RUN --mount=type=cache,target=/root/.yarn/berry/cache,sharing=locked \
    yarn install --immutable

# --- Source layer ------------------------------------------------------------
# COPY merges into the existing tree and never deletes, and .dockerignore
# excludes node_modules, so the install above survives this copy.
COPY . ./
