package transport

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
)

type SSEEvent struct {
	event string
	data  string
}

// ReadSSEStream continuously reads the SSE stream and processes events.
func ReadSSEStream(ctx context.Context, reader io.ReadCloser, onEvent func(event SSEEvent)) error {
	defer func(reader io.ReadCloser) {
		err := reader.Close()
		if err != nil {
			fmt.Printf("Error closing reader: %v\n", err)
		}
	}(reader)

	scanner := bufio.NewScanner(reader)
	var event, data strings.Builder

	processEvent := func() {
		if event.Len() > 0 || data.Len() > 0 {
			onEvent(SSEEvent{event: event.String(), data: data.String()})
			event.Reset()
			data.Reset()
		}
	}

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil
		default:
			line := scanner.Text()

			switch {
			case strings.HasPrefix(line, "event:"):
				event.WriteString(strings.TrimSpace(strings.TrimPrefix(line, "event:")))
			case strings.HasPrefix(line, "data:"):
				if data.Len() > 0 {
					data.WriteString("\n")
				}
				data.WriteString(strings.TrimSpace(strings.TrimPrefix(line, "data:")))
			case line == "":
				processEvent()
			}
		}
	}
	// EOF Handle the last event after reaching EOF
	processEvent()

	return scanner.Err()
}
