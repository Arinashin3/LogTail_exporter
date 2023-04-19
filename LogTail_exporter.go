package main

import (
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	promVersion "github.com/prometheus/common/version"
  common "github.com/Arinashin3/LogTail_exporter"
	"log"
	_ "net/http/pprof"
	"os"
	"regexp"
	"strings"
)

var version string

func printManual() {
	fmt.Print(`Usage: LogTail_exporter [options] --config config.yml

` + "\n")
}

func init() {
	promVersion.Version = "v0.1.0"
	prometheus.MustRegister(promVersion.NewCollector("LogTail_exporter"))
}

type (
	prefixRegex struct {
		prefix string
		regex  *regexp.Regexp
	}
	nameMapperRegex struct {
		mapping map[string]*prefixRegex
	}
)

func (nmr *nameMapperRegex) String() string {
	return fmt.Sprintf("%+v", nmr.mapping)
}

func parseNameMapper(s string) (*nameMapperRegex, error) {
	mapper := make(map[string]*prefixRegex)
	if s == "" {
		return &nameMapperRegex{mapper}, nil
	}

	toks := strings.Split(s, ",")
	if len(toks)%2 == 1 {
		return nil, fmt.Errorf("bad namemapper: odd number of tokens")
	}

	for i, tok := range toks {
		if tok == "" {
			return nil, fmt.Errorf("bad namemapper: token %d is empty", i)
		}
		if i%2 == 1 {
			name, regexstr := toks[i-1], tok
			matchName := name
			prefix := name + ":"

			if r, err := regexp.Compile(regexstr); err != nil {
				return nil, fmt.Errorf("error compiling regexp '%s': %v", regexstr, err)
			} else {
				mapper[matchName] = &prefixRegex{prefix: prefix, regex: r}
			}
		}
	}

	return &nameMapperRegex{mapper}, nil
}

func main() {
	var (
		listenAddress     = flag.String("web.listen-address", ":9003", "Address on which to expose metrics and web interface.")
		metricsPath       = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics")
		onceToSTdoutDelay = flag.Duration("once-to-stdout-delay", 0, "Don't bindm just wait this much time, print the metrics once to stdout, and exit")
		logfilePath       = flag.String("logfile.path", "", "path to read log data from")
		logfilePattern    = flag.String("logfile.pattern", "", "Log pattern to filter")
		configPath        = flag.String("config", "", "path to YAML config file")
		debug             = flag.Bool("debug", false, "log debugging information and exit")
		showVersion       = flag.Bool("version", false, "print version information and exit")
	)
	flag.Parse()
	if *showVersion {
		fmt.Printf("%s\n", promVersion.Print("LogTail_exporter"))
		os.Exit(0)
	}

	var matchnamer common.MatchNamer

	if *configPath != "" {
		if *nameMapping != "" {
			log.Fatalf("--config cannot be used")
		}

		cfg, err := config.ReadFile(*configPath, *debug)
		if err != nil {
			log.Fatalf("error reading config file %q: %v", *configPath, err)
		}
		log.Printf("Reading metrics from %s based on %q", *procfsPath, *configPath)
		matchnamer = cfg.MatchNamers
		if *debug {
			log.Printf("using config matchnamer: %v", cfg.MatchNamers)
		}
	} else {
		namemapper, err := parseNameMapper(*nameMapping)
		if err != nil {
			log.Fatalf("Error parsing -namemapping argument '%s': %v", *nameMapping, err)
		}

		var names []string
		for _, s := range strings.Split(*procNames, ",") {
			if s != "" {
				if _, ok := namemapper.mapping[s]; !ok {
					namemapper.mapping[s] = nil
				}
				names = append(names, s)
			}
		}

		log.Printf("Reading metrics from %s for procnames: %v", *procfsPath, names)
		if *debug {
			log.Printf("using cmdline matchnamer: %v", namemapper)
		}
		matchnamer = namemapper
	}
	fmt.Println(listenAddress, metricsPath, onceToSTdoutDelay, logfilePath, logfilePattern, configPath, debug, showVersion)
}
