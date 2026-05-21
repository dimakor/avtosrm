param(
    [string]$HostName = "88.218.66.102",
    [string]$SSHPort = "22",
    [string]$Port = "4569",
    [string]$User = "dimakor",
    [string]$SSHKey = "$env:USERPROFILE\.ssh\id_ed25519"
)

$ErrorActionPreference = "Stop"

Write-Host "=== Building Linux binary ==="
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -ldflags="-s -w" -o avtosrm .
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
Write-Host "Build OK: $((Get-Item avtosrm).Length) bytes"

Write-Host "=== Creating remote directory ==="
ssh -p $SSHPort -i $SSHKey "$User@$HostName" "mkdir -p /opt/avtosrm"

Write-Host "=== Uploading files ==="
scp -P $SSHPort -i $SSHKey avtosrm "$User@$HostName:/opt/avtosrm/"
scp -P $SSHPort -i $SSHKey .env "$User@$HostName:/opt/avtosrm/"
scp -P $SSHPort -i $SSHKey deploy/avtosrm.service "$User@$HostName:/tmp/avtosrm.service"

Write-Host "=== Installing systemd service ==="
ssh -p $SSHPort -i $SSHKey "$User@$HostName" @"
sudo mv /tmp/avtosrm.service /etc/systemd/system/avtosrm.service
sudo systemctl daemon-reload
sudo systemctl enable avtosrm
sudo systemctl restart avtosrm
sudo systemctl status avtosrm --no-pager
"@

Write-Host "=== Deployment complete ==="
Write-Host "Test: curl -H 'Authorization: Bearer <key>' 'http://$HostName/api/v1/route?from=54.72845,55.9486&to=55.7821,49.12377'"
