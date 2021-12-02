variable "inventories" {
  type    = list(string)
  default = []
}

variable "db_connection_params" {
  type = map(string)
  default = {}
}

resource "local_file" "ansible_inventory" {
  content = templatefile("${path.module}/inventory.tpl", {
    inventories   = var.inventories
    db_host = var.db_connection_params.host
    db_port = var.db_connection_params.port
    db_user = var.db_connection_params.user
    db_password = var.db_connection_params.password
    db_name = var.db_connection_params.dbname
  })
  file_permission = "0644"
  filename        = "${path.module}/hosts"
}
