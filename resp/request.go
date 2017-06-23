package resp

import (
	"io"
)

// RequestReader is used by servers to wrap a client connection and convert
// requests into commands.
type RequestReader struct {
	r *bufioR
}

// NewRequestReader wraps any reader interface
func NewRequestReader(rd io.Reader) *RequestReader {
	r := new(bufioR)
	r.reset(mkStdBuffer(), rd)
	return &RequestReader{r: r}
}

// Buffered returns the number of unread bytes.
func (r *RequestReader) Buffered() int {
	return r.r.Buffered()
}

// Reset resets the reader to a new reader and recycles internal buffers.
func (r *RequestReader) Reset(rd io.Reader) {
	r.r.Reset(rd)
}

// PeekCmd peeks the next command name.
func (r *RequestReader) PeekCmd() (string, error) {
	return r.peekCmd(0)
}

// ReadCmd reads the next command. It optionally recycles the cmd passed.
func (r *RequestReader) ReadCmd(cmd *Command) (*Command, error) {
	if cmd == nil {
		cmd = new(Command)
	} else {
		cmd.reset()
	}

	err := parseCommand(cmd, r.r)
	return cmd, err
}

// StreamCmd reads the next command as a stream.
func (r *RequestReader) StreamCmd(cmd *CommandStream) (*CommandStream, error) {
	if cmd == nil {
		cmd = new(CommandStream)
	} else {
		cmd.reset()
	}

	err := parseCommand(cmd, r.r)
	return cmd, err
}

// SkipCmd skips the next command.
func (r *RequestReader) SkipCmd() error {
	c, err := r.r.PeekByte()
	if err != nil {
		return err
	}

	if c != '*' {
		_, err = r.r.ReadLine()
		return err
	}

	n, err := r.r.ReadArrayLen()
	if err != nil {
		return err
	}
	if n < 1 {
		return r.SkipCmd()
	}

	for i := 0; i < n; i++ {
		if err := r.r.SkipBulk(); err != nil {
			return err
		}
	}
	return nil
}

func (r *RequestReader) peekCmd(offset int) (string, error) {
	line, err := r.r.PeekLine(offset)
	if err != nil {
		return "", err
	}
	offset += len(line)

	if len(line) == 0 {
		return "", nil
	} else if line[0] != '*' {
		return line.FirstWord(), nil
	}

	n, err := line.ParseSize('*', errInvalidMultiBulkLength)
	if err != nil {
		return "", err
	}

	if n < 1 {
		return r.peekCmd(offset)
	}

	line, err = r.r.PeekLine(offset)
	if err != nil {
		return "", err
	}
	offset += len(line)

	n, err = line.ParseSize('$', errInvalidBulkLength)
	if err != nil {
		return "", err
	}

	data, err := r.r.PeekN(offset, int(n))
	return string(data), err
}

// --------------------------------------------------------------------

// RequestWriter is used by clients to send commands to servers.
type RequestWriter struct {
	w *bufioW
}

// NewRequestWriter wraps any Writer interface
func NewRequestWriter(wr io.Writer) *RequestWriter {
	w := new(bufioW)
	w.reset(mkStdBuffer(), wr)
	return &RequestWriter{w: w}
}

// Reset resets the writer with an new interface
func (w *RequestWriter) Reset(wr io.Writer) {
	w.w.Reset(wr)
}

// Buffered returns the number of buffered bytes
func (w *RequestWriter) Buffered() int {
	return w.w.Buffered()
}

// Flush flushes the output buffer. Call this after you have completed your pipeline
func (w *RequestWriter) Flush() error {
	return w.w.Flush()
}

// WriteCmd writes a full command as part of a pipeline. To execute the pipeline,
// you must call Flush.
func (w *RequestWriter) WriteCmd(cmd string, args ...[]byte) {
	w.w.AppendArrayLen(len(args) + 1)
	w.w.AppendBulkString(cmd)
	for _, arg := range args {
		w.w.AppendBulk(arg)
	}
}

// WriteCmdString writes a full command as part of a pipeline. To execute the pipeline,
// you must call Flush.
func (w *RequestWriter) WriteCmdString(cmd string, args ...string) {
	w.w.AppendArrayLen(len(args) + 1)
	w.w.AppendBulkString(cmd)
	for _, arg := range args {
		w.w.AppendBulkString(arg)
	}
}

// WriteMultiBulkSize is a low-level method to write a multibulk size.
// For normal operation, use WriteCmd or WriteCmdString.
func (w *RequestWriter) WriteMultiBulkSize(n int) error {
	if n < 0 {
		return errInvalidMultiBulkLength
	}
	w.w.AppendArrayLen(n)
	return nil
}

// WriteBulk is a low-level method to write a bulk.
// For normal operation, use WriteCmd or WriteCmdString.
func (w *RequestWriter) WriteBulk(b []byte) {
	w.w.AppendBulk(b)
}

// WriteBulkString is a low-level method to write a bulk.
// For normal operation, use WriteCmd or WriteCmdString.
func (w *RequestWriter) WriteBulkString(s string) {
	w.w.AppendBulkString(s)
}

// CopyBulk is a low-level method to copy a large bulk of data directly to the writer.
// For normal operation, use WriteCmd or WriteCmdString.
func (w *RequestWriter) CopyBulk(r io.Reader, n int64) error {
	return w.w.CopyBulk(r, n)
}
