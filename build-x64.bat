@echo off
chcp 65001 >nul
setlocal
cd /d "%~dp0"

echo ==========================================================
echo   Picui 图床上传工具 - x64 (windows/amd64) 编译
echo ==========================================================

where wails >nul 2>nul
if %errorlevel%==0 goto wails_build
goto go_build

:wails_build
echo [方案A] 检测到 wails CLI，使用 wails build 编译...
wails build -platform windows/amd64 -clean -ldflags "-s -w -H windowsgui"
if %errorlevel%==0 (
    echo.
    echo [成功] 输出文件: build\bin\PicuiUploader.exe
    goto :done
)
echo [wails build 失败，回退到 go build 方案]
goto go_build

:go_build
echo [方案B] 使用 go build 直接编译（需先完成前端构建）...
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo [错误] 未检测到 go，请先安装 Go 1.22+: https://go.dev/dl/
    exit /b 1
)
where npm >nul 2>nul
if %errorlevel% neq 0 (
    echo [错误] 未检测到 npm，请先安装 Node.js 18+: https://nodejs.org/
    exit /b 1
)

echo [1/4] 安装前端依赖...
pushd frontend
call npm install
if %errorlevel% neq 0 ( echo npm install 失败 & popd & exit /b 1 )
echo [2/4] 构建前端...
call npm run build
if %errorlevel% neq 0 ( echo 前端构建失败 & popd & exit /b 1 )
popd

echo [3/4] 生成图标资源 (icon_windows.syso)...
if exist icon_windows.syso del icon_windows.syso
go install github.com/akavel/rsrc@latest 2>nul
if exist "%USERPROFILE%\go\bin\rsrc.exe" (
    "%USERPROFILE%\go\bin\rsrc.exe" -ico build\windows\icon.ico -arch amd64 -o icon_windows.syso
    echo 图标资源已生成
) else (
    echo [提示] rsrc 工具不可用，跳过图标资源生成（exe 仍可用，但资源管理器图标为默认）
)

echo [4/4] 编译 Go (GOOS=windows GOARCH=amd64)...
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0
if not exist build\bin mkdir build\bin
go mod tidy
go build -ldflags "-s -w -H windowsgui" -o "build\bin\PicuiUploader-x64.exe" .
if %errorlevel%==0 (
    echo.
    echo [成功] 输出文件: build\bin\PicuiUploader-x64.exe
    goto :done
)
echo [错误] 编译失败
exit /b 1

:done
echo.
echo 完成。生成的 exe 为单文件绿色程序，可直接运行。
endlocal
