package chromatix

import (
  "bytes"

  "io"
  "log"
  "net/http"
  "os"
  "text/template"
  "time"

  "github.com/urfave/negroni"
)

// LogEntry is the structure
// passed to the template.
type LogEntry struct {
  StartTime string
  Status    int
  Duration  time.Duration
  Hostname  string
  Method    string
  Path      string
}

// LogDefaultFormat is the format
// logged used by the default Log instance.
var LogDefaultFormat = "{{.StartTime}} | {{.Status}} | \t {{.Duration}} | {{.Hostname}} | {{.Method}} {{.Path}} \n"

// LogDefaultDateFormat is the
// format used for date by the
// default Log instance.
var LogDefaultDateFormat = time.RFC3339

// LoggerInterface interface
type LoggerInterface interface {
  Println(v ...interface{})
  Printf(format string, v ...interface{})
  SetOutput(w io.Writer)
}

// Log is a middleware handler that logs the request as it goes in and the response as it goes out.
type Logger struct {
  // LoggerInterface implements more log.Logger interface to be compatible with other implementations
  LoggerInterface
  dateFormat string
  template   *template.Template
}

// NewLogger returns a new Logger instance
func NewLogger(w io.Writer) *Logger {
  if w == nil {
    w = os.Stdout
  }
  logger := &Logger{LoggerInterface: log.New(w, "[negroni] ", 0), dateFormat: LogDefaultDateFormat}
  logger.SetFormat(LogDefaultFormat)
  return logger
}

func (l *Logger) SetFormat(format string) {
  l.template = template.Must(template.New("negroni_parser").Parse(format))
}

func (l *Logger) SetDateFormat(format string) {
  l.dateFormat = format
}

func (l *Logger) SetOutput(w io.Writer) {
  l.LoggerInterface.SetOutput(w)
}

func (l *Logger) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
  start := time.Now()

  next(rw, r)

  res := rw.(negroni.ResponseWriter)
  log := LogEntry{
    StartTime: start.Format(l.dateFormat),
    Status:    res.Status(),
    Duration:  time.Since(start),
    Hostname:  r.Host,
    Method:    r.Method,
    Path:      r.URL.Path,
  }

  buff := &bytes.Buffer{}
  l.template.Execute(buff, log)
  l.Printf(buff.String())
}
