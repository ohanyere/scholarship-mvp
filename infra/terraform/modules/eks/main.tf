resource "aws_iam_role" "cluster" {
  name = "${var.cluster_name}-cluster-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "eks.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })

  tags = var.common_tags
}

resource "aws_iam_role_policy_attachment" "cluster_policy" {
  role       = aws_iam_role.cluster.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
}

resource "aws_iam_role_policy_attachment" "cluster_vpc_policy" {
  role       = aws_iam_role.cluster.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSVPCResourceController"
}

# checkov:skip=CKV_AWS_39:Dev cluster keeps the public API endpoint enabled so local kubectl access works without VPN or bastion setup.
# checkov:skip=CKV_AWS_38:Public API access is temporarily open during this learning environment phase until a fixed office/VPN CIDR is available.
# checkov:skip=CKV_AWS_58:Enabling envelope encryption after cluster creation is intentionally deferred because it can require disruptive cluster replacement for this existing dev environment.
# checkov:skip=CKV_AWS_339:Cluster version 1.30 remains AWS EKS-supported through extended support; a multi-version upgrade will be handled in a scheduled maintenance window instead of forcing it here.
resource "aws_eks_cluster" "this" {
  name                      = var.cluster_name
  role_arn                  = aws_iam_role.cluster.arn
  version                   = var.cluster_version
  enabled_cluster_log_types = var.cluster_enabled_log_types

  vpc_config {
    subnet_ids              = var.private_subnet_ids
    endpoint_private_access = var.endpoint_private_access
    endpoint_public_access  = var.endpoint_public_access
    public_access_cidrs     = var.public_access_cidrs
  }

  tags = merge(
    var.common_tags,
    {
      Name = var.cluster_name
    }
  )

  depends_on = [
    aws_iam_role_policy_attachment.cluster_policy,
    aws_iam_role_policy_attachment.cluster_vpc_policy
  ]
}

# Fetch the TLS certificate for the cluster OIDC issuer so Terraform can
# create the IAM OIDC provider used for IRSA.
data "tls_certificate" "oidc" {
  url = aws_eks_cluster.this.identity[0].oidc[0].issuer
}

# IAM OIDC provider for Kubernetes service accounts to assume IAM roles.
# This is required for IRSA and later for Karpenter controller access.
resource "aws_iam_openid_connect_provider" "this" {
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = [data.tls_certificate.oidc.certificates[0].sha1_fingerprint]
  url             = aws_eks_cluster.this.identity[0].oidc[0].issuer

  tags = var.common_tags
}

resource "aws_iam_role" "node_group" {
  name = "${var.cluster_name}-node-group-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })

  tags = var.common_tags
}

resource "aws_iam_role_policy_attachment" "node_group_worker_policy" {
  role       = aws_iam_role.node_group.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
}

resource "aws_iam_role_policy_attachment" "node_group_cni_policy" {
  role       = aws_iam_role.node_group.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
}

resource "aws_iam_role_policy_attachment" "node_group_ecr_policy" {
  role       = aws_iam_role.node_group.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
}

# Optional but strongly useful for debugging and node access through SSM.
resource "aws_iam_role_policy_attachment" "node_group_ssm_policy" {
  role       = aws_iam_role.node_group.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

# Small baseline managed node group for core cluster services.
# Karpenter can later add more nodes dynamically for workloads.
resource "aws_eks_node_group" "baseline" {
  cluster_name    = aws_eks_cluster.this.name
  node_group_name = "${var.cluster_name}-baseline"
  node_role_arn   = aws_iam_role.node_group.arn
  subnet_ids      = var.private_subnet_ids
  instance_types  = [var.node_instance_type]
  capacity_type   = "ON_DEMAND"

  scaling_config {
    desired_size = var.node_desired_size
    min_size     = var.node_min_size
    max_size     = var.node_max_size
  }

  labels = {
    purpose = "core-services"
  }

  tags = merge(
    var.common_tags,
    {
      Name                     = "${var.cluster_name}-baseline"
      "karpenter.sh/discovery" = var.cluster_name
    }
  )

  depends_on = [
    aws_iam_role_policy_attachment.node_group_worker_policy,
    aws_iam_role_policy_attachment.node_group_cni_policy,
    aws_iam_role_policy_attachment.node_group_ecr_policy,
    aws_iam_role_policy_attachment.node_group_ssm_policy
  ]
}
