package runnercli

import (
	"sync"
	"sync/atomic"
)

type fileProg struct {
	Name  string  `json:"name"`
	Bytes float64 `json:"bytes"`
	Total float64 `json:"totalBytes"`
	Pct   float64 `json:"percentage"`
	Speed float64 `json:"speed"`
}

type fileProgress struct {
	m           map[string]*fileProg
	mu          sync.Mutex
	ver         uint64
	copied      []string
	recent      []fileProg
	recentCap   int
	recentStart int
	recentLen   int
}

func (fp *fileProgress) pushRecent(p fileProg) {
	if fp.recentCap <= 0 {
		return
	}
	if fp.recentLen < fp.recentCap {
		fp.recent = append(fp.recent, p)
		fp.recentLen++
		return
	}
	fp.recent[fp.recentStart] = p
	fp.recentStart = (fp.recentStart + 1) % fp.recentCap
}

func (fp *fileProgress) update(name string, bytes, total, speed, pct float64) {
	fp.mu.Lock()
	p, ok := fp.m[name]
	if !ok {
		p = &fileProg{Name: name}
		fp.m[name] = p
	}
	if total > 0 {
		p.Total = total
	}
	if bytes >= 0 {
		p.Bytes = bytes
	}
	if speed >= 0 {
		p.Speed = speed
	}
	if pct >= 0 {
		p.Pct = pct
	}
	fp.pushRecent(*p)
	fp.mu.Unlock()
	atomic.AddUint64(&fp.ver, 1)
}

func (fp *fileProgress) markCopied(name string) {
	fp.mu.Lock()
	fp.copied = append(fp.copied, name)
	if v, ok := fp.m[name]; ok {
		fp.pushRecent(*v)
	}
	fp.mu.Unlock()
}

func (fp *fileProgress) snapshot(limit int) []fileProg {
	fp.mu.Lock()
	defer fp.mu.Unlock()
	if limit > 0 && fp.recentCap != limit {
		fp.recentCap = limit
	}
	if fp.recentLen == 0 {
		return nil
	}
	out := make([]fileProg, 0, fp.recentLen)
	if fp.recentStart+fp.recentLen <= fp.recentCap {
		out = append(out, fp.recent[fp.recentStart:fp.recentStart+fp.recentLen]...)
	} else {
		first := fp.recentCap - fp.recentStart
		out = append(out, fp.recent[fp.recentStart:]...)
		out = append(out, fp.recent[:fp.recentLen-first]...)
	}
	return out
}

func (fp *fileProgress) copiedList() []string {
	fp.mu.Lock()
	defer fp.mu.Unlock()
	out := make([]string, len(fp.copied))
	copy(out, fp.copied)
	return out
}