package hls

import (
	"errors"
	"os"
)

// Check if HLS can be used
func SetupHLS(hlsDir, coder string) error {
	// Supplied dir should not be empty
	if hlsDir == "" {
		return errors.New("No HLS directory specified")
	}
	// Dir must exist
	d, err := os.Stat(hlsDir)
	if os.IsNotExist(err) {
		// Try to create it
		if err = os.MkdirAll(hlsDir, 0777); err != nil {
			return errors.New("Directory for HLS does not exist and unable to create")
		}
		d, err = os.Stat(hlsDir)
	} else if err != nil {
		return err
	}
	if !d.IsDir() {
		return errors.New("HLS path is file, not directory")
	}
	_, err = os.Stat(coder)
	if os.IsNotExist(err) {
		return errors.New("ffmpeg not found")
	}
	HLSDir = hlsDir
	Coder = coder
	return nil
}
