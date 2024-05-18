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

# Configuración de API Gateway
module "api_gateway" {
  source = "./modules/api_gateway"
  lambda_execution_event_role_arn_invoke = module.event_processor_lambda.event_processor_lambda_invoke_arn
  lambda_execution_message_role_arn_invoke = module.message_processor_lambda.message_processor_lambda_arn
}

# Configuración de Lambda - Event Processor
module "event_processor_lambda" {
  source = "./modules/lambda/event_processor"
  lambda_execution_role_arn = module.iam.lambda_execution_role_arn
}

# Configuración de Lambda - Message Processor
module "message_processor_lambda" {
  source = "./modules/lambda/message_processor"
  lambda_execution_role_arn = module.iam.lambda_execution_role_arn
}


# Configuración de SQS
module "sqs" {
  source = "./modules/sqs"
}

# Configuración de SNS
module "sns" {
  source = "./modules/sns"
}

module "iam" {
  source = "./modules/iam"
}

resource "aws_lambda_permission" "event_permission" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = module.event_processor_lambda.event_processor_lambda_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${module.api_gateway.api_execution_arn}/*/POST/event"
}

resource "aws_lambda_permission" "message_permission" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = module.message_processor_lambda.message_processor_lambda_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${module.api_gateway.api_execution_arn}/*/POST/message"
}