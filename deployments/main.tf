terraform {
  backend "gcs" {
    bucket = "tf-state-getherscan-staging"
    prefix = "getherscan-staging"
  }
}

provider "google" {
  project = "getherscan-staging"
  region = "us-west-2"
}

provider "google-beta" {
  project = "getherscan-staging"
  region = "us-west-2"
}

# resource "google_container_registry" "gcr" {
#   project = "getherscan-staging"
#   location = "US"
# }

module "database" {
  source = "./database"
}

resource "google_compute_project_metadata" "ssh-key" {
  metadata = {
    ssh-keys = "akirillo:ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDLLIaJiz+fzFT8rzVqGv/yLhYnNu1B7LVEk7EDSjwbsOz5HlH0bb6xB0TY/jsNK+NyUEaKe1Ip9M7WIQC0quReKDA1h0UN7V0Jsda109nAWkKy/vprC3LOyI/SnBiBkGdnUctGCy7s7TYhe9FY3IP2mTEG0Vat+NgyW0237z/sXzX6rzwuaLF1Nu7JA88Ulfr+h2rKy8vQumLO5TaeIo0auQl+rVLoHYJCMjmy31bQh5i8S0o2V8AgTydSbd3K4jrhBR3He028Lixz595lCMSg1HOlSRWaB3rNZzMKm7CEBommuUB3UmISJZqGYGA1b7yMOYaYmZxyMk11DLAbZQr9 akirillo"
  }
}

module "poller" {
  source = "./poller"
}

module "api_server" {
  source = "./api_server"
}

module "ansible" {
  source = "./ansible"
  inventories = [
    module.poller.ansible_inventory,
    module.api_server.ansible_inventory
  ]
  db_connection_params = module.database.db_connection_params
}
