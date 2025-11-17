# test_completo_debug.ps1

Write-Host "Iniciando docker-compose..." -ForegroundColor Green
docker-compose up -d

Write-Host "Esperando que los servicios estén listos..." -ForegroundColor Yellow
Start-Sleep -Seconds 20

Write-Host "Debuggeando ingest CSV..." -ForegroundColor Magenta

$csvPath = "servicioBBDD/data/datos.csv"

Write-Host "Verificando archivo CSV..." -ForegroundColor Yellow
if (-not (Test-Path $csvPath)) {
    Write-Host "Archivo no encontrado: $csvPath" -ForegroundColor Red
    Write-Host "Buscando archivos CSV..." -ForegroundColor Yellow
    Get-ChildItem -Recurse -Filter "*.csv" | ForEach-Object { 
        Write-Host "Encontrado: $($_.FullName)" 
        Write-Host "Tamano: $($_.Length) bytes" 
    }
    exit 1
} else {
    $fileInfo = Get-Item $csvPath
    Write-Host "Archivo encontrado: $csvPath" -ForegroundColor Green
    Write-Host "Tamano: $($fileInfo.Length) bytes" -ForegroundColor Gray
    Write-Host "Modificado: $($fileInfo.LastWriteTime)" -ForegroundColor Gray
}

Write-Host "Intentando ingest CSV..." -ForegroundColor Magenta
try {
    $ingestUrl = "http://localhost:8080/ingest/csv"
    $boundary = [System.Guid]::NewGuid().ToString()
    $fileBytes = [System.IO.File]::ReadAllBytes($csvPath)
    $fileEnc = [System.Text.Encoding]::GetEncoding('iso-8859-1').GetString($fileBytes)
    $LF = "`r`n"
    
    $bodyLines = ( 
        "--$boundary",
        "Content-Disposition: form-data; name=`"file`"; filename=`"$(Split-Path $csvPath -Leaf)`"",
        "Content-Type: application/octet-stream$LF",
        $fileEnc,
        "--$boundary--$LF" 
    ) -join $LF

    $result = Invoke-WebRequest -Uri $ingestUrl -Method Post -ContentType "multipart/form-data; boundary=`"$boundary`"" -Body $bodyLines -UseBasicParsing
    Write-Host "CSV ingestado correctamente: HTTP $($result.StatusCode)" -ForegroundColor Green
    Write-Host "Response: $($result.Content)" -ForegroundColor Gray
    
} catch {
    Write-Host "Error en ingest CSV: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.Exception.Response) {
        Write-Host "Status Code: $($_.Exception.Response.StatusCode)" -ForegroundColor Red
        Write-Host "Status Description: $($_.Exception.Response.StatusDescription)" -ForegroundColor Red
    }
    
    Write-Host "Intentando con curl..." -ForegroundColor Yellow
    try {
        $curlOutput = & curl -X POST -F "file=@$csvPath" -v http://localhost:8080/ingest/csv 2>&1
        Write-Host "Output curl: $curlOutput" -ForegroundColor Gray
    } catch {
        Write-Host "Curl tambien fallo: $($_.Exception.Message)" -ForegroundColor Red
    }
}

Write-Host "Revisando logs de los servicios..." -ForegroundColor Cyan
docker-compose logs servicioA --tail=20
Write-Host "---" -ForegroundColor Gray
docker-compose logs servicioB --tail=10

Write-Host "Iniciando test de carga - 10 requests simultaneos" -ForegroundColor Cyan

$url = "http://localhost:8081/weather/Madrid?date=2024-01-01&days=3"

$jobs = @()

for ($i = 1; $i -le 10; $i++) {
    $job = Start-Job -ScriptBlock {
        param($url, $i)
        try {
            $result = Invoke-WebRequest -Uri $url -UseBasicParsing
            Write-Host "Request $i`: HTTP $($result.StatusCode)" -ForegroundColor Green
        }
        catch {
            Write-Host "Request $i`: ERROR - $($_.Exception.Message)" -ForegroundColor Red
        }
    } -ArgumentList $url, $i
    $jobs += $job
}

Write-Host "Esperando que terminen los requests..." -ForegroundColor Yellow
$jobs | Wait-Job | Receive-Job

Write-Host "Test de carga completado" -ForegroundColor Green

Write-Host "Deteniendo docker-compose..." -ForegroundColor Red
docker-compose down
