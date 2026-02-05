target "postgres16-cloud-native" {
  dockerfile-inline = <<EOT
ARG BASE_IMAGE
FROM ghcr.io/cloudnative-pg/postgresql:16-standard-bookworm
USER root
RUN apt-get update && \
    apt-get install -y --no-install-recommends postgresql-16-pgvector && \
    rm -rf /var/lib/apt/lists/*
USER 26
EOT

  args = {
    BASE_IMAGE = "ghcr.io/cloudnative-pg/postgresql:16-standard-bookworm"
  }

  tags = [
    "postgres:16-cloud-native"
  ]
}
