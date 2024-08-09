@echo off
cd /d C:\Users\黃韻如\Documents\l11y0
go build -o main.exe
main.exe
git add README.md
git commit -m "更新資料"
git push