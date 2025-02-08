@echo off
chcp 65001 > nul
setlocal

REM 构建 Go 程序
go build -o proxy.exe main.go

REM 检查 go build 是否成功
if %errorlevel% neq 0 (
    echo "构建失败"
    exit /b %errorlevel%
)

REM 打包文件和目录
tar -cvf proxy.tar config proxy.exe

REM 检查 tar 命令是否成功
if %errorlevel% neq 0 (
    echo "打包失败"
    exit /b %errorlevel%
)

REM 删除可执行文件
del proxy.exe

REM 检查删除操作是否成功
if %errorlevel% neq 0 (
    echo "删除失败"
    exit /b %errorlevel%
)

echo "操作完成"

endlocal