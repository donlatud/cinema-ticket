# Concurrent seat lock test (Phase 11) — PowerShell
# Usage: .\scripts\concurrency_test.ps1 -ShowtimeId <id> -Token <jwt> [-SeatNo A5] [-Requests 10]

param(
    [Parameter(Mandatory = $true)]
    [string]$ShowtimeId,

    [Parameter(Mandatory = $true)]
    [string]$Token,

    [string]$SeatNo = "A5",
    [int]$Requests = 10,
    [string]$ApiUrl = "http://localhost:8080"
)

$body = @{ seat_nos = @($SeatNo) } | ConvertTo-Json
$headers = @{
    Authorization  = "Bearer $Token"
    "Content-Type" = "application/json"
}

$jobs = 1..$Requests | ForEach-Object {
    $i = $_
    Start-Job -ScriptBlock {
        param($Url, $Headers, $Body, $Index)
        try {
            $response = Invoke-WebRequest -Method POST -Uri $Url -Headers $Headers -Body $Body -UseBasicParsing
            [PSCustomObject]@{ Request = $Index; Status = $response.StatusCode }
        } catch {
            $status = $_.Exception.Response.StatusCode.value__
            [PSCustomObject]@{ Request = $Index; Status = $status }
        }
    } -ArgumentList "$ApiUrl/api/showtimes/$using:ShowtimeId/seats/lock", $headers, $body, $i
}

$results = $jobs | Wait-Job | Receive-Job
$jobs | Remove-Job

$results | Sort-Object Request | Format-Table -AutoSize

$created = ($results | Where-Object { $_.Status -eq 201 }).Count
$conflicts = ($results | Where-Object { $_.Status -eq 409 }).Count

Write-Host ""
Write-Host "201 Created: $created (expected 1)"
Write-Host "409 Conflict: $conflicts (expected $($Requests - 1))"

if ($created -ne 1) {
    exit 1
}
