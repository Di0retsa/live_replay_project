package utils

import (
	"crypto/md5"
	"encoding/hex"
)

// MD5V 对目标字符串取Hash salt：加盐字段，iteration：hash迭代轮数
func MD5V(str string, salt string, iteration int) string {
	b := []byte(str)
	s := []byte(salt)
	h := md5.New()
	h.Write(s) // 要先传盐值
	h.Write(b)
	res := h.Sum(nil)
	for i := 0; i < iteration-1; i++ {
		h.Reset()
		h.Write(res)
		res = h.Sum(nil)
	}
	return hex.EncodeToString(res)
}
