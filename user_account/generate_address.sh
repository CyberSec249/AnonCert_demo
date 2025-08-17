#!/bin/bash

x=2

# 生成私钥
openssl ecparam -name secp256k1 -genkey -noout -out "ca${x}_sk.pem"

# 生成公钥
openssl ec -in "ca${x}_sk.pem" -text -noout 2>/dev/null | sed -n '7,11p' | tr -d ": \n" | awk '{print substr($0,3)}' > "user${x}_pk.pem"

# 生成地址
openssl ec -in "user${x}_sk.pem" -text -noout 2>/dev/null | sed -n '7,11p' | tr -d ": \n" | awk '{print substr($0,3)}' | ./keccak-256sum -x -l | tr -d ' -' | tail -c 41 > "user${x}_addr"

echo "生成成功！"
echo "私钥文件：user${x}_sk.pem"
echo "公钥文件：user${x}_pk.pem"
echo "地址文件：user${x}_addr"
