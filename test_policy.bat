@echo off
chcp 65001 >nul
echo ========================================
echo 政策向量化功能测试脚本
echo ========================================
echo.

set SERVER=http://localhost:8080

echo [1/3] 测试服务健康状态...
curl -s %SERVER%/health
echo.
echo.

echo [2/3] 更新政策到向量数据库（这可能需要几分钟）...
echo 提示：首次使用前必须执行此步骤
curl -X POST %SERVER%/api/policy/update
echo.
echo.

echo [3/3] 测试政策搜索...
echo 查询关键词：就业补贴
curl -s "%SERVER%/api/policy/search?query=就业补贴&topK=3"
echo.
echo.

echo ========================================
echo 测试完成！
echo ========================================
echo.
echo 你也可以通过对话接口测试：
echo curl -X POST %SERVER%/v1/chat/completions -H "Content-Type: application/json" -d "{\"model\":\"qd-job-turbo\",\"messages\":[{\"role\":\"user\",\"content\":\"我想了解就业见习补贴政策\"}]}"
echo.
pause
