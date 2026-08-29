@echo off
chcp 65001 >nul
setlocal
cd /d "%~dp0"

set "PY=%~dp0python_embedded\python.exe"
if not exist "%PY%" set "PY=python"

echo Сборка fioincline.exe (PyInstaller one-file)...
"%PY%" -m PyInstaller --noconfirm --clean build.spec
if errorlevel 1 (
    echo ОШИБКА: сборка не удалась.
    exit /b 1
)

echo Готово: dist\fioincline.exe
endlocal
