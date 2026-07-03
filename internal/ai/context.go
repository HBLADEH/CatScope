package ai

type ContextRequest struct {
	EntryID int64 `json:"entryId"`
	Before  int   `json:"before"`
	After   int   `json:"after"`
}
