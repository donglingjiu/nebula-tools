# 案例：AI辅助交付的五段流水线与守卫链（nebula-tools 7-3）

> 来源：nebula-tools 工具域战役 · 5 战役 4 commit 全程 AI 主驾
> 调用时机：AI 主导的仓库级功能交付——取证→patch→E2E→守卫→commit 流水线设计

## 可复用模式
1. **证据链编号制**：每执行段前先跑取证命令（证N编号 [证31]~[证51]），判读后再写 patch——不猜契约，字段全部从源码取证（opts 键面/响应形态/依赖库）
2. **python3 inline patch**：heredoc + pathlib + 双 assert（锚点存在+防重入）+ print 确认——比 sed 多行安全、可重复、自带防呆
3. **守卫链与链结构**（零误放行的关键）：`[ E2E PASS计数 ] && [ patch grep 计数 ] && [ go build ] && [ go test ] && { commit } || echo '守卫拦截: 回贴FAIL'`——实战两次拦截皆真问题（E2E 实例未起 E1=000、totp 测试契约错），拦截后取证修复再放行
4. **E2E mark 文件**：PASS/FAIL 逐条 append /tmp/mark，`grep -c '^PASS'` 进守卫判据——断言可累积可审计
5. **E2E 实例三隔离**：独立二进制名（-o /tmp/nova-e2e73 防 pkill 误杀开发实例）+ 独立 data dir + PORT env（源码唯一注入点先取证）
6. **断言选型规约**：强断言特征串优先取 registry desc 亲供值；实现未知降级长度/非空；功能真值（magic bytes）与规范符合性（MIME 头）分开断言
7. **内网工具链兜底**：base64/xxd/node 不存在 → python3 stdlib + gofmt + grep 自证
8. **push 重试链**：`git push || git push`（内网 443 SSL 抖动常态，三次内成功）
9. **管道陷阱**：`git clone | tail -1` 管道吃掉 clone 退出码导致重试永不触发——`set -o pipefail`，或干脆用 rev-parse 复查代替管道退出码判定
10. **内网缺 timeout → git 原生防卡**：`git -c http.lowSpeedLimit=1000 -c http.lowSpeedTime=20 clone --depth 1`——低于 1KB/s 持续 20 秒自动中止，不依赖外部 timeout；连接拒绝秒败、真卡死 40 秒内了断
11. **长粘贴分段**：多段 heredoc 长粘贴中断 → shell 卡 `>` 续行提示符，Ctrl+C 丢弃缓冲即恢复；长内容拆两次贴降低暴露

## 反模式记录
- 守卫判据写 `nova-fx>=5` 但 patch 内实仅 1 次 → 判据设计期必须 grep 实数
- 假设"生成型工具零输入可跑" → totp 密钥必需；用例输入按契约（NR_SCHEMA）填，不按直觉省略
- 全量克隆在抖动链路 → `--depth 1` 浅克隆最小暴露窗口
- 多行粘贴前不确认 pwd → 目录错位 redirect 全灭（幸有防呆未污染）；贴前先 pwd

## 效率
5 战役 4 commit；27 工具测试矩阵 0.021s；守卫拦截→修复→放行平均一轮。
