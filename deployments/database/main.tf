variable "db_name" {
  type    = string
  default = "getherscan_database"
}

variable "db_version" {
  type    = string
  default = "POSTGRES_13"
}

variable "db_tier" {
  type    = string
  default = "db-custom-1-3840"
}

variable "disk_size" {
  type    = number
  default = 100
}

data "google_compute_network" "default" {
  name = "default"
}

resource "google_sql_database_instance" "database" {
  name             = var.db_name
  database_version = var.db_version
  settings {
    tier      = var.db_tier
    disk_size = var.disk_size
    ip_configuration {
      ipv4_enabled    = true
    }
  }
}
