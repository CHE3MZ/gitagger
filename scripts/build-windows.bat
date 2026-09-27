@echo off
REM Build gitagger for Windows (amd64 + arm64).
REM Usage: scripts\build-windows.bat
setlocal
cd /d "%~dp0.."
if not exist dist mkdir dist
echo ==^> Building gitagger for windows/amd64
set GOOS=windows
set GOARCH=amd64
go build -trimpath -o dist\gitagger-windows-amd64.exe .\cmd\gitagger
if errorlevel 1 exit /b 1
echo ==^> Building gitagger for windows/arm64
set GOARCH=arm64
go build -trimpath -o dist\gitagger-windows-arm64.exe .\cmd\gitagger
if errorlevel 1 exit /b 1
echo OK:
dir dist\gitagger-windows-*.exe
