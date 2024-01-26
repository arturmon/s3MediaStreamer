package gin

import "time"

const maxUploadSize = 15 << 20 // 5 megabytes

const shutdownTimeout = 5 * time.Second
const ReadHeaderTimeout = 5 * time.Second

func getMusicTypes() map[string]interface{} {
	return map[string]interface{}{
		"audio/mpeg":               nil,
		"audio/flac":               nil,
		"application/octet-stream": nil,
	}
}
