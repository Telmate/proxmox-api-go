package sizeunit

import (
	"fmt"
	"strconv"
)

type SizeUnit int64

const (
	KB SizeUnit = 1 << (10 * (iota + 1))
	MB
	GB
)

var shortUnitMap = map[SizeUnit]string{
	KB: "K",
	MB: "M",
	GB: "G",
}

var longUnitMap = map[SizeUnit]string{
	KB: "kilobyte",
	MB: "megabyte",
	GB: "gigabyte",
}

func FormatToShortString(size int, sizeUnit SizeUnit) string {
	return strconv.Itoa(size) + shortUnitMap[sizeUnit]
}

func FormatToLongString(size int, sizeUnit SizeUnit) string {
	return fmt.Sprintf("%s %s", strconv.Itoa(size), longUnitMap[sizeUnit])
}

func ConvertTo(size int, oldSizeUnit SizeUnit, newSizeUnit SizeUnit) (newSize int, newUnit SizeUnit) {
	return size * int(oldSizeUnit) / int(newSizeUnit), newSizeUnit
}
