package mySocket

// 传送数据
type FormatData struct {
	Size int32 // 先所接受 size 大小 来决定 后续整体包大小
	Id   int32
	Seq  int32
	Body []byte // json、portobuf
}

const (
	max_buffer_size = 65536 // 最大buff长度
	FirstReadSize   = 4     // 首次读取32位 来判断后面部分数据长度
	HeadSize        = 12    // 头部大小   除了body以外 包含size
)
