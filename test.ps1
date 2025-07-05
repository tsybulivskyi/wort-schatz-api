# Invoke-WebRequest "http://localhost:8080/hello"

# # Test POST /words endpoint
$word = @{ 
    original    = "Hallo"
    translation = "Hello"
    tags        = @("greeting", "common")
} | ConvertTo-Json

$response = Invoke-WebRequest -Uri http://localhost:8080/words -Method Post -Body $word -ContentType 'application/json' -SkipCertificateCheck
Write-Host "/words status: $($response.StatusCode)"
Write-Host "Response body: $($response.Content)"


$response = Invoke-WebRequest "http://10.0.2.2:8080/words"
$words = $response.Content | ConvertFrom-Json
Write-Host "Words list:`n$(($words | ConvertTo-Json -Depth 10 -Compress:$false))"


