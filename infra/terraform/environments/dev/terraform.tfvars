project_name    = "scholarship-platform"
cluster_name    = "scholarship-eks"
cluster_version = "1.30"
aws_region      = "us-east-1"

vpc_cidr = "10.0.0.0/16"

availability_zones = [
  "us-east-1a",
  "us-east-1b"
]

public_subnet_cidrs = [
  "10.0.1.0/24",
  "10.0.2.0/24"
]

private_subnet_cidrs = [
  "10.0.11.0/24",
  "10.0.12.0/24"
]

common_tags = {
  ManagedBy = "terraform"
  Project   = "scholarship-platform"
  Owner     = "platform-team"
}

node_instance_type = "t3.small"
node_desired_size  = 2
node_min_size      = 1
node_max_size      = 3
