# Docker & CI/CD Standards

> Applies to: Dockerfiles, `docker-compose*.yml`, and `.github/workflows/**`.

- Multi-stage builds are required for production Dockerfiles.
- In compose files, health checks are required for core dependencies and app services that gate startup (for example, DB/API). Health checks for edge/proxy services are optional unless orchestration depends on them.
- In GitHub Actions workflows, pin every third-party action to a full commit SHA with the version as a trailing comment (e.g. `actions/checkout@<sha> # v7.0.1`). Dependabot (`.github/dependabot.yml`) keeps the pins current.
- Pin base images in production Dockerfiles by digest (`image:tag@sha256:...`) and hard-code the tag in `FROM` rather than interpolating an `ARG`, so Dependabot can update it.
- **Networking:** Ensure compose service names match upstream targets defined in the active reverse-proxy or gateway configuration.

