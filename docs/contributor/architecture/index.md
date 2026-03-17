---
sidebar_position: 0
---

# Architecture

This section documents the system architecture of the PROMPT 2 platform.

The platform consists of:

- A React-based web client written in TypeScript
- A Golang-based core server
- Independently deployed course phase services, each with a microfrontend and, optionally, a backend microservice
- A Keycloak instance for identity and access management

Each chapter below describes a key architectural concern.

For detailed information about the architecture, see:

- [Subsystem Decomposition](./subsystem-decomposition.md)
- [Deployment Architecture](./deployment.md)
- [Access Control](./access-control.md)
- [Data Storage and Inter-Phase Communication](./data-storage.md)
- [File Storage (SeaweedFS)](./file-storage.md)
