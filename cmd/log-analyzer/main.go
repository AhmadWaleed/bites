package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	level = flag.String("level", "info", "comma separated list of log level to analyze. e.g: 'info,warn,error'")
	start = flag.String("start", "", "start time filter. eg. '2021-01-01T00:00:00'")
	end   = flag.String("end", "", "end time filter. eg. '2021-01-01T23:59:59'")
)

var (
	levels    = make(map[string]struct{}, 4)
	startTime time.Time
	endTime   time.Time
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("log-analyzer: ")
	flag.Usage = Usage
	flag.Parse()

	var file string
	if flag.NArg() > 0 {
		file = flag.Arg(0)
	}
	if file == "" {
		log.Fatalln("arg: file name is required")
	} else if !isLogFile(file) {
		log.Fatalf("arg: %s is not a log file", file)
	}

	for _, l := range strings.Split(*level, ",") {
		levels[l] = struct{}{}
	}

	if *start != "" {
		t, err := time.Parse(time.DateTime, *start)
		if err != nil {
			log.Fatalln("invalid start time: ", err)
		}
		startTime = t
	}
	if *end != "" {
		t, err := time.Parse(time.DateTime, *end)
		if err != nil {
			log.Fatalln("invalid end time: ", err)
		}
		endTime = t
	}

	filter := []FilterFunc{
		func(entry LogEntry) bool {
			_, ok := levels[strings.ToLower(entry.level)]
			return !ok
		},
		func(entry LogEntry) bool {
			if !startTime.IsZero() && entry.time.Before(startTime) {
				return true
			}
			return false
		},
		func(entry LogEntry) bool {
			if !endTime.IsZero() && entry.time.After(endTime) {
				return true
			}
			return false
		},
	}

	f, err := os.OpenFile(file, os.O_RDONLY, 0644)
	if err != nil {
		log.Fatalln("failed to open file: ", err)
	}

	logs := ReadFile(f)
	if len(logs) == 0 {
		log.Fatalln("no log entries found")
	}

	report := Analyze(logs, filter...)
	report.Print()
}

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of log-analyzer:\n")
	fmt.Fprintf(os.Stderr, "\tlog-analyzer [-level] [-start,-end 'DD-MM-YYY HH:MM:SS'] filename ... \n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

type FilterFunc func(LogEntry) bool

// Analyze Analyze logs and return the analysis report.
// Each log entry will be tested against the provided filter.
func Analyze(entries []LogEntry, filter ...FilterFunc) *AnalysisReport {
	report := &AnalysisReport{
		MsgFrequency: make(map[string]int, 10),
	}
	for _, entry := range entries {
		for _, skip := range filter {
			if skip(entry) {
				continue
			}
		}
		report.Add(entry)
	}
	return report
}

// ReadFile read given log file and valid log entries.
// Log entry not following the format will be skipped.
func ReadFile(f *os.File) []LogEntry {
	var entries []LogEntry
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		entry, err := NewLogEntry(line)
		if err != nil {
			log.Println("invalid log entry: ", err)
		}
		entries = append(entries, entry)
	}
	return entries
}

func isLogFile(file string) bool {
	_, ext, _ := strings.Cut(file, ".")
	switch ext {
	case "log", "txt":
		return true
	default:
		return false
	}
}

type AnalysisReport struct {
	TotalEntries int
	Info         int
	Warn         int
	Error        int
	Debug        int
	ResponseTime []float64 // in ms
	MsgFrequency map[string]int
}

const (
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
	LevelDebug = "debug"
)

func (report *AnalysisReport) Add(entry LogEntry) {
	report.TotalEntries++

	// Record the log level count.
	switch strings.ToLower(entry.level) {
	case LevelInfo:
		report.Info++
	case LevelWarn:
		report.Warn++
	case LevelError:
		report.Error++
	case LevelDebug:
		report.Debug++
	}

	// Record the response time.
	if strings.HasSuffix(entry.message, "ms") {
		words := strings.Split(strings.TrimSuffix(entry.message, " ms"), " ")
		respTime := words[len(words)-1]
		if n, err := strconv.ParseFloat(respTime, 64); err == nil {
			report.ResponseTime = append(report.ResponseTime, float64(n))
		}
	}

	// Record the frequency of each message.
	report.MsgFrequency[entry.message]++
}

// Total Log Entries: 5000
// INFO: 3000
// DEBUG: 1200
// WARN: 500
// ERROR: 300
// Average Response Time: 245 ms
func (r AnalysisReport) Print() {
	fmt.Printf("Total Log Entries: %d\n", r.TotalEntries)
	fmt.Printf("INFO: %d\n", r.Info)
	fmt.Printf("DEBUG: %d\n", r.Debug)
	fmt.Printf("WARN: %d\n", r.Warn)
	fmt.Printf("ERROR: %d\n", r.Error)
	if len(r.ResponseTime) > 0 {
		var total float64
		for _, v := range r.ResponseTime {
			total += v
		}
		avg := total / float64(len(r.ResponseTime))
		fmt.Printf("Average Response Time: %.2f ms\n", avg)
	}

	var freqCount []int
	for k := range r.MsgFrequency {
		freqCount = append(freqCount, r.MsgFrequency[k])
	}
	sort.Ints(freqCount)
	var freqMsg string
	for k, v := range r.MsgFrequency {
		if v == freqCount[len(freqCount)-1] {
			freqMsg = k
		}
	}
	fmt.Printf("Most frequent mesage: '%s'\n", freqMsg)
}

func NewLogEntry(line string) (LogEntry, error) {
	logLine := strings.SplitN(line, " ", 4)
	if len(logLine) < 4 {
		return LogEntry{}, fmt.Errorf("invalid log entry")
	}
	logDate, logTime, level, msg := logLine[0], logLine[1], logLine[2], logLine[3]
	t, err := time.Parse(time.DateTime, logDate+" "+logTime)
	if err != nil {
		return LogEntry{}, fmt.Errorf("invalid log time: %w", err)
	}
	return LogEntry{
		time:    t,
		level:   level,
		message: msg,
	}, nil
}

type LogEntry struct {
	time    time.Time
	level   string
	message string
}
