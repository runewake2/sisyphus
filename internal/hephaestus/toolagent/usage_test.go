package toolagent

import (
	"strings"
	"testing"
)

const streamJSON = `{"type":"system","subtype":"init"}
{"type":"assistant","message":{"id":"m1","usage":{"input_tokens":100,"output_tokens":20,"cache_creation_input_tokens":5,"cache_read_input_tokens":0}}}
{"type":"assistant","message":{"id":"m1","usage":{"input_tokens":100,"output_tokens":20,"cache_creation_input_tokens":5,"cache_read_input_tokens":0}}}
{"type":"assistant","message":{"id":"m2","usage":{"input_tokens":200,"output_tokens":30,"cache_creation_input_tokens":0,"cache_read_input_tokens":120}}}
{"type":"result","subtype":"success","result":"Done.\n{\"vote\": \"yes\"}","total_cost_usd":0.0123,"num_turns":2}
`

func TestTheMeterCountsEachMessageOnce(t *testing.T) {
	m := NewMeter(0, nil)
	for _, part := range strings.SplitAfter(streamJSON, "e") { // in pieces, as a pipe gives it
		m.Write([]byte(part))
	}

	u := m.Usage()

	if !u.Reported || u.Tokens != 125+350 || u.CostUSD != 0.0123 || u.Turns != 2 {
		t.Fatalf("want 475 tokens, $0.0123, 2 turns, got %+v", u)
	}
	if text := m.Text(streamJSON); text != "Done.\n{\"vote\": \"yes\"}" {
		t.Errorf("want the result text, got %q", text)
	}
}

func TestTheMeterStopsARunAtItsLimit(t *testing.T) {
	stopped := 0
	m := NewMeter(300, func() { stopped++ })

	m.Write([]byte(streamJSON))

	if stopped != 1 {
		t.Errorf("want one stop at the limit, got %d", stopped)
	}
}

func TestPlainOutputIsItsOwnText(t *testing.T) {
	if text := ModelText("I checked it.\n{\"vote\": \"no\"}\n"); text != "I checked it.\n{\"vote\": \"no\"}" {
		t.Errorf("got %q", text)
	}
}
