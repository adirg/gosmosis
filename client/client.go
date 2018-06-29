package client

type Client struct {
	host string
	port uint
}

func NewClient(host string, port uint) *Client {
	c := new(Client)

	c.host = host
	c.port = port

	return c
}
