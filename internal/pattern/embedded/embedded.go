package embedded

import "embed"

//go:embed *.rle
var Embedded embed.FS
