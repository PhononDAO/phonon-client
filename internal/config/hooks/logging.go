package hooks

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// SyslogHook to send logs via syslog.
type LoggingHook struct {
	loggingURL *url.URL
	key        string
}

func NewLoggingHook(apiKey string) *LoggingHook {
	urlstruct, err := url.Parse("https://logs.phonon.network/log")
	if err != nil {
		log.Fatal("Unable to configure logging hook: Unable to parse logging URL: " + err.Error())
	}
	fmt.Println("logging hook initiated")
	return &LoggingHook{
		loggingURL: urlstruct,
		key:        apiKey,
	}
}

func (h *LoggingHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read entry, %v", err)
		return err
	}
	req := &http.Request{
		Method: http.MethodPost,
		URL:    h.loggingURL,
		Body:   ioutil.NopCloser(strings.NewReader(line)),
		Header: http.Header{
			"ContentType": []string{"application/json"},
			"AuthToken":   []string{h.key},
		},
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("unable to send logs to telemetry server", err.Error())
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("unable to send logs to telemetry server", err.Error())
		return err
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println("unable to send logs to telemetry server", string(body))
	}
	return err
}

func (hook *LoggingHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
