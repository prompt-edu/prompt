package keyDTO

type SetKeyRequest struct {
	Key string `json:"key" binding:"required,max=512"`
}
