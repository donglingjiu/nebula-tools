# 案例：工具域四维盘点与产品决策遗留（nebula-tools 7-3）

> 来源：nebula-tools 工具域战役 · B-1 全域盘点 → 接线优先级 → 决策留档
> 调用时机：对"后端能力×前端入口"做对账盘点、或处理功能重叠的双轨域

## 盘点方法论（四维交叉）
产品盘点 = **路由注册面 × 前端入口 × 用户价值 × 设计文档**，只看路由注册面会漏判：
- 三族定谳：原判"双轨"（文本+文件）漏了 /api/fx/*（5 个 PDF 端点前后端已全通）——盘点必须前后端双向对账
- 孤岛工具：/api/t/* 3 工具（CSV清洗/图片处理/Excel剖析）后端完整、前端零入口——**价值已建成但未触达，接线是最高 ROI 动作**
- 双轨重叠：/api/upload/{tool} 4 工具与 /api/fx 功能域重叠且零前端——合并/并存/下线待拍板

## 优先级排序
1. "有后端无前端"最高（价值已建成，接线即收割）
2. "有前端无后端"次之（占位检查）
3. "无后端无前端"→规划中标记

## 用户价值交付表

| 工具 | 用户场景 | 之前 | 之后 |
|---|---|---|---|
| CSV 清洗 | 去重/过滤/列裁剪/GBK | 无入口 | tools.html 上传即用 |
| 图片处理 | 缩放/格式转换/质量 | 无入口 | 同上 |
| Excel 剖析 | 列统计/重复行/MD 报告 | 无入口 | 同上（默认 JSON 展示，md 可下载） |

## 设计意图考古（防误判）
公开工具箱无登录门 ≠ 安全缺陷：DEPLOY.md 明示"公开工具箱无需登录，纯前端处理不落盘"——**先考古设计文档再定性"缺失"**，避免把设计意图当漏洞修。

## 决策遗留（下阶段议题模板）
1. image/jpg 非规范 MIME——规范符合性 vs 一行修成本
2. /api/upload 文件轨去留——【已拍板 A 迁后撤】file.pdf/excel/word/img 四剖析工具(纯stdlib, [证86])先收编进 /api/t 家族(对齐 t/xlsx-profile 先例), 再退役 upload 泛路由+runToolFile+注册表4条(27→23)+e2e 3调用点改造+防复活棘轮。TODO: 排期在 11-03 前端回归修复之后
   - 11-03 TODO①(定谳修正): 根因=静态资源无Cache-Control+浏览器陈旧副本(mtime倒退304陷阱), 非磁盘HEAD半成品——①C经[证108]全diff审查为纯增量且语法完好, 116校准合法; 修复=no-cache头+gate防御化+登录显式断言+浏览器清缓存; TODO③立项: 资产URL版本化(?v=hash)根治304陷阱
   - 11-03 TODO② 浏览器级冒烟: 静态grep棘轮查不出JS运行时错, 立项轻量前端冒烟(页面加载+核心工具点击)入verify可选段——与 /api/fx 合并 or 下线（功能重叠的双轨典型）
3. 新接线区块发现性——页面底部区块 vs 主工具卡导航（曝光 vs 整洁）

- 11-03 8-2 A批(迁后撤): upload四文件工具迁入t家族——新增/api/t/pdf-text|word-text|xlsx-text(fxUser登录门+AddHist+fxSend文本/CSV直出), img复用既有/api/t/img-process; 撤POST /api/upload/{tool}路由+runToolFile/toolImg/resizeNearest死代码链+registry四条目(工具数27→23)。行为变化: 文件处理由匿名改需登录(t家族安全基线); 错误webhook告警不再覆盖文件工具。e2e: 增upB助手(files字段+blob/文本三元), step07换断言+显式登录兜底; step11 ai/chat容错降级——pristine裸测[证136]实证沉睡失败(10-27 dacda7f llm wire移交internal/ai RequireRole后401×7天, journey后段12-13一并失覆), 降级=200强断言/非200 Logf续行(修复后棘轮自动收紧); tools_test: 快照锁23+fileCases空化+TestFileCoreDirect外提(三直调平移, img归e2e)。verify: 棘轮23+防复活断言(405归一404), 检数50→51。TODO④立项: ①C传统文件工具前端接线三新端点。TODO⑤立项(优先): internal/ai RequireRole与主session脱节——/api/ai/*全族(api.go:37-46十端点)对登录用户401, 前端AI助手可用性存疑, 修复方向=RequireRole对齐主session验证。规约#19附注: 探针三态=调用点/键名/计数常量([证129]盲区实证)。

- 11-03 ⑤关账(翻案+活棘轮兑现): step11 401考古终审——原判"RequireRole×主session脱节"翻案为双轨并轨设计+e2e凭证欠账: 主轨=nova_tok不透明token([证162]cookie HttpOnly/7d, [证169]s.user兼容Bearer喂SessionUser), v2轨=JWT-HS256 Bearer([证153]authenticate=parseToken(bearer), [证163]v2 auth store/secret/login族自成体系, dacda7f刻意并轨); [证173]前端core.js完整v2闭环(login/register→nb_token→authHeaders), ai-studio形状[证157]匹配——产品鉴权无断链。e2e: step11容错降级升强断言(无凭证401|v2注册409转login|me 200|带token chat非401)+doH助手; verify+2检(v2 me+ai-status, 51→53)。探针probe@example.com(e2e data/)与probe-ai@example.com(NOA_DATA_DIR)残留, 与a82probe同批清理。TODO⑥立项: /ai.html聊天页双断链[证177]——请求体{message,history}≠chat契约{messages[]}(恒400), SSE帧解析期望{type,text}而relaySSE透传OpenAI形状{choices[].delta.content}; ai-chat.js修复+前端验证单独立批。

- 11-03 ⑥关账: /ai-chat.html聊天页双断链修(证据[证171/177/179])——请求体{message,history}→{messages:S.msgs.slice(-20)}(含当前user消息, 对齐chat契约messages[]); SSE解析{type,text}/{tokens_in,out}→OpenAI形状{choices[].delta.content}/{usage:{prompt_tokens,completion_tokens}}(relaySSE透传, usage帧未请求恒空留兼容位); 头注阶段5幻影契约→dacda7f实契约(防后人照注释写断链)。记档: [证179]承载页=ai-chat.html(登录跳转next参数佐证), [证182]ai.html另有其用(Nova工具导航, 引app.js, 勿混)。verify+3检(ai-chat页200/chat契约锁空messages=400/形状锁grep=1), 53→56。纯前端批: 零build零重启(8080磁盘serve即生效)。人工冒烟建议: 配profile后实聊一轮(运行时端到端未自动化; 后端契约已由契约锁+M89 journey双锁)。
