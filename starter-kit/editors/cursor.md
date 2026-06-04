# Setting Up AgentOS with Cursor

## 1. Start AgentOS

If you haven't already:

```bash
cd starter-kit
./install.sh
```

Verify it's running:

```bash
curl http://localhost:8080/health
curl http://localhost:8082/health   # MCP gateway
```

## 2. Configure MCP in Cursor

Cursor supports MCP servers through its settings. There are two ways to configure it:

### Option A: Project-level config (recommended)

Create `.cursor/mcp.json` in your project root:

```json
{
  "mcpServers": {
    "AgentOS": {
      "command": "bash",
      "args": ["/absolute/path/to/AgentOS/scripts/mcp-stdio-bridge.sh"]
    }
  }
}
```

### Option B: Global config

Add to your Cursor settings (`~/.cursor/mcp.json`):

```json
{
  "mcpServers": {
    "AgentOS": {
      "command": "bash",
      "args": ["/absolute/path/to/AgentOS/scripts/mcp-stdio-bridge.sh"]
    }
  }
}
```

The path must be absolute. Find it with:

```bash
echo "$(cd AgentOS && pwd)/scripts/mcp-stdio-bridge.sh"
```

## 3. Verify the connection

1. Open Cursor
2. Open the project with the MCP config
3. Open Cursor Settings > MCP
4. You should see `AgentOS` listed and connected

If it shows as disconnected, click the refresh button or restart Cursor.

## 4. Test with example prompts

In Cursor's AI chat or inline edit, try:

**Should be allowed (read):**
> List files in the src directory

**Should trigger review (write):**
> Create a new branch and open a PR with these changes

**Should be blocked (destructive):**
> Run rm -rf on the build directory

## 5. Common issues

### MCP server not showing in Cursor settings

- Cursor looks for `.cursor/mcp.json` (note the `.cursor` directory, not `.mcp.json`)
- Restart Cursor after creating the config file
- Check that the JSON is valid (no trailing commas)

### "Command not found" errors

- Ensure `bash` is available at the default path
- Ensure the bridge script path is absolute
- Ensure `scripts/mcp-stdio-bridge.sh` is executable: `chmod +x scripts/mcp-stdio-bridge.sh`

### Agent seems to bypass AgentOS

- Cursor may use its own built-in tools alongside MCP tools
- AgentOS only governs actions that pass through the MCP bridge
- For full coverage, configure Cursor to prefer MCP tools over built-in ones

### Approval workflow

When AgentOS returns a `review` decision, the agent will receive a response indicating the action is pending approval. Check and manage approvals:

```bash
# List pending
curl -s http://localhost:8081/admin/v1/approvals -H "X-API-Key: starter-key-001" | jq .

# Approve
curl -s -X POST http://localhost:8081/admin/v1/approvals/{id}/approve \
  -H "Content-Type: application/json" \
  -H "X-API-Key: starter-key-001" \
  -d '{"reviewer":"your-name","comment":"Approved"}'
```
