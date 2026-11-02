# 案例：双实例交付验证与收官卫生三判据（nebula-tools 7-3）

> 来源：nebula-tools 工具域战役 · E2E 验证设计 + .bak 清理收官
> 调用时机：在线服务的交付验证（不扰运行环境）+ 仓库卫生收官操作

## 交付验证模式
1. **双实例并存**：8080 开发实例全程健康基线（每段前后 curl health 200），9091 E2E 实例独立起停——验证不扰运行环境
2. **认证类端点 E2E 五连标准套**：E1 无 cookie→401（门生效）；E2 register+login→cookie 落地；E3 带参功能断言（csv 去重后行数=3，选双实现语义皆过的数值）；E4 二进制 magic bytes（od -tx1 取 ffd8ff）+ Content-Type 双断言；E5 JSON 结构断言（duplicate_rows:1）
3. **误报取证修正**：FAIL E4:ffd8ff 字面即"magic 正确功能成功、仅断言错"——先 grep 响应头取证（实锤 image/jpg 非规范 MIME）再修断言为 image/(jpeg|jpg)，不盲目重跑
4. **实例清理协议**：kill pid → pgrep 残留检查 → 开发实例终检 200，三步确认

## 卫生收官三判据

| 判据 | 手段 | 决策 |
|---|---|---|
| 跟踪态 | git ls-files | 未跟踪→物理删；已跟踪→git rm |
| 进程占用 | ps -ef | 运行中二进制=实例本体→保留 |
| ignore 覆盖 | git check-ignore -v | 已覆盖→无需补口 |

- 实战：9 个 .bak 物理清 1.1MB（删除前后 find wc -l 快照 9→0）；运行中二进制（PID 实查）与可重建产物保留——本地工作集不动
- 每次删除后 health 检查——"运行实例零受扰"作为卫生操作硬约束

## 沉淀
- .gitignore 闭环模板：`*.bak*` + 日志 + 二进制产物（check-ignore -v 逐文件验证）
- 删除操作双快照：前计数+后计数，量化体积收益留档
