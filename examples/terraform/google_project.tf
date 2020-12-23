terraform {
  required_version = ">= 0.12.29"
}

provider "google" {
  version = "~> 3.49.0"
}

provider "google-beta" {
  version = "~> 3.51.0"
}

resource "google_project" "project" {
  name            = var.project_id
  project_id      = var.project_id
  billing_account = var.billing_account_id
}

resource "google_project_service" "cloudbuild" {
  service = "cloudbuild.googleapis.com"
  project = google_project.project.project_id
}

resource "google_project_service" "artifactregistry" {
  service = "artifactregistry.googleapis.com"
  project = google_project.project.project_id
}

resource "google_project_service" "run" {
  service = "run.googleapis.com"
  project = google_project.project.project_id
}

resource "google_artifact_registry_repository" "containers" {
  provider = google-beta

  location      = var.location
  repository_id = "containers"
  format        = "DOCKER"
  project       = google_project.project.project_id

  depends_on = [google_project_service.artifactregistry]
}

resource "google_cloudbuild_trigger" "build" {
  provider = google-beta

  name        = "build"
  description = "build and push container image"
  project     = google_project.project.project_id

  github {
    owner = var.github_owner
    name  = var.github_name
    push {
      branch = "main"
    }
  }

  filename = "cloudbuild.yaml"
  substitutions = {
    _LOCATION = var.location
  }

  ignored_files = [
    ".gitignore",
    "LICENSE",
    "deploy.sh",
    "README.md",
    "**/*.md",
    "**/*_test.go",
    "**/renovate.json",
    "examples/*",
  ]

  depends_on = [google_project_service.cloudbuild]
}

resource "google_app_engine_application" "app" {
  project       = google_project.project.project_id
  location_id   = var.gae_location
  database_type = "CLOUD_FIRESTORE"
}

resource "google_service_account" "moneysaver" {
  account_id   = "moneysaver"
  display_name = "moneysaver"
  description  = "MoneySaver"
  project      = google_project.project.project_id
}

resource "google_project_iam_member" "moneysaver-datastore_user" {
  project = google_project.project.project_id
  role    = "roles/datastore.user"
  member  = "serviceAccount:${google_service_account.moneysaver.email}"
}

resource "google_cloud_run_service" "moneysaver" {
  name     = "moneysaver"
  location = var.location
  project  = google_project.project.project_id

  template {
    metadata {
      annotations = {
        "autoscaling.knative.dev/maxScale" = "1"
      }
    }
    spec {
      containers {
        image = "${google_artifact_registry_repository.containers.location}-docker.pkg.dev/${google_project.project.project_id}/${google_artifact_registry_repository.containers.repository_id}/moneysaver:latest"
        env {
          name  = "PROJECT_ID"
          value = google_project.project.project_id
        }
        env {
          name  = "SLACK_BOT_TOKEN"
          value = var.slack_bot_token
        }
        env {
          name  = "SLACK_VERIFICATION_TOKEN"
          value = var.slack_verification_token
        }
        env {
          name  = "LIMITS"
          value = var.limits
        }
        ports {
          container_port = 8080
        }
        resources {
          limits = {
            cpu    = "1"
            memory = "100Mi"
          }
        }
      }
      container_concurrency = 80
      timeout_seconds       = 10
      service_account_name  = google_service_account.moneysaver.email
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }
}

resource "google_cloud_run_service_iam_member" "moneysaver-allUsers" {
  location = google_cloud_run_service.moneysaver.location
  project  = google_project.project.project_id
  service  = google_cloud_run_service.moneysaver.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}
