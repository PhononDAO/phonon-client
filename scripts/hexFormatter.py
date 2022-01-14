#!/usr/bin/python
import sys
import os

if len(sys.argv) < 2:
  print("please provide hexstring as argument in the following format: 02b4632d08485ff1df2db55b9dafd23347d1c47a457072a1e87be26896549a8737")
  print("it will be converted into this output format: ['0x02', '0xb4', '0x63', '0x2d', '0x08', '0x48', '0x5f', '0xf1', '0xdf', '0x2d', '0xb5', '0x5b', '0x9d', '0xaf', '0xd2', '0x33', '0x47', '0xd1', '0xc4', '0x7a', '0x45', '0x70', '0x72', '0xa1', '0xe8', '0x7b', '0xe2', '0x68', '0x96', '0x54', '0x9a', '0x87', '0x37']")
  sys.exit(1)

hexString = sys.argv[1]

result = []
for i, x in enumerate(hexString):
  if i % 2 == 0:
    result.append("0x"+hexString[i:i+2])

print(("[{0}]".format(', '.join(map(str, result)))))
print("length: ", len(result))