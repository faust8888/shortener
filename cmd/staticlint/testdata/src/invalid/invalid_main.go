package invalid

import "os"

func main() {
	os.Exit(0) // want "direct call to os.Exit is not allowed in main.main"
}
