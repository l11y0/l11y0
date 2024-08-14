@echo off
cd /d C:\Users\黃韻如\Documents\leetcode
go run main.go
if %errorlevel% neq 0 (
    echo 執行 Go 程式時發生錯誤 >> update_error.log
    exit /b 1
)
echo 更新成功完成 >> update_log.txt