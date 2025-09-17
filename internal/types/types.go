package types

type PingResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Version   string `json:"version"`
	Timestamp int64  `json:"timestamp"`
}
