$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
if(Test-Path (Join-Path $root 'cmd/migrate')){throw 'Indexer must not own a migration command.'}
$expected=@('github.com/euler-gaozl/agent-os-knowledge/indexing','github.com/euler-gaozl/agent-os-knowledge/internal/files','github.com/euler-gaozl/agent-os-knowledge/internal/lexical','github.com/euler-gaozl/agent-os-knowledge/internal/provider')
$deps=@(& go -C $root list -mod=vendor -deps ./cmd/indexer)
if($LASTEXITCODE){throw 'Dependency inspection failed.'}
$owner=@($deps | Where-Object { $_.StartsWith('github.com/euler-gaozl/agent-os-knowledge/') })
if(@(Compare-Object ($expected|Sort-Object) ($owner|Sort-Object)).Count){throw 'Indexer dependency closure includes an unreviewed owner package.'}
if(Test-Path (Join-Path $root 'vendor/github.com/euler-gaozl/agent-os-knowledge/knowledge')){throw 'Knowledge HTTP/migration package must not be vendored into Indexer.'}
Write-Output 'Verified narrow indexing dependency closure and external migration ownership.'
