package hlstream

import (
	"bytes"
	"conf"
	"log"
	"os"
	"os/exec"
	"strings"
)

var (
	HLSDir, Coder string
	streams       map[string]*exec.Cmd
)

func RunStreams(config conf.UrlConfig) {
	streams = make(map[string]*exec.Cmd, len(config))
	for url, cfg := range config {
		go simpleStreamRunner(url, cfg)
	}
}

func StopStreams() {
	for _, cmd := range streams {
		cmd.Process.Kill()
	}
}

func simpleStreamRunner(url string, cfg conf.Url) {
	for {
		simpleStreamer(url, cfg)
	}
}

// Run ffmpeg that reads UDP by itself
func simpleStreamer(url string, cfg conf.Url) {
	var args string
	if cfg.FfmpegArgs != "" {
		args = cfg.FfmpegArgs
	} else if cfg.CopyStream {
		args = "-i udp://@" + cfg.Source + "?fifo_size=1000000&overrun_nonfatal=1 -y -threads 8 -c:a copy -c:v copy -flags -global_header -map 0 -hls_time 10 -hls_list_size 10 -hls_wrap 12 -start_number 1 stream.m3u8"
	} else {
		if cfg.Deinterlace {
			args = "-i udp://@" + cfg.Source + "?fifo_size=1000000&overrun_nonfatal=1 -y -threads 8 -c:a aac -ac 2 -strict -2 -c:v libx264 -vprofile baseline -x264opts level=41 -vf \"yadif=0:-1:0\" -flags -global_header -map 0 -hls_time 10 -hls_list_size 10 -hls_wrap 12 -start_number 1 stream.m3u8"
		} else {
			args = "-i udp://@" + cfg.Source + "?fifo_size=1000000&overrun_nonfatal=1 -y -threads 8 -c:a aac -ac 2 -strict -2 -c:v libx264 -vprofile baseline -x264opts level=41 -flags -global_header -map 0 -hls_time 10 -hls_list_size 10 -hls_wrap 12 -start_number 1 stream.m3u8"
		}
	}
	streams[url] = exec.Command(Coder, strings.Split(args, " ")...)
	streams[url].Dir = getDir(url)
	var errOut bytes.Buffer
	streams[url].Stderr = &errOut
	err := streams[url].Start()
	if err != nil {
		log.Printf("Ffmpeg startup error: %s:\n%s", err, errOut.String())
	}
	log.Printf("Ffmpeg started on source %s", cfg.Source)
	err = streams[url].Wait()
	delete(streams, url)
	if err != nil {
		log.Printf("Ffmpeg stoped with error: %s:\n%s", err, errOut.String())
	}
	log.Printf("Ffmpeg exited (source %s)", cfg.Source)
}

// Prepare full path for give URL and make sure dir exists
func getDir(url string) string {
	destinationDir := HLSDir + url
	if _, err := os.Stat(destinationDir); os.IsNotExist(err) {
		if err = os.MkdirAll(destinationDir, 0755); err != nil {
			log.Fatalf("Could not create stream directory %s", destinationDir)
		}
	}
	return destinationDir
}
