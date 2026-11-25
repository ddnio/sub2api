import { spawnSync } from 'node:child_process'
import { chmodSync, mkdtempSync, readdirSync, readFileSync, rmSync, statSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'
import { createCodexCliFiles } from '@/utils/codexConfig'
import {
  DEFAULT_CODEX_MODEL,
  buildCodexQuickSetupCommand,
  buildOpenAICodexConfigToml,
  ensureOpenAIV1BaseUrl
} from '@/utils/codexAuthConfig'

describe('createCodexCliFiles', () => {
  it('defaults to single-file api-key config matching UseKeyModal and /start', () => {
    const files = createCodexCliFiles({
      baseUrl: 'https://relay.example.com',
      apiKey: 'your-api-key-here',
      os: 'unix',
      providerName: 'Sub2API'
    })

    expect(files).toHaveLength(1)
    expect(files[0].path).toBe('~/.codex/config.toml')
    expect(files[0].content).toBe(
      buildOpenAICodexConfigToml({
        baseUrl: 'https://relay.example.com',
        apiKey: 'your-api-key-here'
      })
    )
    expect(files[0].content).toContain('model_provider = "OpenAI"')
    expect(files[0].content).toContain(`model = "${DEFAULT_CODEX_MODEL}"`)
    expect(files[0].content).toContain('base_url = "https://relay.example.com/v1"')
    expect(files[0].content).toContain('requires_openai_auth = false')
    expect(files[0].content).toContain('experimental_bearer_token = "your-api-key-here"')
    expect(files[0].content).toContain('http_headers = { "x-openai-actor-authorization" = "local-image-extension" }')
    expect(files[0].content).not.toContain('supports_websockets = true')

    // /start quick setup commands must embed the exact same config.toml
    const unixCmd = buildCodexQuickSetupCommand('https://relay.example.com', 'your-api-key-here', 'unix')
    const psCmd = buildCodexQuickSetupCommand('https://relay.example.com', 'your-api-key-here', 'powershell')
    expect(unixCmd).toContain(files[0].content)
    expect(psCmd).toContain(files[0].content)
  })

  it('uses Windows config path in default api-key mode', () => {
    const files = createCodexCliFiles({
      baseUrl: 'https://relay.example.com/v1',
      apiKey: 'sk-real-key',
      os: 'windows',
      providerName: 'Sub2API'
    })

    expect(files).toHaveLength(1)
    expect(files[0].path).toBe('%userprofile%\\.codex/config.toml')
    expect(files[0].content).toContain('base_url = "https://relay.example.com/v1"')
    expect(files[0].content).toContain('experimental_bearer_token = "sk-real-key"')
  })

  it('adds websocket flags only for advanced websocket mode', () => {
    const files = createCodexCliFiles({
      baseUrl: 'https://relay.example.com',
      apiKey: 'sk-websocket',
      os: 'unix',
      websocket: true,
      providerName: 'Sub2API'
    })

    expect(files[0].content).toContain('supports_websockets = true')
    expect(files[0].content).toContain('responses_websockets_v2 = true')
  })

  it('supports legacy auth.json mode when explicitly requested', () => {
    const files = createCodexCliFiles({
      baseUrl: 'https://relay.example.com',
      apiKey: 'sk-legacy-mode',
      os: 'unix',
      authMode: 'legacy',
      providerName: 'Sub2API'
    })

    expect(files).toHaveLength(2)
    expect(files[0].content).toContain('requires_openai_auth = true')
    expect(files[1].path).toBe('~/.codex/auth.json')
    expect(files[1].content).toContain('"OPENAI_API_KEY": "sk-legacy-mode"')
  })

  it('never turns the prerender site name placeholder into a literal provider key', () => {
    const files = createCodexCliFiles({
      baseUrl: 'https://api.prerender.invalid',
      apiKey: 'your-api-key-here',
      os: 'unix',
      providerName: '__PRERENDER_SITE_NAME__'
    })

    expect(files[0].content).toContain('model_provider = "OpenAI"')
    expect(files[0].content).not.toMatch(/prerender_site_name/i)
  })

  it('normalizes trailing slashes and /v1 without producing /v1/v1', () => {
    expect(ensureOpenAIV1BaseUrl('https://relay.example.com/v1//')).toBe('https://relay.example.com/v1')
    expect(ensureOpenAIV1BaseUrl('https://relay.example.com///')).toBe('https://relay.example.com/v1')
    const toml = buildOpenAICodexConfigToml({
      baseUrl: 'https://relay.example.com/v1//',
      apiKey: 'sk-test'
    })
    expect(toml).toContain('base_url = "https://relay.example.com/v1"')
    expect(toml).not.toContain('/v1/v1')
  })

  it('writes atomically via temp file, stops on error, creates unique backups, restricts permissions, and respects CODEX_HOME', () => {
    const unixCmd = buildCodexQuickSetupCommand('https://relay.example.com', 'sk-secret', 'unix')
    expect(unixCmd).toContain('set -e')
    expect(unixCmd).toContain('CODEX_HOME:-$HOME/.codex')
    expect(unixCmd).toContain('umask 077')
    expect(unixCmd).toContain('mktemp "$d/config.toml.tmp.XXXXXX"')
    expect(unixCmd).toContain('chmod 600 "$t"')
    expect(unixCmd).toContain('mktemp "$f.bak-$(date +%Y%m%d%H%M%S).XXXXXX"')
    expect(unixCmd).toContain('mv -f "$t" "$f"')
    expect(unixCmd).toContain('已备份原有配置到')

    const psCmd = buildCodexQuickSetupCommand('https://relay.example.com', 'sk-secret', 'powershell')
    expect(psCmd).toContain("$ErrorActionPreference = 'Stop'")
    expect(psCmd).toContain('$env:CODEX_HOME')
    expect(psCmd).toContain('[System.Security.Principal.WindowsIdentity]::GetCurrent().User.Value')
    expect(psCmd).toContain('icacls $t /inheritance:r /grant:r "*$($u):F"')
    expect(psCmd).toContain('if ($LASTEXITCODE -ne 0)')
    expect(psCmd).toContain('Set-Content -LiteralPath $t')
    expect(psCmd).toContain('Test-Path -LiteralPath $f')
    expect(psCmd).toContain('Copy-Item -LiteralPath $f -Destination $b')
    expect(psCmd).toContain('Move-Item -LiteralPath $t -Destination $f -Force')
    expect(psCmd).toContain('Remove-Item -LiteralPath $t')
    expect(psCmd).not.toMatch(/\s-Path\s/)
    expect(psCmd.indexOf('icacls $t')).toBeLessThan(psCmd.indexOf('Set-Content -LiteralPath $t'))
    expect(psCmd).toContain('.bak-$(Get-Date -Format yyyyMMddHHmmss)-')
  })

  it.skipIf(process.platform === 'win32')(
    'unix quick command preserves existing config on directory write failure, sets 0600, and keeps unique backups',
    () => {
      const dir = mkdtempSync(join(tmpdir(), 'codex-quick-test-'))
      const configPath = join(dir, 'config.toml')
      try {
        // 1. Existing 0644 config in a read-only directory: command must fail and leave original file untouched
        writeFileSync(configPath, 'original-content = true\n', { mode: 0o644 })
        chmodSync(configPath, 0o644)
        chmodSync(dir, 0o555)
        const cmd1 = buildCodexQuickSetupCommand('https://relay.example.com', 'sk-first-secret', 'unix')
        const failRes = spawnSync('sh', ['-c', cmd1], {
          env: { ...process.env, CODEX_HOME: dir },
          encoding: 'utf8'
        })
        expect(failRes.status).not.toBe(0)
        expect(readFileSync(configPath, 'utf8')).toBe('original-content = true\n')

        // 2. Restore dir write permission: atomic replace sets 0600, creates backup, leaves no tmp files
        chmodSync(dir, 0o755)
        const okRes1 = spawnSync('sh', ['-c', cmd1], {
          env: { ...process.env, CODEX_HOME: dir },
          encoding: 'utf8'
        })
        expect(okRes1.status).toBe(0)
        expect(readFileSync(configPath, 'utf8')).toContain('experimental_bearer_token = "sk-first-secret"')
        expect(statSync(configPath).mode & 0o777).toBe(0o600)

        // 3. Second execution in the same second creates a second unique backup without overwriting the first
        const cmd2 = buildCodexQuickSetupCommand('https://relay.example.com', 'sk-second-secret', 'unix')
        const okRes2 = spawnSync('sh', ['-c', cmd2], {
          env: { ...process.env, CODEX_HOME: dir },
          encoding: 'utf8'
        })
        expect(okRes2.status).toBe(0)
        expect(readFileSync(configPath, 'utf8')).toContain('experimental_bearer_token = "sk-second-secret"')
        expect(statSync(configPath).mode & 0o777).toBe(0o600)

        const files = readdirSync(dir)
        const backups = files.filter((f) => f.startsWith('config.toml.bak-'))
        const tmps = files.filter((f) => f.startsWith('config.toml.tmp.'))
        expect(backups).toHaveLength(2)
        expect(tmps).toHaveLength(0)
        const backupContents = backups.map((b) => readFileSync(join(dir, b), 'utf8'))
        expect(backupContents).toContain('original-content = true\n')
        expect(backupContents.some((c) => c.includes('experimental_bearer_token = "sk-first-secret"'))).toBe(true)
      } finally {
        chmodSync(dir, 0o755)
        rmSync(dir, { recursive: true, force: true })
      }
    }
  )
})
