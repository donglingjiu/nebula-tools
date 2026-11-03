#!/usr/bin/env python3
# vhash.py — web/*.html 内本地 assets/*.js|css 引用统一 ?v=<sha256[:8]>
# 用法: vhash.py [bump|check]  bump=重写(有缺失资产则预扫中止不写盘) check=零漂移退出0
import hashlib, os, re, sys
ROOT = os.path.dirname(os.path.abspath(__file__)); WEB = os.path.join(ROOT, 'web')
PAT = re.compile(r'((?:src|href)=")((?:/)?assets/[^"]+?\.(?:js|css))(\?v=[^"]*)?(")')
def scan_apply(mode):
    missing, drift, bumped = [], [], 0
    for name in sorted(os.listdir(WEB)):
        if not name.endswith('.html'): continue
        fp = os.path.join(WEB, name); s = open(fp, encoding='utf-8').read(); n = [0]
        def sub(m):
            head, ref, q, tail = m.groups()
            ap = os.path.join(WEB, ref.lstrip('/'))
            if not os.path.isfile(ap):
                missing.append(f'{name}: {ref}'); return m.group(0)
            want = '?v=' + hashlib.sha256(open(ap,'rb').read()).hexdigest()[:8]
            if q == want: return m.group(0)
            drift.append(f'{name}: {ref}{q or "[无版本]"}'); n[0] += 1
            return head + ref + want + tail if mode=='bump' else m.group(0)
        ns = PAT.sub(sub, s)
        if mode=='bump' and n[0] and ns != s:
            open(fp,'w',encoding='utf-8').write(ns); bumped += 1
    return missing, drift, bumped
mode = sys.argv[1] if len(sys.argv)>1 else 'bump'
missing, drift, _ = scan_apply('scan')
if missing:
    for x in missing: print('MISSING', x)
    print(f'bump中止: {len(missing)}个缺失资产引用, 未写盘'); sys.exit(2)
if mode=='bump':
    _, _, bumped = scan_apply('bump')
    print(f'bump: files={bumped} drift_resolved={len(drift)}')
else:
    for x in drift: print('DRIFT', x)
    print(f'check: drift={len(drift)} missing=0'); sys.exit(1 if drift else 0)
