variable "inventories" {
  type    = list(string)
  default = []
}

variable "db_connection_string" {
  type = string
}

resource "local_file" "ansible_inventory" {
  content = templatefile("${path.module}/inventory.tpl", {
    inventories   = var.inventories
    db_connection_string = var.db_connection_string
  })
  file_permission = "0644"
  filename        = "${path.module}/hosts"
}
