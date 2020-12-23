variable "billing_account_id" {}
variable "project_id" {}

variable "github_owner" {} // nownabe
variable "github_name" {}  // moneysaver

variable "slack_bot_token" {}
variable "slack_verification_token" {}
variable "limits" {}

variable "location" {
  default = "us-central1"
}
variable "gae_location" {
  default = "us-central"
}
