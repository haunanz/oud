package oud


// ReadEventCallback 读事件回调函数类型 多一个时间参数
type ReadEventCallback func(time int)
// EventCallback 事件的处理函数类型
type EventCallback func()
// TimeCallback 事件回调函数别名
type TimeCallback func() 
// ConnectionCallback 连接回调函数别名
type ConnectionCallback func(*TCPConnection)
// CloseCallback 关闭连接的回调函数别名
type CloseCallback func(*TCPConnection)
// WriteCompleteCallback ?
type WriteCompleteCallback func(*TCPConnection)
// HighWaterMarkCallback 高水位回调函数别名
type HighWaterMarkCallback func(*TCPConnection,int)
// MessageCallback 消息处理函数（读事件）回调函数别名
type MessageCallback func(*TCPConnection,*Buffer,int)

