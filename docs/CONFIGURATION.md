# Configuration Guide

AgentOS is configured via a single YAML file. See [configs/agentos.example.yaml](file:///c:/Users/BIT/AgentOS/AegisFlow/configs/agentos.example.yaml) for the full annotated reference.

## Minimal Configuration

This configuration starts the AgentOS gateway on port 8080 (admin on port 8081) with a mock provider and a single tenant.

```yaml
server:
  port: 8080
  admin_port: 8081

providers:
  - name: "mock"
    type: "mock"
    enabled: true
    default: true

tenants:
  - id: "default"
    api_keys: ["agentos-test-default-001"]
    rate_limit:
      requests_per_minute: 60
      tokens_per_minute: 100000

routes:
  - match:
      model: "*"
    providers: ["mock"]
    strategy: "priority"
```

## Policy Configuration

Add input and output filters to your policy configurations. In this example, we block jailbreak keywords and detect PII before it reaches providers.

```yaml
policies:
  input:
    - name: "block-jailbreak"
      type: "keyword"
      action: "block"
      keywords:
        - "ignore previous instructions"
        - "ignore all instructions"
        - "DAN mode"
    - name: "pii-detection"
      type: "pii"
      action: "warn"
      patterns: ["ssn", "email", "credit_card"]
  output:
    - name: "content-filter"
      type: "keyword"
      action: "log"
      keywords: ["harmful-keyword"]
```

## Multi-Provider Configuration with Fallback

This setup configures OpenAI and Anthropic models with a fallback route to the mock provider in case the principal providers fail or are rate-limited.

```yaml
providers:
  - name: "openai"
    type: "openai"
    enabled: true
    base_url: "https://api.openai.com/v1"
    api_key_env: "OPENAI_API_KEY"
    models: ["openai-chat", "openai-fast"]

  - name: "anthropic"
    type: "anthropic"
    enabled: true
    base_url: "https://api.anthropic.com/v1"
    api_key_env: "ANTHROPIC_API_KEY"
    models: ["claude-sonnet-4-20250514"]

routes:
  - match:
      model: "openai-*"
    providers: ["openai", "mock"]
    strategy: "priority"

  - match:
      model: "claude-*"
    providers: ["anthropic", "mock"]
    strategy: "priority"
```
