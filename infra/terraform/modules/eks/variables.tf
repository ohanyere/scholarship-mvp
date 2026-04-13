variable "cluster_name" {
  description = "Name of the EKS cluster."
  type        = string
}

variable "cluster_version" {
  description = "Kubernetes version for the EKS cluster."
  type        = string
}

variable "private_subnet_ids" {
  description = "Private subnet IDs used by the cluster and node group."
  type        = list(string)
}

variable "endpoint_private_access" {
  description = "Whether the EKS endpoint is reachable privately."
  type        = bool
  default     = true
}

variable "endpoint_public_access" {
  description = "Whether the EKS endpoint is reachable publicly."
  type        = bool
  default     = true
}

variable "node_instance_type" {
  description = "Managed node group instance type."
  type        = string
  default     = "t3.medium"
}

variable "node_desired_size" {
  description = "Managed node group desired size."
  type        = number
  default     = 2
}

variable "node_min_size" {
  description = "Managed node group minimum size."
  type        = number
  default     = 1
}

variable "node_max_size" {
  description = "Managed node group maximum size."
  type        = number
  default     = 3
}

variable "common_tags" {
  description = "Common tags applied to all resources."
  type        = map(string)
  default     = {}
}
