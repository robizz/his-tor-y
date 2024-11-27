package conf

// Config structs
type ExitNode struct {
	// DownloadURLTemplate contains the template for the exit node compressed files URL.
	// The string is supposed to be:
	// https://collector.torproject.org/archive/exit-lists/exit-list-2024-01.tar.xz
	DownloadURLTemplate string
}

type Config struct {
	ExitNode ExitNode
}
