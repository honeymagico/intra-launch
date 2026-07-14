@echo off
setlocal
chcp 65001 >nul

set "APP_DIR=%~dp0"
set "APP_EXE=%APP_DIR%intra-launch.exe"
set "APP_NEW=%APP_DIR%intra-launch.exe.new"

if exist "%APP_NEW%" (
    if exist "%APP_EXE%" (
        del /f /q "%APP_EXE%" >nul 2>&1
        if exist "%APP_EXE%" goto :update_blocked
    )

    ren "%APP_NEW%" "intra-launch.exe" >nul 2>&1
    if exist "%APP_NEW%" goto :update_blocked
)

if not exist "%APP_EXE%" (
    echo 找不到內網應用啟動器：%APP_EXE%
    pause
    exit /b 1
)

start "" "%APP_EXE%"
exit /b 0

:update_blocked
echo.
echo 有版本更新，請關閉內網應用啟動器後再次啟動，以完成更新。
echo.
pause
exit /b 1
