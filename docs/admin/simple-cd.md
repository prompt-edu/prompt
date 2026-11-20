---
title: Deploying with simple-cd
sidebar_position: 70
---

# Deploying with simple-cd

[simple-cd](https://github.com/Mtze/simple-cd) deploys PROMPT to a VM as a Docker Compose target, next to the
GitHub deploy workflows (`deploy-docker.yml`), which stay unchanged.

## How it works

- `.cd.yaml` in the repository root names `docker-compose.prod.yml` as the Compose file.
- The `publish-cd` job in `dev.yml` runs on every pull request and every push to `main` that builds images.
  It looks up the digest of each image the build pushed and publishes them with `Mtze/simple-cd/cd-publish` as
  `deploy.json`.
  simple-cd only picks it up when the whole run succeeds.
- Each image in `deploy.json` is named after its service in `docker-compose.prod.yml`, for example `server-core`
  or `client-assessment-component`.
  On a deploy, simple-cd writes `.cd.env` with `SERVER_CORE_IMAGE=ghcr.io/...@sha256:...` and so on.
- `docker-compose.prod.yml` uses `${<SERVICE>_IMAGE:-<image>:${<...>_IMAGE_TAG}}` for every PROMPT image.
  simple-cd sets `<SERVICE>_IMAGE`, the GitHub deploy sets only the `*_IMAGE_TAG` variables, so both work with the
  same file.
- simple-cd uploads the file to the target directory as `docker-compose.yml` and runs
  `docker compose --env-file .env --env-file .cd.env pull` and `up -d --remove-orphans` there.
  Everything else stays on the VM.

When you add a service with a PROMPT image, add it to `docker-compose.prod.yml` with the same variable pattern and
to the `IMAGES` list of the `publish-cd` job.

## Differences to the GitHub deploy

- The services are not stopped before the update, so the maintenance page does not show during a deploy.
- Old images are not pruned.
- Turning AI off (`COMPOSE_PROFILES` without `ai`) does not stop `server-ai`, `db-ai` and `client-ai-component`;
  stop them by hand once.
- The version shown on the System Status page comes from `GITHUB_SHA`, `GITHUB_REF` and the `*_IMAGE_TAG`
  variables in the VM's `.env`; simple-cd does not update them.

## Prepare a VM

Use a VM of its own, never one the GitHub deploy writes to.

1. Install Docker Engine with the Compose plugin.
2. Create the deploy user with a POSIX login shell and add it to the `docker` group:
   ```sh
   sudo useradd --create-home --shell /bin/sh deploy
   sudo usermod -aG docker deploy
   ```
3. Add the public key of the simple-cd key pair to `/home/deploy/.ssh/authorized_keys`.
4. Allow SSH on port 22 from the cluster that runs simple-cd. simple-cd connects directly; it cannot use the
   deployment gateway as a jump host.
5. Create the target directory, for example `/opt/prompt`, owned by `deploy`.
6. Create the Docker network: `docker network create prompt-network`.
7. Copy the `maintenance/` directory of this repository into the target directory; the `maintenance` service mounts
   it.
8. Create `.env` in the target directory with the variables `deploy-docker.yml` writes to `.env.prod` (see
   [Production Setup](productionSetup.md)). Leave out the `*_IMAGE_TAG` variables or set them for the version
   display only. Set `COMPOSE_PROFILES=ai` together with `AI_ENABLED=true`.
9. Point DNS for `CORE_HOST` and `s3.<CORE_HOST>` at the VM and open ports 80 and 443 for Let's Encrypt.
10. Record the host key, compare its fingerprint with the one on the VM, and store it in Vault together with the
    private key:
    ```sh
    ssh-keyscan -t ed25519 <vm-host> > known_hosts
    ssh-keygen -lf known_hosts
    vault kv put secret/cd/ssh/prompt-dev private_key=@id_ed25519 known_hosts=@known_hosts user=deploy
    ```

## simple-cd configuration

The target in the simple-cd target configuration:

```yaml
targets:
  - id: prompt-dev
    project: prompt-edu/prompt
    type: compose
    lockable: true
    host: <vm-host>
    ssh: vault:secret/cd/ssh/prompt-dev
    dir: /opt/prompt
    jobTimeout: 15m
```

The Vault secret `secret/cd/ssh/prompt-dev` holds `private_key`, `known_hosts` and `user` (default `deploy`).
The simple-cd GitHub App must be installed on `prompt-edu/prompt`, and the Keycloak roles
`project:prompt-edu/prompt:developer` and `project:prompt-edu/prompt:maintainer` must exist.
