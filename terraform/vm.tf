resource "proxmox_virtual_environment_vm" "vm" {
  name      = var.servername
  node_name = var.node_name

  clone {
    vm_id = 9001
  }

  cpu {
    cores = var.cpu
  }

  memory {
    dedicated = var.memory
  }

  disk {
    datastore_id = "local-lvm"
    interface    = "scsi0"
    size         = var.hdd
  }

  network_device {
    bridge  = "vmbr0"
    model   = "virtio"
    vlan_id = 20
  }

  initialization {
    ip_config {
      ipv4 {
        address = "172.32.0.50/24"
        gateway = "172.32.0.254"
      }
    }

    user_data_file_id = proxmox_virtual_environment_file.cloudcfg.id
  }

  agent {
    enabled = true
  }
}

output "vm_ip" {
  value = try(
    proxmox_virtual_environment_vm.vm.ipv4_addresses[0].address,
    "waiting"
  )
}