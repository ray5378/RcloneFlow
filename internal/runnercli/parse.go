package runnercli

import (
	"fmt"
	"regexp"
	"strings"
)

var statsRe = regexp.MustCompile(`INFO\s*:\s*([\d.]+\s*[KMGTP]?i?B?)\s*/\s*([\d.]+\s*[KMGTP]?i?B?),\s*([\d.]+)%,\s*([\d.]+\s*[KMGTP]?i?B?)/s`)
var fileLineRe = regexp.MustCompile(`(?i)INFO\s*:\s*([^:]+):\s*(\d+(?:\.\d+)?)\s*([KMGTPE]?i?)B\s*/\s*(\d+(?:\.\d+)?)\s*([KMGTPE]?i?)B,\s*(\d+(?:\.\d+)?)%?,\s*(\d+(?:\.\d+)?)\s*([KMGTPE]?i?)B/s`)
var fileCopiedRe = regexp.MustCompile(`(?i)INFO\s*:\s*([^:]+):\s*Copied\s*\(new\)`)
var fileCASMatchedRe = regexp.MustCompile(`(?i)(?:INFO|NOTICE)\s*:\s*([^:]+):\s*CAS compatible match after source cleanup\b`)

var oneLineRe = regexp.MustCompile(`(?i)^\s*(\d+(?:\.\d+)?)\s*([KMGTPE]?i?)(?:B)?\s*/\s*(\d+(?:\.\d+)?)\s*([KMGTPE]?i?)(?:B)?\s*,\s*(\d+(?:\.\d+)?)\s*([KMGTPE]?i?)(?:B)?/s\s*,\s*(?:ETA\s*([0-9hms:.-]+)|ETA\s*-|([0-9]{1,3})%)`) // kept for reference

var bytesPairRe = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*([KMGTPE]?i?)B\s*/\s*(\d+(?:\.\d+)?)\s*([KMGTPE]?i?)B`)
var speedTokenRe = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*([KMGTPE]?i?)(?:B)?/s`)
var pctTokenRe = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)%`)
var etaTokenRe = regexp.MustCompile(`(?i)ETA\s*([0-9dhms:.-]+|-)`)
var aggregateOneLineRe = regexp.MustCompile(`(?i)^\s*(?:\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2}\s+)?(?:INFO|NOTICE)\s*:\s*\d+(?:\.\d+)?\s*[KMGTPE]?i?B\s*/\s*\d+(?:\.\d+)?\s*[KMGTPE]?i?B\s*,\s*\d+(?:\.\d+)?%\s*,\s*\d+(?:\.\d+)?\s*[KMGTPE]?i?B/s\s*,\s*ETA\s*[0-9dhms:.-]+(?:\s*\(xfr#\d+(?:/\d+)?\))?\s*$`)

func unitToMul(u string) float64 {
	u = strings.ToUpper(u)
	switch u {
	case "K", "KI":
		return 1024
	case "M", "MI":
		return 1024 * 1024
	case "G", "GI":
		return 1024 * 1024 * 1024
	case "T", "TI":
		return 1024 * 1024 * 1024 * 1024
	case "P", "PI":
		return 1024 * 1024 * 1024 * 1024 * 1024
	default:
		return 1
	}
}

func parseUnit(s string) float64 {
	s = strings.TrimSpace(s)
	var num float64
	var unit string
	re := regexp.MustCompile(`([\d.]+)\s*([KMGTP]?i?B?)`)
	m := re.FindStringSubmatch(s)
	if len(m) < 3 {
		fmt.Sscanf(s, "%f", &num)
		return num
	}
	fmt.Sscanf(m[1], "%f", &num)
	unit = strings.ToUpper(m[2])
	unit = strings.TrimSuffix(unit, "B")
	if !strings.HasSuffix(unit, "I") && len(unit) > 0 && unit[len(unit)-1] == 'I' {
	} else if strings.HasSuffix(unit, "I") {
	} else {
		if unit == "K" || unit == "M" || unit == "G" || unit == "T" || unit == "P" {
			unit = unit + "I"
		}
	}
	return num * unitToMul(unit)
}

func parseETA(s string) int {
	s = strings.TrimSpace(s)
	if s == "-" || s == "" {
		return 0
	}
	if strings.Contains(s, ":") {
		parts := strings.Split(s, ":")
		if len(parts) == 3 {
			var h, m, sec int
			fmt.Sscanf(s, "%d:%d:%d", &h, &m, &sec)
			return h*3600 + m*60 + sec
		}
		if len(parts) == 2 {
			var m, sec int
			fmt.Sscanf(s, "%d:%d", &m, &sec)
			return m*60 + sec
		}
	}
	sec := 0
	hasD := strings.Contains(s, "d")
	hasH := strings.Contains(s, "h")
	hasM := strings.Contains(s, "m")
	hasS := strings.Contains(s, "s")
	var d, h, m, ss int
	switch {
	case hasD && hasH && hasM && hasS:
		fmt.Sscanf(s, "%dd%dh%dm%ds", &d, &h, &m, &ss)
	case hasD && hasH && hasM:
		fmt.Sscanf(s, "%dd%dh%dm", &d, &h, &m)
	case hasD && hasH && hasS:
		fmt.Sscanf(s, "%dd%dh%ds", &d, &h, &ss)
	case hasD && hasM && hasS:
		fmt.Sscanf(s, "%dd%dm%ds", &d, &m, &ss)
	case hasD && hasH:
		fmt.Sscanf(s, "%dd%dh", &d, &h)
	case hasD && hasM:
		fmt.Sscanf(s, "%dd%dm", &d, &m)
	case hasD && hasS:
		fmt.Sscanf(s, "%dd%ds", &d, &ss)
	case hasD:
		fmt.Sscanf(s, "%dd", &d)
	case hasH && hasM && hasS:
		fmt.Sscanf(s, "%dh%dm%ds", &h, &m, &ss)
	case hasH && hasM:
		fmt.Sscanf(s, "%dh%dm", &h, &m)
	case hasH && hasS:
		fmt.Sscanf(s, "%dh%ds", &h, &ss)
	case hasM && hasS:
		fmt.Sscanf(s, "%dm%ds", &m, &ss)
	case hasH:
		fmt.Sscanf(s, "%dh", &h)
	case hasM:
		fmt.Sscanf(s, "%dm", &m)
	case hasS:
		fmt.Sscanf(s, "%ds", &ss)
	}
	sec += d*24*3600 + h*3600 + m*60 + ss
	return sec
}

func parseOneLineProgress(line string) (map[string]any, bool) {
	l := strings.TrimSpace(line)
	if fileLineRe.MatchString(l) || fileCopiedRe.MatchString(l) {
		return nil, false
	}
	if !aggregateOneLineRe.MatchString(l) {
		return nil, false
	}
	xfrDone := float64(0)
	planned := float64(0)
	if i := strings.Index(l, "("); i >= 0 {
		paren := l[i:]
		l = strings.TrimSpace(l[:i])
		if j := strings.Index(strings.ToLower(paren), "xfr#"); j >= 0 {
			var a, b int
			n, _ := fmt.Sscanf(paren[j:], "xfr#%d/%d", &a, &b)
			if n >= 1 && a > 0 {
				xfrDone = float64(a)
			}
			if n == 2 && b > 0 {
				planned = float64(b)
			}
			if n == 0 {
				var aa int
				if _, err := fmt.Sscanf(paren[j:], "xfr#%d", &aa); err == nil && aa > 0 {
					xfrDone = float64(aa)
				}
			}
		}
	}
	bps := bytesPairRe.FindAllStringSubmatch(l, -1)
	if len(bps) == 0 {
		return nil, false
	}
	bp := bps[0]
	var cur, tot float64
	fmt.Sscanf(bp[1], "%f", &cur)
	fmt.Sscanf(bp[3], "%f", &tot)
	curBytes := cur * unitToMul(bp[2])
	totBytes := tot * unitToMul(bp[4])
	var sp float64
	spmAll := speedTokenRe.FindAllStringSubmatch(l, -1)
	var spm []string
	if len(spmAll) > 0 {
		spm = spmAll[len(spmAll)-1]
	}
	if len(spm) > 0 {
		fmt.Sscanf(spm[1], "%f", &sp)
	}
	spBytes := sp * unitToMul(spmValue(spm, 2))
	eta := 0
	if emAll := etaTokenRe.FindAllStringSubmatch(l, -1); len(emAll) > 0 {
		em := emAll[len(emAll)-1]
		if em[1] != "-" {
			eta = parseETA(em[1])
		}
	}
	prog := map[string]any{"bytes": curBytes, "totalBytes": totBytes, "speed": spBytes, "eta": float64(eta)}
	if pmAll := pctTokenRe.FindAllStringSubmatch(l, -1); len(pmAll) > 0 {
		pm := pmAll[len(pmAll)-1]
		var pct float64
		fmt.Sscanf(pm[1], "%f", &pct)
		prog["percentage"] = pct
	}
	if xfrDone > 0 {
		prog["completedFiles"] = xfrDone
	}
	if planned > 0 {
		prog["plannedFiles"] = planned
	}
	return prog, true
}

func spmValue(m []string, i int) string {
	if len(m) > i {
		return m[i]
	}
	return ""
}

func recomputeProgressPct(prog map[string]any) {
	if b, ok := prog["bytes"].(float64); ok {
		if tb, ok2 := prog["totalBytes"].(float64); ok2 && tb > 0 {
			prog["percentage"] = (b / tb) * 100
		}
	}
}