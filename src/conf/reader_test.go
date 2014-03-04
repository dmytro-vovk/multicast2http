package conf

import "testing"

var (
	nonExisting string = "nonexisting.json"
	broken string      = "test/broken.json"
	config1 string     = "test/test-sources-1.json"
)

func TestReadUrls(t *testing.T) {
	_, err := ReadUrls(&nonExisting)
	if err == nil {
		t.Error("Failed to report non-existing config file")
	}
	_, err = ReadUrls(&broken)
	if err == nil {
		t.Error("Failed to detect broken JSON")
	}
	s1, err := ReadUrls(&config1)
	if err != nil {
		t.Error("Failed to read valid config file")
	}
	if len(s1) != 4 {
		t.Error("Sources count mismatch")
	}
}
