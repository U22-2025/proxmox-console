variable "proxmox_endpoint" {
  type        = string
  description = "Proxmox API endpoint"
}

variable "proxmox_api_token" {
  type      = string
  sensitive = true
}

variable "node_name" {}

variable "servername" {}
variable "cpu" {}
variable "memory" {}
variable "hdd" {}
variable "username" { sensitive = true }
variable "password_hash" { sensitive = true }