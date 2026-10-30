param(
    [string]$Go = 'go',
    [string]$Scratch = (Join-Path $PSScriptRoot '..\.local-check')
)
$ErrorActionPreference = 'Stop'
$repo = Split-Path $PSScriptRoot -Parent
$Scratch = [IO.Path]::GetFullPath($Scratch)
New-Item -ItemType Directory -Force $Scratch | Out-Null
$saved = @{}
foreach ($name in @('GOCACHE','GOMODCACHE','TEMP','TMP')) { $saved[$name] = [Environment]::GetEnvironmentVariable($name, 'Process') }
Push-Location $repo
try {
    $env:TEMP = $Scratch
    $env:TMP = $Scratch
    if (!$env:GOCACHE) { $env:GOCACHE = Join-Path $Scratch 'build-cache' }
    if (!$env:GOMODCACHE) { $env:GOMODCACHE = Join-Path $Scratch 'mod-cache' }
    # Upstream's module requires/replaces itself using moving branch names.
    # Omit only those directives in an external modfile; keep go.mod unchanged.
    $mod = [IO.File]::ReadAllText((Join-Path $repo 'go.mod'))
    $mod = $mod -replace '(?m)^require github\.com/lejianwen/rustdesk-api/v2 [^\r\n]*\r?\n', ''
    $mod = $mod -replace '(?m)^replace github\.com/lejianwen/rustdesk-api/v2 => [^\r\n]*\r?\n', ''
    $modPath = Join-Path $Scratch 'verify.mod'
    [IO.File]::WriteAllText($modPath, $mod, [Text.UTF8Encoding]::new($false))
    if (Test-Path 'go.sum') { Copy-Item 'go.sum' (Join-Path $Scratch 'verify.sum') }
    & $Go mod tidy "-modfile=$modPath"
    if ($LASTEXITCODE -ne 0) { throw 'Module preparation failed' }
    & $Go test "-modfile=$modPath" -timeout 180s ./http/controller/... ./http/request/... ./http/router ./service ./model/... -count=1
    if ($LASTEXITCODE -ne 0) { throw 'API/book tests failed' }
    & $Go build "-modfile=$modPath" -o (Join-Path $Scratch 'apimain.exe') ./cmd/apimain.go
    if ($LASTEXITCODE -ne 0) { throw 'Build failed' }
} finally {
    Pop-Location
    foreach ($name in $saved.Keys) { [Environment]::SetEnvironmentVariable($name, $saved[$name], 'Process') }
}
