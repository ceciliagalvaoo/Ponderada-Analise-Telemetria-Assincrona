param(
  [string[]]$CpuProfiles = @('0.25', '0.5', '1.0'),
  [int]$Runs = 3,
  [string]$OutputRoot = 'test-load/reports/benchmarks',
  [string]$Duration = '45s',
  [int]$Vus = 30
)

Add-Type -AssemblyName System.Globalization

$ErrorActionPreference = 'Stop'
$PSNativeCommandUseErrorActionPreference = $false

function Parse-CpuProfile {
  param([string]$CpuValue)

  $normalized = $CpuValue.Replace(',', '.')
  return [double]::Parse($normalized, [System.Globalization.CultureInfo]::InvariantCulture)
}

function Wait-ApiReady {
  param(
    [string]$Url,
    [int]$MaxAttempts = 120,
    [int]$DelaySeconds = 2
  )

  for ($attempt = 1; $attempt -le $MaxAttempts; $attempt++) {
    try {
      $response = Invoke-WebRequest -Uri $Url -Method Post -UseBasicParsing -TimeoutSec 3 -Headers @{ 'Content-Type' = 'application/json' } -Body '{"device_id":"warmup","timestamp":"2026-03-22T00:00:00Z","sensor_type":"temperature","reading_type":"analog","value":1}'
      if ($response.StatusCode -eq 202) {
        return
      }
    } catch {
      # Keep retrying while containers warm up.
    }

    Start-Sleep -Seconds $DelaySeconds
  }

  throw "API nao ficou pronta no tempo esperado: $Url"
}

function Get-Median {
  param([double[]]$Values)

  if (-not $Values -or $Values.Count -eq 0) {
    return 0
  }

  $sorted = $Values | Sort-Object
  $count = $sorted.Count
  if ($count % 2 -eq 1) {
    return [double]$sorted[[int]($count / 2)]
  }

  $left = [double]$sorted[($count / 2) - 1]
  $right = [double]$sorted[($count / 2)]
  return ($left + $right) / 2
}

function Read-SummaryMetrics {
  param([string]$SummaryPath)

  $json = Get-Content -Raw -Path $SummaryPath | ConvertFrom-Json
  return [PSCustomObject]@{
    req_s     = [double]$json.metrics.http_reqs.rate
    p95_ms    = [double]$json.metrics.http_req_duration.'p(95)'
    fail_rate = [double]$json.metrics.http_req_failed.value
    checks_ok = [int]$json.metrics.checks.passes
    checks_ko = [int]$json.metrics.checks.fails
  }
}

$timestamp = Get-Date -Format 'yyyyMMdd-HHmmss'
$runRoot = Join-Path $OutputRoot $timestamp
New-Item -ItemType Directory -Force -Path $runRoot | Out-Null

$allRuns = @()

foreach ($cpu in $CpuProfiles) {
  $cpuNumeric = Parse-CpuProfile -CpuValue $cpu

  Write-Host "\n=== Perfil CPU $cpu ==="
  $env:BENCH_CPUS = $cpu
  $env:TEST_PROFILE = "cpu-$cpu"
  $env:K6_DURATION = $Duration
  $env:K6_VUS = "$Vus"

  docker compose -f docker-compose.yml -f docker-compose.bench.yml down | Out-Null
  docker compose -f docker-compose.yml -f docker-compose.bench.yml up --build -d | Out-Null

  Write-Host 'Aguardando API ficar pronta...'
  Wait-ApiReady -Url 'http://localhost:8080/telemetry'

  $profileDirName = "cpu-$($cpu.Replace('.', '_'))"
  $profileDir = Join-Path $runRoot $profileDirName
  New-Item -ItemType Directory -Force -Path $profileDir | Out-Null

  for ($run = 1; $run -le $Runs; $run++) {
    $summaryPath = Join-Path $profileDir ("run-$run-summary.json")
    $logPath = Join-Path $profileDir ("run-$run.log")

    Write-Host "Rodada $run/$Runs - cpu=$cpu"
    k6 run --summary-export $summaryPath test-load/telemetry.js *>&1 | Tee-Object -FilePath $logPath | Out-Null

    if ($LASTEXITCODE -ne 0) {
      throw "k6 falhou na rodada $run para cpu=$cpu"
    }

    $m = Read-SummaryMetrics -SummaryPath $summaryPath
    $allRuns += [PSCustomObject]@{
      cpu       = $cpuNumeric
      run       = $run
      req_s     = [double]$m.req_s
      p95_ms    = [double]$m.p95_ms
      fail_rate = [double]$m.fail_rate
      checks_ok = [int]$m.checks_ok
      checks_ko = [int]$m.checks_ko
      summary   = $summaryPath
      log       = $logPath
    }
  }
}

docker compose -f docker-compose.yml -f docker-compose.bench.yml down | Out-Null

$groups = $allRuns | Group-Object cpu | Sort-Object { [double]$_.Group[0].cpu }
$aggregated = @()

foreach ($g in $groups) {
  $cpu = [double]$g.Group[0].cpu
  $reqList = @($g.Group | ForEach-Object { [double]$_.req_s })
  $p95List = @($g.Group | ForEach-Object { [double]$_.p95_ms })
  $failList = @($g.Group | ForEach-Object { [double]$_.fail_rate })

  $reqAvg = [double](($reqList | Measure-Object -Average).Average)
  $reqMedian = [double](Get-Median -Values $reqList)
  $p95Avg = [double](($p95List | Measure-Object -Average).Average)
  $failAvg = [double](($failList | Measure-Object -Average).Average)

  $aggregated += [PSCustomObject]@{
    cpu                = $cpu
    runs               = [int]$g.Count
    req_s_avg          = [math]::Round($reqAvg, 4)
    req_s_median       = [math]::Round($reqMedian, 4)
    p95_ms_avg         = [math]::Round($p95Avg, 4)
    fail_rate_avg      = [math]::Round($failAvg, 6)
    estimated_1core    = if ($cpu -lt 1) { [math]::Round($reqAvg / $cpu, 4) } else { $null }
    extrapolation_note = if ($cpu -lt 1) { 'regra_de_3' } else { 'medicao_real' }
  }
}

$realOneCore = ($aggregated | Where-Object { $_.cpu -eq 1.0 } | Select-Object -First 1)
foreach ($row in $aggregated) {
  if ($row.cpu -lt 1 -and $realOneCore) {
    $row | Add-Member -NotePropertyName real_1core_req_s -NotePropertyValue $realOneCore.req_s_avg
    $err = (($row.estimated_1core - $realOneCore.req_s_avg) / $realOneCore.req_s_avg) * 100
    $row | Add-Member -NotePropertyName extrapolation_error_pct -NotePropertyValue ([math]::Round($err, 2))
  } else {
    $row | Add-Member -NotePropertyName real_1core_req_s -NotePropertyValue $null
    $row | Add-Member -NotePropertyName extrapolation_error_pct -NotePropertyValue $null
  }
}

$rawJsonPath = Join-Path $runRoot 'raw-runs.json'
$summaryJsonPath = Join-Path $runRoot 'benchmark-summary.json'
$summaryCsvPath = Join-Path $runRoot 'benchmark-summary.csv'
$summaryMdPath = Join-Path $runRoot 'benchmark-summary.md'
$latestPath = Join-Path $OutputRoot 'latest.txt'

$allRuns | ConvertTo-Json -Depth 6 | Set-Content -Path $rawJsonPath
$aggregated | ConvertTo-Json -Depth 6 | Set-Content -Path $summaryJsonPath
$aggregated | Export-Csv -Path $summaryCsvPath -NoTypeInformation -Encoding UTF8

$md = @()
$md += '# Benchmark k6 com limite de CPU/RAM'
$md += ''
$md += "Data: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
$md += "Rodadas por perfil: $Runs"
$md += "Duracao por rodada: $Duration"
$md += "VUs por rodada: $Vus"
$md += ''
$md += '| CPU (core) | req/s medio | req/s mediana | p95 medio (ms) | fail_rate medio | estimado 1 core | 1 core real | erro extrapolacao (%) |'
$md += '| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |'
foreach ($row in $aggregated | Sort-Object cpu) {
  $md += "| $($row.cpu) | $($row.req_s_avg) | $($row.req_s_median) | $($row.p95_ms_avg) | $($row.fail_rate_avg) | $($row.estimated_1core) | $($row.real_1core_req_s) | $($row.extrapolation_error_pct) |"
}
$md += ''
$md += 'Formula usada: req_s_estimado_1core = req_s_medido / cpu_perfil'

$md -join [Environment]::NewLine | Set-Content -Path $summaryMdPath
Set-Content -Path $latestPath -Value $timestamp

Write-Host "\nBenchmark finalizado."
Write-Host "Diretorio: $runRoot"
Write-Host "Resumo Markdown: $summaryMdPath"
Write-Host "Resumo JSON: $summaryJsonPath"
Write-Host "Resumo CSV: $summaryCsvPath"
