package launcher

type logDownload Download

type FileDownload struct {
	logDownload
	ID string `json:"id"`
}

type LoggingConfiguration struct {
	Argument string       `json:"argument"`
	File     FileDownload `json:"file"`
	Type     string       `json:"type"`
}
