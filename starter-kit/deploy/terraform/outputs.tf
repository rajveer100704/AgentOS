output "ecs_cluster_name" {
  description = "ECS cluster name"
  value       = aws_ecs_cluster.AgentOS.name
}

output "ecs_service_name" {
  description = "ECS service name"
  value       = aws_ecs_service.AgentOS.name
}

output "ecr_repository_url" {
  description = "ECR repository URL for pushing AgentOS images"
  value       = aws_ecr_repository.AgentOS.repository_url
}

output "security_group_id" {
  description = "Security group ID for the AgentOS service"
  value       = aws_security_group.AgentOS.id
}

output "task_definition_arn" {
  description = "ARN of the ECS task definition"
  value       = aws_ecs_task_definition.AgentOS.arn
}

output "log_group_name" {
  description = "CloudWatch log group name"
  value       = aws_cloudwatch_log_group.AgentOS.name
}
