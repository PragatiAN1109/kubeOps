
# Create a dedicated namespace managed by Terraform
resource "kubernetes_namespace_v1" "kubeops" {
  metadata {
    name = "kubeops-terraform"

    labels = {
      app        = "kubeops"
      managed-by = "terraform"
    }
  }
}

# Create application configuration inside the namespace
resource "kubernetes_config_map_v1" "app_config" {
  metadata {
    name      = "kubeops-config"
    namespace = kubernetes_namespace_v1.kubeops.metadata[0].name

    labels = {
      app        = "kubeops"
      managed-by = "terraform"
    }
  }

  data = {
    APP_NAME    = "kubeops-api"
    ENVIRONMENT = "development"
    LOG_LEVEL   = "info"
  }
}
