package controllers

type errorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

type ocrResponse struct {
	Success bool   `json:"success"`
	Result  string `json:"result"`
	Version string `json:"version"`
}
