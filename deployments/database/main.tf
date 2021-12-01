variable "db_instance_name" {
  type    = string
  default = "getherscan-db-instance"
}

variable "db_version" {
  type    = string
  default = "POSTGRES_13"
}

variable "db_region" {
  type = string
  default = "us-west2"
}

variable "db_tier" {
  type    = string
  default = "db-custom-1-3840"
}

variable "disk_size" {
  type    = number
  default = 100
}

variable "db_name" {
  type    = string
  default = "getherscan_database"
}

variable "user_name" {
  type = string
  default = "postgres"
}

variable "user_password" {
  type = string
  default = "12345"
}

data "google_compute_network" "default" {
  name = "default"
}

resource "google_sql_database_instance" "instance" {
  name             = var.db_instance_name
  database_version = var.db_version
  region = var.db_region
  settings {
    tier      = var.db_tier
    disk_size = var.disk_size
    ip_configuration {
      ipv4_enabled    = true
    }
  }
}

resource "google_sql_database" "database" {
  name     = var.db_name
  instance = google_sql_database_instance.instance.name
}

resource "google_sql_user" "user" {
  name     = var.user_name
  instance = google_sql_database_instance.instance.name
  password = var.user_password
}

output "db_connection_string" {
  value = "host=${google_sql_database_instance.instance.public_ip_address} port=5432 user=${google_sql_user.user.name} password=${google_sql_user.user.password} sslmode=disable"
}
