# B3 治理战役战报（v2.9.1 - v2.9.7）

## 战线与终态
- 拆甲：raw 27→19→13→12 终档。批1 八刀（mt系列/m0/mlauto/grid3/grid2，零竞争零碰撞）；批2 六刀（flex1/py4/ico32/ico31/fs15/fs13，零碰撞+六表零冲突）；批3 改判豁免 12（opt-input 7 / searchbar-full 2 / hero-strip 2 / c-ink 1）——威胁扫描 66 呈堂（.searchbar input、.x strong 家族为结构性实锤；扫描器逗号组/标签限定两类假阳自首记档），法理：组件甲是 B2 忠实翻译内联优先级的载体。tokens 终态 = 纯设计令牌 + 12 组件豁免区 + 防回潮卫兵（want=12）。
- 9px 归一：census（全站 var 消费面=1）→ A/B → v2.9.5 采纳 10px（锚点同步）→ v2.9.7 孪生合并 s9→s10（外部引用为零实证）。
- #2d3d60 归一：v2.9.6 新立 --nv-ink，nv-c-ink 规则值 var 化（零视觉）；第 2 处为 B2 注释提及（CTX 定性）。

## 制度产出（军规 #5-#11）
- #5 行数≠次数（grep -c vs grep -o|wc -l，三犯）；#6 附件：双引号内 $() 使内层单引号失效（安全形态=顶层赋值）；#7 禁百分号格式化构造落盘文本；#8 写盘当场 bash -n；#9 环境基线 Py3.6；#10 依赖块单块分阶段断言（①炸②照跑案）；#10 附件：python open() 不自建目录，写盘前 mkdir/os.makedirs 先行（战报首跑阵亡案）；#11 推送真相在服务端（range 行即回执）。

## 闸门战果
- POSTx27 写盘前拦截；PCT 闸自咬军规条文；假账 changelog 被 ! 引爆挡在 git 门外；RIVALS-REVIEW-NEEDED 拒绝批3 手术；push 假失败识破；docs 目录缺失拦截战报（零损伤）。
- 全程零返工零回滚；32 发体检全绿；三页截图（analytics/workflows/ai）视觉抽验通过。

## 遗留与新账
- 仪器卫生：verify 锚点多为 grep -c 数行（单行定义下安全），可统一升级为数次；
- B4 候选：~98 处硬编码 9px 归一到令牌（15 文件，nova-v6 40 + styles 21 主力）；
- C 线：模型接入层已由 llm.html 交付（/api/llm/chat 同源代理 + Key 服务端保险柜 + 多档案故障切换），参考架构 ai-toolsBox（SQLite 先行/DB 地址可配/一键迁移）——C1 从 ai 工作台打头。

## 附件2：军规 #12 内容自愈写盘（2026-10-23）
- 案：传输层向 CJK 串内插空格，ASCII 无恙→带病执行落盘；既有闸门对中文完整性全盲（连过三发入库）。
- 法：凡含中文落盘，python 块内 normalize（CJK 邻接空格坍缩）+ 关键短语断言双闸；超长纯文本可改 base64 载荷（纯 ASCII 免疫）。
- 修复：本轮 0 处（CHANGELOG 仅限我方 v2.9.4+ 区，原史区不碰）。

## 附件2 补记：乱码案结案（2026-10-23）
- 法证定谳：磁盘零损伤，伪影只活在显示层，落盘字节全程干净；前轮损伤表撤回，修复实数 0。
- 全仓余 4 处 CJK 邻接空格系原史合法排版，永不改动；#12 仅作用新落盘内容。
- 附件3 围栏洁癖：交付落盘内容禁含三反引号——内嵌围栏击穿聊天围栏致复制残缺，本轮起执行。

## 附件4：军规 #13 普查禁截断（2026-10-23）
- 案：源码失踪悬案真相——find 普查接 head -20，GOROOT 的 go/test 千余 .go 按字母序占满配额，本仓 15 个 .go 被截断在证据外，误判源码不在本机；实为一直在本仓。
- 法：普查命令（find/grep 全域扫描）先计数后输出，禁接 head/tail；head 仅许已定向下钻单文件。
- 附记：二进制口供定驱动——CGO_ENABLED=0 烙死于构建，mattn 出局，modernc.org/sqlite（驱动名 sqlite）当选。

## 附件6：军规 #14 heredoc 禁 TAB（2026-10-26）
- 案：交互式 bash 的 heredoc 逐行经 readline，行首 TAB 触发补全并吞字（Display all 313 即列家目录），db.go 带伤写盘被 go test 编译闸当场拒收。
- 法：粘贴内容零 TAB，Go 缩进用空格；写盘前断言内容无 TAB；gofmt 在盘上归一属安全侧。

## 附件7：军规 #15 闸门环境同源（2026-10-26）
- 案：go test 默认 CGO_ENABLED=1，拉起本机老 gcc 编 runtime/cgo，cc1 不识 -fno-stack-protector 炸膛；生产 build.sh 烙死 CGO_ENABLED=0 从未中招——闸门测的是另一种产物。
- 法：本盒 go test/vet 一律前置 CGO_ENABLED=0 与生产同源；原则：闸门环境必须等于产物环境，否则绿灯是假绿灯。
