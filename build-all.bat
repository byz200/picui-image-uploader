@echo off
chcp 65001 >nul
setlocal
cd /d "%~dp0"

echo ==========================================================
echo   Picui 图床上传工具 - 双架构批量编译 (x64 + x86)
echo ==========================================================
echo.

echo ============ [1/2] 编译 x64 (windows/amd64) ============
call "%~dp0build-x64.bat"
if %errorlevel% neq 0 (
    echo x64 编译失败，终止。
    exit /b 1
)
echo.
echo ============ [2/2] 编译 x86 (windows/386) ============
call "%~dp0build-x86.bat"
if %errorlevel% neq 0 (
    echo x86 编译失败。
    exit /b 1
)

echo.
echo ==========================================================
echo   全部编译完成
echo ==========================================================
echo 输出目录: build\bin\
dir /b build\bin\*.exe 2>nul
endlocal
