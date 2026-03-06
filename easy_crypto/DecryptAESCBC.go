package easy_crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

func DecryptAES_CBC(base64Str, publicKey, privateKey string) ([]byte, error) {
	// publicKey,privateKey 密钥（必须是16/24/32字节）
	// Base64 解码
	encryptedData, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return nil, fmt.Errorf("base64 decode error: %v", err)
	}

	// 创建 AES 密码块
	block, err := aes.NewCipher([]byte(publicKey))
	if err != nil {
		return nil, fmt.Errorf("aes new cipher error: %v", err)
	}

	// 检查数据长度
	if len(encryptedData) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// CBC 模式解密
	iv := []byte(privateKey)
	mode := cipher.NewCBCDecrypter(block, iv)

	// 解密数据（原地解密）
	decrypted := make([]byte, len(encryptedData))
	mode.CryptBlocks(decrypted, encryptedData)

	// 去除 ZeroPadding
	decrypted = removeZeroPadding(decrypted)

	//// 解析 JSON
	//var result map[string]interface{}
	//if err := json.Unmarshal(decrypted, &result); err != nil {
	//	return nil, fmt.Errorf("json unmarshal error: %v", err)
	//}
	//fmt.Println("1")
	return decrypted, nil
}
func removeZeroPadding(data []byte) []byte {
	// 找到最后一个非零字节的位置
	length := len(data)
	for length > 0 && data[length-1] == 0 {
		length--
	}
	return data[:length]
}

// 如果需要更严格的 ZeroPadding 处理，可以使用以下版本：
func removeZeroPaddingStrict(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return data, nil
	}

	// 找到填充的零字节数量
	padding := 0
	for i := length - 1; i >= 0; i-- {
		if data[i] == 0 {
			padding++
		} else {
			break
		}
	}

	// 验证填充是否有效
	if padding == 0 {
		return data, nil
	}

	// 返回去除填充后的数据
	return data[:length-padding], nil
}

// EncryptAES_CBC 使用AES-CBC算法加密数据并返回Base64字符串
func EncryptAES_CBC(data []byte, publicKey, privateKey string) (string, error) {
	// 创建 AES 密码块
	block, err := aes.NewCipher([]byte(publicKey))
	if err != nil {
		return "", fmt.Errorf("aes new cipher error: %v", err)
	}

	// 添加 ZeroPadding
	paddedData := addZeroPadding(data, aes.BlockSize)

	// CBC 模式加密
	iv := []byte(privateKey)
	mode := cipher.NewCBCEncrypter(block, iv)

	// 加密数据
	encrypted := make([]byte, len(paddedData))
	mode.CryptBlocks(encrypted, paddedData)

	// Base64 编码
	base64Str := base64.StdEncoding.EncodeToString(encrypted)

	return base64Str, nil
}

// addZeroPadding 添加 ZeroPadding
func addZeroPadding(data []byte, blockSize int) []byte {
	// 计算需要填充的长度
	padding := blockSize - (len(data) % blockSize)
	if padding == blockSize {
		// 如果数据长度正好是块大小的整数倍，需要添加一个完整的块
		padding = blockSize
	}

	// 创建新的切片，长度为原数据长度 + 填充长度
	padded := make([]byte, len(data)+padding)

	// 复制原数据
	copy(padded, data)

	// 后面的字节默认已经是0，不需要额外操作

	return padded
}
