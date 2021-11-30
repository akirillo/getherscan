variable "inventories" {
  type    = list(string)
  default = []
}

resource "local_file" "ansible_inventory" {
  content = templatefile("${path.module}/inventory.tpl", {
    inventories   = var.inventories
  })
  file_permission = "0644"
  filename        = "${path.module}/hosts"
}
