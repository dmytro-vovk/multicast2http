package hls

import (
	"bytes"
	"conf"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var (
	HLSDir, Coder string
	streams       map[string]*exec.Cmd
	streamsList   map[string]string = make(map[string]string)
)

func RunStreams(config conf.UrlConfig) {
	streams = make(map[string]*exec.Cmd, len(config))
	streamsList = make(map[string]string, len(config))
	for url, cfg := range config {
		go simpleStreamRunner(url, cfg)
		streamsList[url] = cfg.Title
	}
}

func StopStreams() {
	for _, cmd := range streams {
		log.Print("Stopping ffmpeg...")
		err := cmd.Process.Kill()
		if err != nil {
			log.Printf("Failed to stop ffmpeg")
		} else {
			log.Printf("Stopped.")
		}
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
	chunkLen := strconv.Itoa(conf.Conf().Hls.ChunkLen)
	if cfg.FfmpegArgs != "" {
		args = cfg.FfmpegArgs
	} else if cfg.CopyStream {
		args = "-i udp://@" + cfg.Source + "?fifo_size=1000000&overrun_nonfatal=1 -y -threads 8 -c:a copy -c:v copy -flags -global_header -map 0 -hls_time " + chunkLen + " -hls_list_size 10 -hls_wrap 12 -start_number 1 stream.m3u8"
	} else {
		if cfg.Deinterlace {
			args = "-y -i udp://@" + cfg.Source + " -vcodec libx264 -crf 23 -preset superfast -profile:v baseline -deinterlace -level 3.0 -g 25 -acodec libmp3lame -flags -global_header -map 0 -f segment -segment_time " + chunkLen + " -segment_list_size 6 -segment_list stream.m3u8 -segment_list_type m3u8 -segment_format mpegts stream%01d.ts"
		} else {
			args = "-y -i udp://@" + cfg.Source + " -vcodec libx264 -crf 23 -preset superfast -profile:v baseline -level 3.0 -g 25 -acodec libmp3lame -flags -global_header -map 0 -f segment -segment_time " + chunkLen + " -segment_list_size 6 -segment_list stream.m3u8 -segment_list_type m3u8 -segment_format mpegts stream%01d.ts"
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
