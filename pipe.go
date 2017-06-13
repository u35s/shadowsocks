package shadowsocks

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"time"
)

func SetReadTimeout(c net.Conn) {
	if readTimeout != 0 {
		c.SetReadDeadline(time.Now().Add(readTimeout))
	}
}

// PipeThenClose copies data from src to dst, closes dst when done.
func PipeThenClose(src, dst net.Conn) (flow int) {
	bts := leakyBuf.GetBts()
	defer leakyBuf.PutBts(bts)
	defer dst.Close()
	for {
		SetReadTimeout(src)
		if n, err := src.Read(bts); err != nil {
			break
		} else if n > 0 {
			if _, err := dst.Write(bts[0:n]); err != nil {
				break
			} else {
				flow += n
			}
		}
	}
	return
}

// PipeThenClose copies data from src to dst, closes dst when done, with ota verification.
func PipeThenCloseOta(src *Conn, dst net.Conn) {
	const (
		dataLenLen  = 2
		hmacSha1Len = 10
		idxData0    = dataLenLen + hmacSha1Len
	)

	defer func() {
		dst.Close()
	}()
	// sometimes it have to fill large block
	buf := leakyBuf.GetBts()
	defer leakyBuf.PutBts(buf)
	for i := 1; ; i += 1 {
		SetReadTimeout(src)
		if n, err := io.ReadFull(src, buf[:dataLenLen+hmacSha1Len]); err != nil {
			if err == io.EOF {
				break
			}
			Debug.Printf("conn=%p #%v read header error n=%v: %v", src, i, n, err)
			break
		}
		dataLen := binary.BigEndian.Uint16(buf[:dataLenLen])
		expectedHmacSha1 := buf[dataLenLen:idxData0]

		var dataBuf []byte
		if len(buf) < int(idxData0+dataLen) {
			dataBuf = make([]byte, dataLen)
		} else {
			dataBuf = buf[idxData0 : idxData0+dataLen]
		}
		if n, err := io.ReadFull(src, dataBuf); err != nil {
			if err == io.EOF {
				break
			}
			Debug.Printf("conn=%p #%v read data error n=%v: %v", src, i, n, err)
			break
		}
		chunkIdBytes := make([]byte, 4)
		chunkId := src.GetAndIncrChunkId()
		binary.BigEndian.PutUint32(chunkIdBytes, chunkId)
		actualHmacSha1 := HmacSha1(append(src.GetIv(), chunkIdBytes...), dataBuf)
		if !bytes.Equal(expectedHmacSha1, actualHmacSha1) {
			Debug.Printf("conn=%p #%v read data hmac-sha1 mismatch, iv=%v chunkId=%v src=%v dst=%v len=%v expeced=%v actual=%v", src, i, src.GetIv(), chunkId, src.RemoteAddr(), dst.RemoteAddr(), dataLen, expectedHmacSha1, actualHmacSha1)
			break
		}
		if n, err := dst.Write(dataBuf); err != nil {
			Debug.Printf("conn=%p #%v write data error n=%v: %v", dst, i, n, err)
			break
		}
	}
}
