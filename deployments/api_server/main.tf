variable "zone" {
  type = string
  default = "us-west2"
}

variable "name" {
  type = string
  default = "api_server"
}

variable "instance-type" {
  type = string
  default = "g1-small"
}

variable "disk-size" {
  type = number
  default = 100
}

resource "google_compute_instance" "api_server" {
  name                      = var.name
  description               = "Getherscan API Server"
  allow_stopping_for_update = false
  deletion_protection       = false
  tags = [var.name]
  lifecycle { ignore_changes = [metadata, boot_disk] }
  machine_type = var.instance-type
  zone = var.zone
  network_interface {
    network = "default"
    // assign public ip
    access_config {}
  }
  boot_disk {
    initialize_params {
      size  = var.disk-size
      type  = "pd-ssd"                          # SSD
      image = "ubuntu-os-cloud/ubuntu-2004-lts" # Latest 20.04 LTS
    }
  }
  shielded_instance_config {
    enable_secure_boot          = true
    enable_vtpm                 = true
    enable_integrity_monitoring = true
  }
  service_account {
    scopes = [
      "cloud-platform",   #
      "userinfo-email",   # Account info
      "logging-write",    # Write to stackdriver
      "monitoring-write", # Write to stackdriver monitoring,
      "storage-rw",       # read/write access to cloud storage
    ]
  }
}

output "ansible_inventory" {
  value = templatefile("${path.module}/inventory.tpl", {
    ip = google_compute_instance.coordinator.network_interface[0].network_ip
    name = var.name
  })
}
