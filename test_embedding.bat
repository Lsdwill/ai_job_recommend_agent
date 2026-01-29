@echo off
chcp 65001 >nul
echo ========================================
echo 测试 Embedding API
echo ========================================
echo.

echo 测试请求...
powershell -Command "$body = '{\"inputs\":\"就业补贴政策测试\"}'; Invoke-WebRequest -Uri 'http://39.98.44.136:6017/emb/embed' -Method POST -ContentType 'application/json' -Body $body -UseBasicParsing | Select-Object -ExpandProperty Content | ConvertFrom-Json | Select-Object -First 10"

echo.
echo ========================================
echo 测试完成！
echo ========================================
pause
