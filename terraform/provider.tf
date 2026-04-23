terraform {
  required_providers {
    proxmox = {
      source = "bpg/proxmox"
    }
  }
}

provider "proxmox" {
  endpoint = "https://proxmox-host1:8006/"
  username = "root@pam"
  password = "your-password"
  insecure = true
}