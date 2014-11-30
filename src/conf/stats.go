package conf

type StatsType struct {
	RunningStreams uint
}

var (
	Stats        StatsType
	StatsChannel chan bool = make(chan bool, 10)
)

func init() {
	go statsCollector()
}

// Stats listener
func statsCollector() {
	for {
		if s := <-StatsChannel; s {
			Stats.RunningStreams++
		} else if Stats.RunningStreams > 0 { // Just to prevent uint underflow
			Stats.RunningStreams--
		}
	}
}
