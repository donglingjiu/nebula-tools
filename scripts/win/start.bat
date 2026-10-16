@echo off
chcp 65001 >nul
cd /d %~dp0
if exist runtime\mysql\bin\mysqld.exe if exist data\mysql\mysql (
  start "" /b runtime\mysql\bin\mysqld.exe --no-defaults --basedir="%cd%\runtime\mysql" --datadir="%cd%\data\mysql" --port=3307 --bind-address=127.0.0.1 --pid-file="%cd%\data\mysql.pid" --console >> data\mysql.log 2>&1
  timeout /t 3 /nobreak >nul
)
start "" /b nebula.exe >> nebula.log 2>&1
echo 已启动: http://localhost:8080
pause
