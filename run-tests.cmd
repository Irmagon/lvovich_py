@echo off
chcp 65001 >nul
setlocal EnableDelayedExpansion
cd /d "%~dp0"

set "PY=%~dp0python_embedded\python.exe"
if not exist "%PY%" set "PY=python"

set "OUT_DIR=%~dp0test-output"
if not exist "%OUT_DIR%" mkdir "%OUT_DIR%"

for /f "usebackq delims=" %%i in (`powershell -NoProfile -Command "(Get-Date).ToString('yyyy-MM-dd_HHmmss')"`) do set "STAMP=%%i"
set "LOG=%OUT_DIR%\tests_%STAMP%.log"

echo ===================================================
echo  TESTS: %PY% -m pytest tests -v
echo  LOG:   %LOG%
echo ===================================================
"%PY%" -m pytest tests -v > "%LOG%" 2>&1
set "PYRESULT=%errorlevel%"
type "%LOG%"

if not "%PYRESULT%"=="0" goto :testfail
goto :testok

:testfail
echo.
echo [ERROR] TESTS FAILED - %DATE% %TIME%
set "RC=TEST_FAIL:%PYRESULT%"
goto :summary

:testok
echo.
echo [OK] ALL TESTS PASSED - %DATE% %TIME%
set "RC=OK"

:summary
set "RESFILE=%OUT_DIR%\_LAST_RESULT.txt"
echo %DATE% %TIME% RC=!RC! > "%RESFILE%"
echo %LOG%>> "%RESFILE%"
echo.
echo ===================================================
echo  RESULT: !RC!
echo  Log file: %LOG%
echo ===================================================
endlocal
