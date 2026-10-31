# simulate.ps1 -> mete trafico falso pa que el dashboard se mueva en vivo
#
# uso, con el stack arriba y el dashboard abierto en localhost:3000/home
#   .\scripts\simulate.ps1 -Mode bruteforce
#   .\scripts\simulate.ps1 -Mode storm -Count 20
#   .\scripts\simulate.ps1 -Mode normal -Count 30
#   .\scripts\simulate.ps1 -Mode outage
#
# bruteforce -> logins fallidos seguidos contra un usuario, como un atacante probando passwords
# storm      -> eventos high y critical directo a events, salen los avisos arriba a la derecha
# normal     -> mezcla de logins buenos y malos con pausas random
# outage     -> apaga events, hace logins y lo prende. Muestra lo asincrono de kafka
#
# todo entra por el middleware en :9000 igual que el frontend

param(
    [ValidateSet("bruteforce", "storm", "normal", "outage")]
    [string]$Mode = "normal",

    # cuantas peticiones manda, en outage son los logins que hace con events apagado
    [int]$Count = 10,

    # pausa entre peticiones en ms, 0 = usa la de cada modo
    [int]$DelayMs = 0,

    [string]$Api = "http://localhost:9000"
)

$ErrorActionPreference = "Stop"

# PowerShell 5.1 manda Expect: 100-continue en cada POST
# con eso el middleware responde 200 aunque el backend haya dado 401, y se ven codigos falsos
# apagarlo hace que mande el body de una vez como un navegador
[System.Net.ServicePointManager]::Expect100Continue = $false

# el script vive en scripts/ y docker compose se corre desde la raiz
$root = Split-Path -Parent $PSScriptRoot

# usuario de prueba pa los logins
$demoUser = "simulador"
$demoPass = "demo1234"

function Pause-Ms([int]$ms) {
    Start-Sleep -Milliseconds $ms
}

# manda un POST con json y regresa el codigo http
# en PowerShell 5.1 un 401 o un 409 lanza excepcion, y aqui es normal que pase
# por eso se atrapa y se saca el codigo de la respuesta
function Send-Post([string]$path, $body) {
    try {
        $r = Invoke-WebRequest -Uri "$Api$path" -Method Post -ContentType "application/json" `
            -Body ($body | ConvertTo-Json) -UseBasicParsing -TimeoutSec 5
        return [int]$r.StatusCode
    }
    catch {
        if ($_.Exception.Response) {
            return [int]$_.Exception.Response.StatusCode
        }
        Write-Host "no hay conexion con $Api, esta arriba el stack?" -ForegroundColor Red
        exit 1
    }
}

# crea el usuario simulador, si ya existe da 409 y no pasa nada
function Ensure-DemoUser {
    $code = Send-Post "/api/accounts/register" @{ username = $demoUser; password = $demoPass }
    if ($code -eq 201) { Write-Host "usuario $demoUser creado" -ForegroundColor DarkGray }
}

function Login([string]$pass) {
    return Send-Post "/api/accounts/login" @{ username = $demoUser; password = $pass }
}

Write-Host "== simulate: $Mode ==" -ForegroundColor Magenta

switch ($Mode) {

    "bruteforce" {
        Ensure-DemoUser
        if ($DelayMs -eq 0) { $DelayMs = 300 }
        Write-Host "atacante probando passwords contra '$demoUser'"
        for ($i = 1; $i -le $Count; $i++) {
            $code = Login "intento$i"
            Write-Host ("  intento {0,2}/{1}  password 'intento{0}'  -> {2}" -f $i, $Count, $code)
            Pause-Ms $DelayMs
        }
        Write-Host "listo, deberian salir $Count login_failed en el dashboard"
    }

    "storm" {
        if ($DelayMs -eq 0) { $DelayMs = 800 }
        $types = @("escaneo_puertos_interno", "trafico_saliente_anomalo", "proceso_no_firmado", "consulta_dns_sospechosa")
        $sevs = @("high", "critical")
        Write-Host "tormenta de eventos graves directo a events"
        for ($i = 1; $i -le $Count; $i++) {
            $type = $types | Get-Random
            $sev = $sevs | Get-Random
            $code = Send-Post "/api/events" @{ type = $type; severity = $sev }
            Write-Host ("  {0,2}/{1}  {2}  {3}  -> {4}" -f $i, $Count, $sev.PadRight(8), $type, $code)
            Pause-Ms $DelayMs
        }
    }

    "normal" {
        Ensure-DemoUser
        Write-Host "trafico normal, 7 de cada 10 logins salen bien"
        for ($i = 1; $i -le $Count; $i++) {
            $ok = (Get-Random -Maximum 10) -lt 7
            $code = Login ($(if ($ok) { $demoPass } else { "mala$i" }))
            Write-Host ("  {0,2}/{1}  login {2}  -> {3}" -f $i, $Count, $(if ($ok) { "bueno" } else { "malo " }), $code)
            # pausa random pa que no se vea mecanico
            Pause-Ms $(if ($DelayMs -gt 0) { $DelayMs } else { Get-Random -Minimum 300 -Maximum 1500 })
        }
    }

    "outage" {
        Ensure-DemoUser
        Push-Location $root
        try {
            Write-Host "1. apagando events"
            docker compose stop events | Out-Null

            Write-Host "2. haciendo $Count logins con events apagado"
            for ($i = 1; $i -le $Count; $i++) {
                $code = Login $demoPass
                Write-Host ("  login {0,2}/{1} -> {2}  accounts responde normal" -f $i, $Count, $code)
                Pause-Ms 200
            }

            Write-Host "3. mensajes esperando en kafka, el LAG es lo que falta por leer" -ForegroundColor Yellow
            docker exec kafka /opt/kafka/bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 --describe --group events-writer

            Read-Host "enter pa prender events otra vez"

            Write-Host "4. prendiendo events"
            docker compose start events | Out-Null
            Write-Host "en unos segundos el dashboard muestra los $Count login_success de golpe" -ForegroundColor Green
        }
        finally {
            Pop-Location
        }
    }
}
