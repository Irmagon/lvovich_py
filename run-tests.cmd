@echo off
chcp 65001 >nul
setlocal EnableDelayedExpansion
cd /d "%~dp0"

set "OUT_DIR=%~dp0test-output"
if not exist "%OUT_DIR%" mkdir "%OUT_DIR%"

for /f "usebackq delims=" %%i in (`powershell -NoProfile -Command "(Get-Date).ToString('yyyy-MM-dd_HHmmss')"`) do set "STAMP=%%i"
set "LOG=%OUT_DIR%\tests_%STAMP%.log"

echo ===================================================
echo  BUILD: go build -o fioincline-server.exe ./cmd/server
echo ===================================================
call go build -o fioincline-server.exe ./cmd/server
if errorlevel 1 goto :buildfail

echo.
echo ===================================================
echo  TESTS: go test ./... -v
echo  LOG:   %LOG%
echo ===================================================
call go test ./... -v > "%LOG%" 2>&1
set "GORESULT=%errorlevel%"
type "%LOG%"

if not "%GORESULT%"=="0" goto :testfail
goto :testok

:buildfail
echo.
echo [ERROR] BUILD FAILED - %DATE% %TIME%
set "RC=BUILD_FAIL"
goto :summary

:testfail
echo.
echo [ERROR] TESTS FAILED - %DATE% %TIME%
set "RC=TEST_FAIL:%GORESULT%"
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