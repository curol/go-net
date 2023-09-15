package message

import "net"

type ResponseMessage struct {
	*RequestMessage
}

func NewResponseMessageFromConnection(conn net.Conn) (*ResponseMessage, error) {
	req, err := NewRequestMessageFromConnection(conn)
	if err != nil {
		return nil, err
	}
	return &ResponseMessage{req}, nil
}
