@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion
cd /d %~dp0

if not exist config mkdir config
if not exist data mkdir data
if not exist config\db.yaml echo driver: sqlite> config\db.yaml

if not exist runtime\mysql\bin\mysqld.exe goto startnebula
if exist data\mysql\mysql goto mysqlalready

echo [1/3] 初始化随包 MySQL...
mkdir data\mysql 2>nul
runtime\mysql\bin\mysqld.exe --no-defaults --initialize-insecure --basedir="%cd%\runtime\mysql" --datadir="%cd%\data\mysql" >nul 2>&1
if not exist data\mysql\mysql (
  echo        ✖ 初始化失败，回退 SQLite（见 data\mysql.log）
  goto startnebula
)
start "" /b runtime\mysql\bin\mysqld.exe --no-defaults --basedir="%cd%\runtime\mysql" --datadir="%cd%\data\mysql" --port=3307 --bind-address=127.0.0.1 --pid-file="%cd%\data\mysql.pid" --console >> data\mysql.log 2>&1
set /a n=0
:waitmy
runtime\mysql\bin\mysql.exe --no-defaults -h 127.0.0.1 -P 3307 -u root -e "SELECT 1" >nul 2>&1 && goto myok
set /a n+=1
if !n! geq 30 goto myfail
timeout /t 1 /nobreak >nul
goto waitmy
:myfail
echo        ✖ MySQL 启动失败，回退 SQLite（见 data\mysql.log）
echo driver: sqlite> config\db.yaml
goto startnebula
:myok
runtime\mysql\bin\mysql.exe --no-defaults -h 127.0.0.1 -P 3307 -u root -e "CREATE DATABASE IF NOT EXISTS nebula CHARACTER SET utf8mb4" >nul 2>&1
(echo driver: mysql
 echo host: 127.0.0.1
 echo port: 3307
 echo user: root
 echo password: ""
 echo database: nebula) > config\db.yaml
echo        ✓ MySQL 就绪（127.0.0.1:3307）
goto startnebula

:mysqlalready
echo [1/3] 复用已有 MySQL 数据目录
start "" /b runtime\mysql\bin\mysqld.exe --no-defaults --basedir="%cd%\runtime\mysql" --datadir="%cd%\data\mysql" --port=3307 --bind-address=127.0.0.1 --pid-file="%cd%\data\mysql.pid" --console >> data\mysql.log 2>&1
goto startnebula

:startnebula
echo [2/3] 启动 Nebula...
start "" /b nebula.exe > nebula.log 2>&1
echo [3/3] 等待就绪...
set /a m=0
:waitn
powershell -NoProfile -Command "try{Invoke-RestMethod http://localhost:8080/api/health|Out-Null;exit 0}catch{exit 1}" >nul 2>&1 && goto appok
set /a m+=1
if !m! geq 30 goto appfail
timeout /t 1 /nobreak >nul
goto waitn
:appfail
echo ✖ 启动失败，nebula.log 尾部：
powershell -NoProfile -Command "Get-Content nebula.log -Tail 15"
pause
exit /b 1
:appok
echo.
echo ✅ Nebula Tools 已启动
echo    前台   http://localhost:8080
echo    管理   右上角 ⚙️（admin / admin123）
echo    停止   双击 stop.bat
pause
