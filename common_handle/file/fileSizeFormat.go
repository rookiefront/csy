package file

import "fmt"

// ByteUnit 字节单位类型
type ByteUnit string

const (
	B  ByteUnit = "B"
	KB ByteUnit = "KB"
	MB ByteUnit = "MB"
	GB ByteUnit = "GB"
	TB ByteUnit = "TB"
	PB ByteUnit = "PB"
	EB ByteUnit = "EB"
	ZB ByteUnit = "ZB"
	YB ByteUnit = "YB"
)

// unitBytes 单位对应的字节数映射
var unitBytes = map[ByteUnit]float64{
	B:  1,
	KB: 1 << 10,
	MB: 1 << 20,
	GB: 1 << 30,
	TB: 1 << 40,
	PB: 1 << 50,
	EB: 1 << 60,
	ZB: 1 << 70,
	YB: 1 << 80,
}

// unitOrder 单位顺序，用于智能单位选择
var unitOrder = []ByteUnit{B, KB, MB, GB, TB, PB, EB, ZB, YB}

// FormatByteUnit 将字节数格式化为合适的单位（自动选择最大单位）
func (f *FileHandle) FormatByteUnit(bytes float64) string {
	return f.FormatByteUnitFrom(bytes, B)
}

// FormatByteUnitFrom 将指定起始单位的数值转换为合适的单位
// value: 数值
// fromUnit: 起始单位
func (f *FileHandle) FormatByteUnitFrom(value float64, fromUnit ByteUnit) string {
	// 将值转换为字节
	bytes := value * unitBytes[fromUnit]

	// 自动选择合适的单位
	for i := len(unitOrder) - 1; i >= 0; i-- {
		unit := unitOrder[i]
		if bytes >= unitBytes[unit] {
			if unit == B {
				return fmt.Sprintf("%.0f %s", bytes, unit)
			}
			return fmt.Sprintf("%.2f %s", bytes/unitBytes[unit], unit)
		}
	}

	return fmt.Sprintf("%.0f %s", bytes, B)
}

// FormatByteUnitTo 将字节数转换为指定单位
// bytes: 字节数
// toUnit: 目标单位
func (file *FileHandle) FormatByteUnitTo(bytes float64, toUnit ByteUnit) string {
	if toUnit == B {
		return fmt.Sprintf("%.0f %s", bytes, toUnit)
	}
	return fmt.Sprintf("%.2f %s", bytes/unitBytes[toUnit], toUnit)
}

// FormatByteUnitFromTo 将指定起始单位的数值转换为指定目标单位
// value: 数值
// fromUnit: 起始单位
// toUnit: 目标单位
func (file *FileHandle) FormatByteUnitFromTo(value float64, fromUnit, toUnit ByteUnit) string {
	// 先转换为字节
	bytes := value * unitBytes[fromUnit]
	// 再转换为目标单位
	return file.FormatByteUnitTo(bytes, toUnit)
}

// FormatByteUnitSmart 智能格式化，可指定最大单位
// bytes: 字节数
// maxUnit: 最大允许使用的单位（nil表示无限制）
func (file *FileHandle) FormatByteUnitSmart(bytes float64, maxUnit *ByteUnit) string {
	var maxIndex int = len(unitOrder) - 1

	if maxUnit != nil {
		// 找到最大单位对应的索引
		for i, u := range unitOrder {
			if u == *maxUnit {
				maxIndex = i
				break
			}
		}
	}

	// 从最大允许单位向下查找
	for i := maxIndex; i >= 0; i-- {
		unit := unitOrder[i]
		if bytes >= unitBytes[unit] {
			if unit == B {
				return fmt.Sprintf("%.0f %s", bytes, unit)
			}
			return fmt.Sprintf("%.2f %s", bytes/unitBytes[unit], unit)
		}
	}

	return fmt.Sprintf("%.0f %s", bytes, B)
}

/*
fh := &file.FileHandle{}

// 1. 基本用法（自动选择单位）
fmt.Println(fh.FormatByteUnit(1536))           // "1.50 KB"
fmt.Println(fh.FormatByteUnit(1048576))        // "1.00 MB"

// 2. 从指定起始单位转换
fmt.Println(fh.FormatByteUnitFrom(2.5, file.MB)) // 将2.5MB转换为合适单位
fmt.Println(fh.FormatByteUnitFrom(1024, file.KB)) // "1.00 MB"

// 3. 转换为指定目标单位
fmt.Println(fh.FormatByteUnitTo(3145728, file.MB))   // "3.00 MB"
fmt.Println(fh.FormatByteUnitTo(1073741824, file.GB)) // "1.00 GB"

// 4. 从指定起始单位转换到指定目标单位
fmt.Println(fh.FormatByteUnitFromTo(2.5, file.GB, file.MB)) // "2560.00 MB"
fmt.Println(fh.FormatByteUnitFromTo(1000, file.KB, file.MB)) // "0.98 MB"

// 5. 智能格式化（限制最大单位）
maxUnit := file.MB
fmt.Println(fh.FormatByteUnitSmart(1500000000, &maxUnit)) // 限制最大MB级别

// 6. 不限制最大单位
fmt.Println(fh.FormatByteUnitSmart(1500000000, nil)) // 自动选择GB或TB
*/
