variable "project_name" {
  description = "Project name used for naming resources."
  type        = string
}

variable "cluster_name" {
  description = "EKS cluster name."
  type        = string
}

variable "cluster_version" {
  description = "EKS Kubernetes version."
  type        = string
}

variable "aws_region" {
  description = "AWS region for the environment."
  type        = string
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC."
  type        = string
}

variable "availability_zones" {
  description = "Availability zones used by the VPC."
  type        = list(string)
}

variable "public_subnet_cidrs" {
  description = "CIDR blocks for the public subnets."
  type        = list(string)
}

variable "private_subnet_cidrs" {
  description = "CIDR blocks for the private subnets."
  type        = list(string)
}

variable "common_tags" {
  description = "Common tags applied to all resources."
  type        = map(string)
}

variable "node_instance_type" {
  description = "Instance type for the baseline managed node group."
  type        = string
}

variable "node_desired_size" {
  description = "Desired number of nodes in the baseline managed node group."
  type        = number
}

variable "node_min_size" {
  description = "Minimum number of nodes in the baseline managed node group."
  type        = number
}

variable "node_max_size" {
  description = "Maximum number of nodes in the baseline managed node group."
  type        = number
}
