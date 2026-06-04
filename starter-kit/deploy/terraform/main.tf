terraform {
  required_version = ">= 1.5"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

# --- VPC (use existing or create) ---

data "aws_vpc" "selected" {
  id = var.vpc_id
}

data "aws_subnets" "private" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.selected.id]
  }

  tags = {
    Tier = "private"
  }
}

# --- ECR Repository ---

resource "aws_ecr_repository" "AgentOS" {
  name                 = "AgentOS"
  image_tag_mutability = "MUTABLE"
  force_delete         = false

  image_scanning_configuration {
    scan_on_push = true
  }
}

# --- ECS Cluster ---

resource "aws_ecs_cluster" "AgentOS" {
  name = "${var.name_prefix}-cluster"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}

# --- CloudWatch Log Group ---

resource "aws_cloudwatch_log_group" "AgentOS" {
  name              = "/ecs/${var.name_prefix}"
  retention_in_days = 30
}

# --- IAM ---

resource "aws_iam_role" "ecs_task_execution" {
  name = "${var.name_prefix}-task-execution"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ecs-tasks.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "ecs_task_execution" {
  role       = aws_iam_role.ecs_task_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_iam_role" "ecs_task" {
  name = "${var.name_prefix}-task"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ecs-tasks.amazonaws.com"
      }
    }]
  })
}

# --- Security Group ---

resource "aws_security_group" "AgentOS" {
  name_prefix = "${var.name_prefix}-"
  vpc_id      = data.aws_vpc.selected.id

  ingress {
    from_port   = 8080
    to_port     = 8082
    protocol    = "tcp"
    cidr_blocks = [data.aws_vpc.selected.cidr_block]
    description = "AgentOS gateway, admin, and MCP ports"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound"
  }
}

# --- ECS Task Definition ---

resource "aws_ecs_task_definition" "AgentOS" {
  family                   = var.name_prefix
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.task_cpu
  memory                   = var.task_memory
  execution_role_arn       = aws_iam_role.ecs_task_execution.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  container_definitions = jsonencode([{
    name  = "AgentOS"
    image = "${aws_ecr_repository.AgentOS.repository_url}:${var.image_tag}"

    portMappings = [
      { containerPort = 8080, protocol = "tcp" },
      { containerPort = 8081, protocol = "tcp" },
      { containerPort = 8082, protocol = "tcp" },
    ]

    healthCheck = {
      command     = ["CMD-SHELL", "curl -f http://localhost:8080/health || exit 1"]
      interval    = 10
      timeout     = 5
      retries     = 3
      startPeriod = 10
    }

    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.AgentOS.name
        "awslogs-region"        = var.aws_region
        "awslogs-stream-prefix" = "AgentOS"
      }
    }

    environment = [
      { name = "AgentOS_GOVERNANCE_MODE", value = "governance" },
    ]
  }])
}

# --- ECS Service ---

resource "aws_ecs_service" "AgentOS" {
  name            = var.name_prefix
  cluster         = aws_ecs_cluster.AgentOS.id
  task_definition = aws_ecs_task_definition.AgentOS.arn
  desired_count   = var.desired_count
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = data.aws_subnets.private.ids
    security_groups  = [aws_security_group.AgentOS.id]
    assign_public_ip = false
  }

  deployment_circuit_breaker {
    enable   = true
    rollback = true
  }
}
