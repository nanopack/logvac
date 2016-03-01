package logvac

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/jcelliott/lumber"
)

type Publisher interface {
	Publish(tag []string, data string) error
}

func Filter(drain Drain, level int) Drain {
	return func(log Logger, msg Message) {
		if msg.Priority >= level {
			drain(log, msg)
		}
	}
}

func WriteDrain(writer io.Writer) Drain {
	return func(log Logger, msg Message) {
		writer.Write([]byte(fmt.Sprintf("[%s][%s] <%d> %s\n", msg.Type, msg.Time, msg.Priority, msg.Content)))
	}
}

func PublishDrain(publisher Publisher) Drain {
	return func(log Logger, msg Message) {
		tags := []string{"log", msg.Type}
		severities := []string{"trace", "debug", "info", "warn", "error", "fatal"}
		tags = append(tags, severities[:((msg.Priority+1)%6)]...)
		data, err := json.Marshal(msg)
		if err != nil {
			return
		}
		publisher.Publish(tags, string(data))
	}
}

func LogDrain(logger Logger) Drain {
	return func(log Logger, msg Message) {
		switch msg.Priority {
		case lumber.TRACE:
			logger.Trace(msg.Content)
		case lumber.DEBUG:
			logger.Debug(msg.Content)
		case lumber.INFO:
			logger.Info(msg.Content)
		case lumber.WARN:
			logger.Warn(msg.Content)
		case lumber.ERROR:
			logger.Error(msg.Content)
		case lumber.FATAL:
			logger.Fatal(msg.Content)
		}
	}
}
