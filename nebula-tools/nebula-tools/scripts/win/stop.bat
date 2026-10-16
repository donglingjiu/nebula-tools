@echo off
chcp 65001 >nul
taskkill /IM nebula.exe /F >nul 2>&1 && echo ✓ nebula 已停止 || echo · nebula 未在运行
if exist data\mysql.pid (
  for /f %%p in (data\mysql.pid) do taskkill /PID %%p /F >nul 2>&1
  echo ✓ 随包 MySQL 已停止
) else taskkill /IM mysqld.exe /F >nul 2>&1
pause
