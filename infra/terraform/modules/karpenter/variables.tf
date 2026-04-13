variable "cluster_name" {
  description = "EKS cluster name used for discovery and naming."
  type        = string
}

variable "cluster_oidc_provider_arn" {
  description = "OIDC provider ARN for the EKS cluster."
  type        = string
}

variable "cluster_oidc_issuer_url" {
  description = "OIDC issuer URL for the EKS cluster."
  type        = string
}

variable "service_account_namespace" {
  description = "Namespace for the future Karpenter service account."
  type        = string
  default     = "karpenter"
}

variable "service_account_name" {
  description = "Service account name for the future Karpenter controller."
  type        = string
  default     = "karpenter"
}

variable "common_tags" {
  description = "Common tags applied to all resources."
  type        = map(string)
  default     = {}
}
