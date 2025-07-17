provider "aws" {
  region = "us-west-2"
}

module "k3s_cluster" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 18.0"

  cluster_name = "sme-platform"
  vpc_id       = module.vpc.vpc_id
  subnets      = module.vpc.private_subnets

  node_groups = {
    primary = {
      desired_capacity = 3
      max_capacity     = 10
      min_capacity     = 1
      instance_types   = ["t3.medium"]
    }
  }
}

resource "aws_db_instance" "platform_db" {
  allocated_storage    = 20
  engine               = "postgres"
  instance_class       = "db.t3.micro"
  username             = "admin"
  password             = var.db_password
  publicly_accessible  = false
  skip_final_snapshot  = true
}