package toolagent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// Usage is what one model run used, as its runner reports it.
type Usage struct {
	// Tokens are the input, output, and cache tokens of the messages of the model, each message once.
	Tokens int `json:"tokens"`
	// CostUSD is the cost that the runner reports at the end, if any.
	CostUSD float64 `json:"cost-usd,omitempty"`
	Turns   int     `json:"turns,omitempty"`
	// Reported is true when the runner reported any usage.
	Reported bool `json:"reported"`
}

// Meter reads the output of a runner as it comes, and counts the tokens of the model. It reads JSON
// lines in the form of claude -p --output-format stream-json --verbose: an "assistant" line has the
// usage of one message (a message with several parts repeats it, so each message id counts once),
// and the "result" line has the cost, the turns, and the final text. Other output passes through.
type Meter struct {
	mu       sync.Mutex
	partial  []byte
	messages map[string]int
	usage    Usage
	result   string
	hasText  bool
	// over is called once, when the tokens pass limit. limit 0: no limit.
	limit int
	over  func()
	fired bool
}

func NewMeter(limit int, over func()) *Meter {
	return &Meter{messages: map[string]int{}, limit: limit, over: over}
}

func (m *Meter) Write(p []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.partial = append(m.partial, p...)
	for {
		i := bytes.IndexByte(m.partial, '\n')
		if i < 0 {
			break
		}
		m.line(m.partial[:i])
		m.partial = m.partial[i+1:]
	}
	return len(p), nil
}

func (m *Meter) line(line []byte) {
	line = bytes.TrimSpace(line)
	if len(line) == 0 || line[0] != '{' {
		return
	}
	var event struct {
		Type    string `json:"type"`
		Message struct {
			ID    string `json:"id"`
			Usage *struct {
				Input         int `json:"input_tokens"`
				Output        int `json:"output_tokens"`
				CacheCreation int `json:"cache_creation_input_tokens"`
				CacheRead     int `json:"cache_read_input_tokens"`
			} `json:"usage"`
		} `json:"message"`
		Result  *string `json:"result"`
		CostUSD float64 `json:"total_cost_usd"`
		Turns   int     `json:"num_turns"`
	}
	if json.Unmarshal(line, &event) != nil {
		return
	}
	switch event.Type {
	case "assistant":
		u := event.Message.Usage
		if u == nil {
			return
		}
		m.usage.Reported = true
		id := event.Message.ID
		if id == "" {
			id = fmt.Sprintf("#%d", len(m.messages)) // a message with no id counts on its own
		}
		m.messages[id] = u.Input + u.Output + u.CacheCreation + u.CacheRead
		total := 0
		for _, tokens := range m.messages {
			total += tokens
		}
		m.usage.Tokens = total
		if m.limit > 0 && total > m.limit && !m.fired {
			m.fired = true
			if m.over != nil {
				m.over()
			}
		}
	case "result":
		m.usage.Reported = true
		m.usage.CostUSD, m.usage.Turns = event.CostUSD, event.Turns
		if event.Result != nil {
			m.result, m.hasText = *event.Result, true
		}
	}
}

// Usage returns what the run used so far.
func (m *Meter) Usage() Usage {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.usage
}

// Text returns the final text of the model: the result of a stream-json run, or else all of output.
func (m *Meter) Text(output string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.partial) > 0 {
		m.line(m.partial)
		m.partial = nil
	}
	if m.hasText {
		return m.result
	}
	return output
}

// ModelText returns the final text of the output of a runner: the result of a stream-json run, or
// else the output as it is.
func ModelText(output string) string {
	m := NewMeter(0, nil)
	m.Write([]byte(output))
	return m.Text(strings.TrimRight(output, "\n"))
}
