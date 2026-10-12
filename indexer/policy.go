package indexer

func Authorize(caller, method, path string) bool {
	return (path == "/readyz" || path == "/metrics") && (caller == "indexer" || caller == "control")
}
