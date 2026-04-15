terraform {
  backend "s3" {
    bucket       = "scholarship-mvp-tfstate-859217211731-use1"
    key          = "dev/terraform.tfstate"
    region       = "us-east-1"
    encrypt      = true
    use_lockfile = true
  }
}
