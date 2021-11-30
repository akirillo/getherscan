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

module "poller" {
  source = "./poller"
}

module "api_server" {
  source = "./api_server"
}

module "database" {
  source = "./database"
}

module "ansible" {
  source = "./ansible"
  inventories = [
    module.poller.ansible_inventory,
    module.api_server.ansible_inventory
  ]
}
