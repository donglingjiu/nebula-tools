#!/usr/bin/env python3
# js_gate.py v2 <file.js> - mini JS lexer with regex-literal state.
# bad: raw newline inside quote-strings; raw newline inside regex literals;
# unterminated quote-string at EOF. Templates span lines; dollar-brace
# re-enters code via ctx stack. Comments skipped. Regex vs division by
# previous significant token: value-like -> division; op/keyword -> regex.
import sys
KW=('return','typeof','case','in','of','new','delete','void','instanceof','do','else','throw')

def scan(v):
    Q=chr(39);D=chr(34);B=chr(96);S=chr(92)
    NL=chr(10);DS=chr(36);LB=chr(123);RB=chr(125);SL=chr(47)
    i=0;n=len(v);ctx=[];st=None;prev='op';word=''
    while i<n:
        c=v[i]
        if st==Q or st==D:
            if c==S:
                i+=2;continue
            if c==NL:
                return ('raw-newline-in-string',i)
            if c==st:
                st=None;prev='val'
        elif st==B:
            if c==S:
                i+=2;continue
            if c==DS and i+1<n and v[i+1]==LB:
                ctx.append(1);st=None;prev='op';i+=1
            elif c==B:
                st=None;prev='val'
        elif st=='re':
            if c==S:
                i+=2;continue
            if c==NL:
                return ('raw-newline-in-regex',i)
            if c==SL:
                st=None;prev='val'
        else:
            if c==Q or c==D or c==B:
                word=''
                st=c
            elif c==S:
                word=''
                i+=2;continue
            elif c==SL and i+1<n and v[i+1]==SL:
                word=''
                j=v.find(NL,i)
                i=(n-1) if j<0 else j
            elif c==SL and i+1<n and v[i+1]==chr(42):
                word=''
                j=v.find('*/',i+2)
                i=(n-1) if j<0 else (j+1)
                prev='op'
            elif c==SL:
                if word in KW:
                    prev='op'
                word=''
                if prev=='val':
                    prev='op'
                else:
                    st='re'
            elif c==LB:
                word=''
                if ctx:
                    ctx[-1]+=1
                prev='op'
            elif c==RB:
                word=''
                if ctx:
                    ctx[-1]-=1
                    if ctx[-1]==0:
                        ctx.pop();st=B
                prev='op'
            elif c==')' or c==']':
                word=''
                prev='val'
            elif c.isalnum() or c=='_' or c=='$':
                word=word+c
                prev='val'
            elif c==' ' or c==chr(9) or c==NL or c==chr(13):
                if word in KW:
                    prev='op'
                word=''
            else:
                if word in KW:
                    prev='op'
                word=''
                prev='op'
        i+=1
    if st==Q or st==D:
        return ('unterminated-string-at-eof',n-1)
    return None

def lc(v,i):
    return v.count(chr(10),0,i)+1, i-(v.rfind(chr(10),0,i)+1)+1

if __name__=='__main__':
    v=open(sys.argv[1],encoding='utf-8').read()
    r=scan(v)
    if r is None:
        print('ok')
    else:
        l,c2=lc(v,r[1])
        sys.stderr.write('%s at line %d col %d%s'%(r[0],l,c2,chr(10)))
        print('bad')
