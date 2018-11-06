package mySocket

import (
	"bufio"
	"errors"
	"io"
	"net"
	"runtime/debug"
	"sync/atomic"
	"time"

	. "github.com/soekchl/myUtils"
)

var (
	SessionClosedError = errors.New("Session Closed.")
	SendDataIsNilError = errors.New("Send Data Is Nil.")
	ReadError          = errors.New("Read Error.")
	ReadOverflow       = errors.New("tcp_conn: read buffer overflow")
)

// 会话
type Session struct {
	ByteSendChan chan []byte //	数据发送
	ByteRecvChan chan []byte // 数据接收
	closeFlag    int32
	CloseChan    chan bool
	conn         net.Conn
	firstBuff    []byte
	lastUseTime  int64 // 最后使用会话时间戳
}

func newSession(conn *net.Conn, chanSize int) *Session {
	session := &Session{
		CloseChan: make(chan bool, 1),
		conn:      *conn,
		firstBuff: make([]byte, FirstReadSize),
	}

	if chanSize <= 0 {
		chanSize = 5
	}
	session.ByteRecvChan = make(chan []byte, chanSize*2)
	session.ByteSendChan = make(chan []byte, chanSize)
	go session.readLoop()
	go session.sendByteLoop()

	return session
}

func (session *Session) Send(data *FormatData) error {
	if data == nil {
		return SendDataIsNilError
	}
	if session.IsClosed() {
		return SessionClosedError
	}

	buff := make([]byte, len(data.Body)+HeadSize)
	copy(buff[HeadSize:], data.Body)

	data.Size = int32(len(data.Body) + HeadSize - FirstReadSize)
	EncodeUint32(uint32(data.Size), buff)
	EncodeUint32(uint32(data.Id), buff[4:])
	EncodeUint32(uint32(data.Seq), buff[8:])

	select {
	case session.ByteSendChan <- buff:
	case <-session.CloseChan:
		session.Close()
		return SessionClosedError
	}
	return nil
}

func (session *Session) sendByteLoop() {
	defer session.Close()
	var buff []byte
	var ok bool
	var err error
	for {
		select {
		case buff, ok = <-session.ByteSendChan:
			if !ok {
				return
			}
			session.lastUseTime = time.Now().Unix()
			_, err = session.conn.Write(buff)
			if err != nil {
				Error("[sendByteLoop] ", err)
				return
			}
		case <-session.CloseChan:
			return
		}
	}
}

func (session *Session) Receive() (*FormatData, error) {
	var buff []byte
	var ok bool
	select {
	case buff, ok = <-session.ByteRecvChan:
		if !ok {
			session.Close()
			return nil, ReadError
		}
	case <-session.CloseChan:
		session.Close()
		return nil, SessionClosedError
	}

	// buff 解析传送
	return &FormatData{
		Id:   int32(DecodeUint32(buff[0:])),
		Seq:  int32(DecodeUint32(buff[4:])),
		Body: buff[8:],
	}, nil
}

func (session *Session) readLoop() {
	Debug("[readLoop]")
	defer ErrorShow()
	defer session.Close()
	rbuf := bufio.NewReader(session.conn)
	for {
		//		session.conn.SetReadDeadline(time.Now().Add(session.readTimeOut))
		buf, err := session.readByte(rbuf)
		if err != nil {
			Error("[conn] read error: ", err)
			return
		}
		session.lastUseTime = time.Now().Unix()
		session.ByteRecvChan <- buf // 接收的数据发送处理
	}
}

func (session *Session) readByte(r io.Reader) ([]byte, error) {
	_, err := io.ReadFull(r, session.firstBuff)
	if err != nil {
		return nil, err
	}
	//    [x][x][x][x][x][x][x][x]...
	//    |  (int32) || (binary)
	//    |  4-byte  || N-byte
	//    ------------------------...
	//        size       data			size就是data长度
	msgSize := DecodeUint32(session.firstBuff)
	if msgSize < FirstReadSize ||
		msgSize > uint32(max_buffer_size) {
		Error("[conn] pack length error: ", session.RemoteAddr(), " len:", msgSize, " buff=", session.firstBuff)
		return nil, ReadOverflow
	}

	buf := make([]byte, msgSize)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		Error("[conn] io read data error: ", session.RemoteAddr(), err)
		return nil, err
	}
	return buf, nil
}

func (session *Session) IsClosed() bool {
	return atomic.LoadInt32(&session.closeFlag) == 1
}

func (session *Session) RemoteAddr() string {
	return session.conn.RemoteAddr().String()
}

func (session *Session) Close() error {
	if atomic.CompareAndSwapInt32(&session.closeFlag, 0, 1) {
		Debug("Session Closed!!!    ->", session)
		session.conn.Close()
		close(session.CloseChan)
		close(session.ByteRecvChan)
		close(session.ByteSendChan)
		return nil
	}
	return SessionClosedError
}

func (session *Session) GetLastSessionUseTimeStamp() int64 {
	return session.lastUseTime
}

func ErrorShow() {
	err := recover()
	if err == nil {
		return
	}
	Error("Err:", err, " Debug:", string(debug.Stack()))
}
