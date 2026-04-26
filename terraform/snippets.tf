resource "proxmox_virtual_environment_file" "cloudcfg" {
  content_type = "snippets"
  datastore_id = "local"
  node_name    = var.node_name

  source_raw {
    data = templatefile("${path.module}/cloud-config.yaml", {
      username      = var.username
      password_hash = var.password_hash
    })
  }
}