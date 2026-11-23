
output "namespace_name" {
  description = "Name of the Terraform-managed namespace"
  value       = kubernetes_namespace_v1.kubeops.metadata[0].name
}

output "configmap_name" {
  description = "Name of the Terraform-managed ConfigMap"
  value       = kubernetes_config_map_v1.app_config.metadata[0].name
}
