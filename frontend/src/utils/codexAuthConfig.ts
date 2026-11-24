import { buildCodexModelCatalogUrl } from '@/api/codex'

/**
 * Codex 0.164 checks a static actor header when registering its native image
 * extension for a custom API-key provider. The marker is not a credential;
 * requests authenticate with the provider's normal bearer token.
 */
export function buildCodexApiKeyAuthConfig(apiKey: string): string {
  return `requires_openai_auth = false
experimental_bearer_token = ${JSON.stringify(apiKey)}
http_headers = { "x-openai-actor-authorization" = "local-image-extension" }`
}

export const DEFAULT_CODEX_MODEL = 'gpt-6.1-sol'

export type CodexAuthMode = 'legacy' | 'api-key'
export type CodexQuickSetupShell = 'unix' | 'powershell'

export interface OpenAICodexConfigOptions {
  baseUrl: string
  apiKey: string
  authMode?: CodexAuthMode
  model?: string
  reasoningEffortLine?: string
  catalogMode?: 'remote' | 'file'
  localCatalogToml?: string
  websocket?: boolean
}

export function ensureOpenAIV1BaseUrl(baseUrl: string): string {
  const baseRoot = baseUrl.trim().replace(/\/+$/, '').replace(/\/v1$/, '').replace(/\/+$/, '')
  return `${baseRoot}/v1`
}

export function escapeTomlBasicString(value: string): string {
  return value.replace(/\\/g, '\\\\').replace(/"/g, '\\"')
}

/**
 * Shared OpenAI Codex config.toml generator used by UseKeyModal, StartView, and Docs
 * so the console modal, onboarding quick-setup command, and docs always emit identical config.
 */
export function buildOpenAICodexConfigToml(options: OpenAICodexConfigOptions): string {
  const apiBase = ensureOpenAIV1BaseUrl(options.baseUrl)
  const authMode: CodexAuthMode = options.authMode ?? 'api-key'
  const model = options.model ?? DEFAULT_CODEX_MODEL
  const reasoningEffortLine = options.reasoningEffortLine ?? ''
  const catalogMode = options.catalogMode ?? 'remote'
  const localCatalogToml = options.localCatalogToml ?? ''
  const websocketProviderLine = options.websocket ? '\nsupports_websockets = true' : ''
  const websocketFeatureLine = options.websocket ? 'responses_websockets_v2 = true\n' : ''
  const providerAuthConfig =
    authMode === 'api-key' ? buildCodexApiKeyAuthConfig(options.apiKey) : 'requires_openai_auth = true'

  return `model_provider = "OpenAI"
model = "${model}"
review_model = "${model}"
web_search = "live"
${reasoningEffortLine}disable_response_storage = true
${localCatalogToml}network_access = "enabled"
windows_wsl_setup_acknowledged = true

[model_providers.OpenAI]
name = "OpenAI"
base_url = "${apiBase}"
${catalogMode === 'remote' ? `model_catalog_url = "${escapeTomlBasicString(buildCodexModelCatalogUrl(apiBase))}"\n` : ''}wire_api = "responses"
${providerAuthConfig}${websocketProviderLine}

[features]
${catalogMode === 'remote' ? 'api_key_model_discovery = true\n' : ''}${authMode === 'api-key' ? 'image_generation = true\n' : ''}${websocketFeatureLine}goals = true`
}

export function buildCodexQuickSetupCommand(
  baseUrl: string,
  apiKey: string,
  shell: CodexQuickSetupShell
): string {
  const configToml = buildOpenAICodexConfigToml({ baseUrl, apiKey })
  if (shell === 'powershell') {
    return `& { $ErrorActionPreference = 'Stop'; $d = if ($env:CODEX_HOME) { $env:CODEX_HOME } else { Join-Path $HOME '.codex' }; $f = Join-Path $d 'config.toml'; [System.IO.Directory]::CreateDirectory($d) | Out-Null; $t = Join-Path $d ("config.toml.tmp-" + [guid]::NewGuid().ToString('N')); $u = [System.Security.Principal.WindowsIdentity]::GetCurrent().User.Value; try { [System.IO.File]::WriteAllBytes($t, [byte[]]@()); icacls $t /inheritance:r /grant:r "*$($u):F" | Out-Null; if ($LASTEXITCODE -ne 0) { throw "icacls failed ($LASTEXITCODE)" }; @'\n${configToml}\n'@ | Set-Content -LiteralPath $t -Encoding UTF8; if (Test-Path -LiteralPath $f) { $b = "$f.bak-$(Get-Date -Format yyyyMMddHHmmss)-$([guid]::NewGuid().ToString('N').Substring(0,6))"; Copy-Item -LiteralPath $f -Destination $b -Force; Write-Host "已备份原有配置到: $b" }; Move-Item -LiteralPath $t -Destination $f -Force; Write-Host "已写入配置: $f" } catch { if (Test-Path -LiteralPath $t) { Remove-Item -LiteralPath $t -Force -ErrorAction SilentlyContinue }; throw } }`
  }
  return `(set -e; umask 077; d="\${CODEX_HOME:-$HOME/.codex}"; f="$d/config.toml"; mkdir -p "$d"; t="$(mktemp "$d/config.toml.tmp.XXXXXX")"; trap 'rm -f "$t"' EXIT; cat <<'EOF' > "$t"\n${configToml}\nEOF\nchmod 600 "$t"; if [ -f "$f" ]; then b="$(mktemp "$f.bak-$(date +%Y%m%d%H%M%S).XXXXXX")"; cp -p "$f" "$b"; echo "已备份原有配置到: $b"; fi; mv -f "$t" "$f"; trap - EXIT; echo "已写入配置: $f")`
}
