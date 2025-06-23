package readeckapi

type ReadeckConn struct {
	endpoint    string
	bearerToken string
}

func NewConn(endpoint string,
	bearerToken string) ReadeckConn {
	return ReadeckConn{
		endpoint:    endpoint,
		bearerToken: bearerToken,
	}
}
