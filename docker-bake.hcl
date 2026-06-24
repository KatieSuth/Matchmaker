# Multi-architecture build definition for publishing to Docker Hub.
#
# This file is layered on top of the Compose files so the build contexts,
# image names/tags, and build args are reused verbatim (single source of
# truth). It only adds the target platforms. Image names and tags still come
# from the same ${IMAGE}/${VERSION} variables (and root .env) used by Compose.
#
# Usage (preferred via the Makefile):
#   make build-multi   # build for all platforms into the build cache
#   make push-multi    # build and push the multi-arch manifest to the registry
#
# Equivalent raw command:
#   docker buildx bake \
#     -f docker-compose.yml -f docker-compose.prod.yml -f docker-bake.hcl --push

variable "PLATFORMS" {
  default = ["linux/amd64", "linux/arm64"]
}

# Target names match the Compose service names so these definitions merge with
# (rather than replace) the build config inherited from the Compose files.
target "api" {
  platforms = PLATFORMS
}

target "frontend" {
  platforms = PLATFORMS
}
