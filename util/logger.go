package util

import (
	"log/slog"
	"os"
	"path/filepath"
)

// LogPath returns the absolute path of the application log file.
func LogPath(cfgDir string) string {
	return filepath.Join(cfgDir, "dnc", "dnc.log")
}

// InitLogger opens (or creates) the log file and sets it as the slog default.
// Rotation: if the existing file is >= maxBytes it is renamed to dnc.log.1
// before the new file is opened. Pass 0 to disable rotation.
// Returns a cleanup func that closes the file.
func InitLogger(cfgDir string, maxBytes int64) (func(), error) {
	logPath := LogPath(cfgDir)
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return func() {}, err
	}
	if maxBytes > 0 {
		if fi, err := os.Stat(logPath); err == nil && fi.Size() >= maxBytes {
			_ = os.Rename(logPath, logPath+".1")
		}
	}
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return func() {}, err
	}
	h := slog.NewTextHandler(f, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(h))
	return func() { _ = f.Close() }, nil
}
