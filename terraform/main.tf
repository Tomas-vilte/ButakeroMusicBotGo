module "vpc" {
  source = "./modules/vpc"

  vpc_cidr_block = var.vpc_cidr_block
}

module "subnet" {
  source = "./modules/subnet"

  vpc_id           = module.vpc.vpc_id
  subnet_cidr_block = var.subnet_cidr_block
  internet_gateway_id = module.vpc.internet_gateway_id
}

module "security_group" {
  source = "./modules/security-group"

  vpc_id = module.vpc.vpc_id
}

module "ec2" {
  source = "./modules/ec2"

  ami_id           = var.ami_id
  instance_type    = var.instance_type
  key_name         = var.key_name
  subnet_id        = module.subnet.subnet_id
  security_group_id = module.security_group.security_group_id
}