# -*- mode: python ; coding: utf-8 -*-
# PyInstaller spec: сборка fioincline.exe (one-file).
# Запуск:  python_embedded\python.exe -m PyInstaller build.spec

a = Analysis(
    ["main.py"],
    pathex=["."],
    binaries=[],
    datas=[
        ("fioincline/wsdl/service.wsdl", "fioincline/wsdl"),
    ],
    hiddenimports=[
        "uvicorn.logging",
        "uvicorn.loops.auto",
        "uvicorn.protocols.http.auto",
        "uvicorn.protocols.websockets.auto",
        "uvicorn.lifespan.on",
    ],
    hookspath=[],
    hooksconfig={},
    runtime_hooks=[],
    excludes=["pytest"],
    noarchive=False,
)

pyz = PYZ(a.pure)

exe = EXE(
    pyz,
    a.scripts,
    a.binaries,
    a.datas,
    [],
    name="fioincline",
    debug=False,
    bootloader_ignore_signals=False,
    strip=False,
    upx=True,
    upx_exclude=[],
    runtime_tmpdir=None,
    console=True,
    disable_windowed_traceback=False,
    argv_emulation=False,
    target_arch=None,
    codesign_identity=None,
    entitlements_file=None,
)
