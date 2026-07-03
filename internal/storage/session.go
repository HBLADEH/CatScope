package storage

type Session struct {
	Name      string `json:"name"`
	Device    string `json:"device"`
	CreatedAt string `json:"createdAt"`
}
