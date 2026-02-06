package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"testing"
	"time"
)

func createRequestBuf(t byte, num1 int, num2 int) ([]byte, error) {
	// write type byte
	var err error
	buf := append([]byte{}, t)

	if t == byte('I') {
		buf, err = binary.Append(buf, binary.BigEndian, int32(time.Now().Unix()))
		if err != nil {
			return []byte{}, fmt.Errorf("could not encode timestamp")
		}

		buf, err = binary.Append(buf, binary.BigEndian, int32(num2))
		if err != nil {
			return []byte{}, fmt.Errorf("could not encode price")
		}
	} else if t == byte('Q') {
		// append mintime
		buf, err = binary.Append(buf, binary.BigEndian, int32(num1))
		if err != nil {
			return []byte{}, fmt.Errorf("could not encode minTime")
		}

		// append maxtime
		buf, err = binary.Append(buf, binary.BigEndian, int32(num2))
		if err != nil {
			return []byte{}, fmt.Errorf("could not encode maxTime")
		}
	}

	if len(buf) != 9 {
		return []byte{}, fmt.Errorf("wrong buffer size")
	}

	return buf, nil
}

func TestInsert(t *testing.T) {
	t.Run("handle basic insert", func(t *testing.T) {
		client, err := net.Dial("tcp", ":8080")
		if err != nil {
			t.Errorf("could not connect to :8080")
		}

		for n := range 5 {
			r, err := createRequestBuf(byte('I'), int(time.Now().Unix()), n)
			if err != nil {
				t.Error(err)
			}
			_, err = client.Write(r)

			if err != nil {
				t.Error(err)
			}
		}
	})
}

func TestQuery(t *testing.T) {
	t.Run("handle basic query", func(t *testing.T) {
		conn, err := net.Dial("tcp", ":8080")
		if err != nil {
			t.Errorf("could not connect to :8080")
		}

		for n := range 5 {
			r, err := createRequestBuf(byte('I'), int(time.Now().Unix()), n)
			if err != nil {
				t.Error(err)
			}
			time.Sleep(time.Second)
			_, err = conn.Write(r)

			if err != nil {
				t.Error(err)
			}
		}

		insertRequest, err := createRequestBuf(byte('Q'), int(time.Date(2020, time.January, 1, 0, 0, 0, 0, time.Local).Unix()), int(time.Now().Unix()))
		conn.Write(insertRequest)

		time.Sleep(time.Second)

		buf := make([]byte, 4)
		conn.Read(buf)

		fmt.Println(binary.BigEndian.Uint32(buf))
	})
}
