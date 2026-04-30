# This is the base dockerfile. Here the base image is pulled and the base setup is done for the project.
# Make sure to include the base setup for lerna here.
FROM node:24.15.0 AS base

RUN corepack enable
RUN yarn config set httpTimeout 300000 --home

WORKDIR /app

COPY . ./

RUN yarn install --immutable
