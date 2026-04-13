variable "project_name" {
  description = "Project name used for resource naming."
  type        = string
}

variable "cluster_name" {
  description = "EKS cluster name used for subnet discovery tags."
  type        = string
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC."
  type        = string
}

variable "availability_zones" {
  description = "Availability zones used for the subnets."
  type        = list(string)
}

variable "public_subnet_cidrs" {
  description = "CIDR blocks for public subnets."
  type        = list(string)
}

variable "private_subnet_cidrs" {
  description = "CIDR blocks for private subnets."
  type        = list(string)
}

variable "common_tags" {
  description = "Common tags applied to all resources."
  type        = map(string)
  default     = {}
}
