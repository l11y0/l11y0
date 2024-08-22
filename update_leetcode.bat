@echo off
cd /d C:\Users\黃韻如\Documents\l11y0

echo 正在拉取最新的遠端更改...
git pull origin master

echo 正在執行 Go 程式...
go run main.go
if %errorlevel% neq 0 (
    echo 執行 Go 程式時發生錯誤 >> update_error.log
    exit /b 1
)

echo 正在添加所有更改...
git add .

echo 正在提交更改...
git commit -m "更新 README.md 和 main.go"

echo 正在推送更改到 GitHub...
git push origin master

echo 更新成功完成 >> update_log.txt
echo 更新過程已完成。
pause