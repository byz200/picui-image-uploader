# setup-github.ps1 - 一键初始化并推送 Picui 项目到 GitHub
# 前置：已安装 git 与 GitHub CLI (gh) 并完成 gh auth login
#
# 用法：
#   ./setup-github.ps1 -RepoName picui-uploader -Visibility public
#   ./setup-github.ps1 -RepoName picui-uploader -Visibility private
#
# 参数：
#   -RepoName      仓库名称（默认 picui-uploader）
#   -Visibility    public 或 private（默认 public）
#   -Description   仓库描述
#   -SkipProtect   跳过分支保护规则配置

param(
    [string]$RepoName = "picui-image-uploader",
    [ValidateSet('public', 'private')][string]$Visibility = "public",
    [string]$Description = "Picui 图床上传工具 - Wails v2 + Go + Vue3 桌面端图床上传工具（Windows x86/x64）",
    [switch]$SkipProtect
)

$ErrorActionPreference = "Stop"
Set-Location -Path $PSScriptRoot

function Assert-Cmd {
    param([string]$Name)
    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        Write-Error "未检测到 $Name，请先安装。详见 README.md 的「GitHub 发布流程」一节。"
        exit 1
    }
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Picui 项目 GitHub 初始化" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "仓库名: $RepoName"
Write-Host "可见性: $Visibility"
Write-Host ""

Assert-Cmd git
Assert-Cmd gh

# 1. 校验 gh 登录状态
Write-Host "[1/6] 校验 GitHub CLI 登录状态..." -ForegroundColor Yellow
$auth = gh auth status 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "未登录 GitHub，请先运行: gh auth login" -ForegroundColor Red
    Write-Host $auth
    exit 1
}
Write-Host "已登录。" -ForegroundColor Green

# 2. git 初始化
Write-Host "[2/6] 初始化 git 仓库..." -ForegroundColor Yellow
if (-not (Test-Path ".git")) {
    git init
    git branch -M main
} else {
    Write-Host "已存在 .git，跳过 init。"
}

# 3. 提交代码
Write-Host "[3/6] 暂存并提交代码..." -ForegroundColor Yellow
git add -A
$status = git status --porcelain
if ($status) {
    git commit -m "feat: 初始化 Picui 图床上传工具（Wails v2 + Go + Vue3）

- 双站点隔离（picui.cn / v2.picui.cn）配置与历史
- 拖拽/剪贴板/截图三入口上传，异步并发队列与重试
- 客户端压缩（JPG/PNG）、自动复制 Markdown
- 相册管理、图库浏览、上传历史
- 系统托盘常驻、全局快捷键、明暗主题
- GitHub Actions 双架构（amd64/386）自动构建与发布"
} else {
    Write-Host "无变更需要提交。"
}

# 4. 创建 GitHub 仓库
Write-Host "[4/6] 创建 GitHub 仓库..." -ForegroundColor Yellow
$existing = gh repo view $RepoName --json name 2>$null
if ($existing) {
    Write-Host "仓库 $RepoName 已存在，跳过创建。" -ForegroundColor Green
} else {
    gh repo create $RepoName --$Visibility --description $Description --source=. --remote=origin --push
    Write-Host "仓库已创建并推送。" -ForegroundColor Green
}

# 5. 推送
Write-Host "[5/6] 推送代码到 origin/main..." -ForegroundColor Yellow
git remote remove origin 2>$null
$user = (gh api user --jq .login)
git remote add origin "https://github.com/$user/$RepoName.git"
git push -u origin main

# 6. 分支保护
if (-not $SkipProtect) {
    Write-Host "[6/6] 配置 main 分支保护规则..." -ForegroundColor Yellow
    $protection = @{
        required_status_checks = @{
            strict = $true
            contexts = @("Build (x64)", "Build (x86)")
        }
        enforce_admins = $false
        required_pull_request_reviews = $null
        restrictions = $null
        allow_force_pushes = $false
        allow_deletions = $false
    } | ConvertTo-Json -Depth 5 -Compress
    try {
        $protection = $protection -replace '"required_pull_request_reviews":null', '"required_pull_request_reviews":{}'
        gh api -X PUT "repos/$user/$RepoName/branches/main/protection" -f "required_status_checks[strict]=true" -f "required_status_checks[checks][]=Build (x64)" -f "required_status_checks[checks][]=Build (x86)" -f "enforce_admins=false" -f "restrictions=" 2>$null | Out-Null
        if ($LASTEXITCODE -eq 0) {
            Write-Host "分支保护已配置（要求 CI 状态检查通过）。" -ForegroundColor Green
        } else {
            Write-Host "分支保护配置失败（private 仓库或权限不足时可能跳过），可稍后在 GitHub 网页配置。" -ForegroundColor DarkYellow
        }
    } catch {
        Write-Host "分支保护配置异常，可稍后手动配置: $_" -ForegroundColor DarkYellow
    }
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "  完成！" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host "仓库地址: https://github.com/$user/$RepoName" -ForegroundColor Cyan
Write-Host ""
Write-Host "下一步：" -ForegroundColor White
Write-Host "  - 推送代码后 GitHub Actions 会自动触发构建（见 Actions 标签页）"
Write-Host "  - 发布 Release: git tag v1.0.0 ; git push origin v1.0.0"
Write-Host "  - Release 会自动附带 x64 / x86 两份 exe"
