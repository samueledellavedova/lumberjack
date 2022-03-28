package log

import (
  "os"
  "strconv"
  "sync"
)

var entryPool = sync.Pool{
  New: func() interface{} {
    return newEntry()
  },
}

type entry struct {
  logger *Logger
  fields map[string]*field
  buf    []byte
  exit   bool
}

func newEntry() *entry {
  return &entry{
    fields: make(map[string]*field),
    buf:    make([]byte, 0, 512),
  }
}

func (e *entry) Logger() *Logger {
  return e.logger
}

func (e *entry) Flush() {
  if e == nil || e.logger.cfg.Writer == nil {
    return
  }

  if e.exit {
    defer os.Exit(0)
  }

  buf := append(e.buf, '\n')
  _, _ = e.logger.cfg.Writer.Write(buf)

  for k, f := range e.fields {
    delete(e.fields, k)
    fieldPool.Put(f)
  }

  if cap(buf) <= 1<<16 {
    e.buf = buf[:0]
    entryPool.Put(e)
  }
}

func (e *entry) split(klen, vlen, start, end int) (head, tail []byte) {
  //        ' '  ....  '='
  voffset := 1 + klen + 1

  if diff := end - start - voffset - vlen; diff != 0 {
    e.buf = append(e.buf[:end-diff], e.buf[end:]...)

    for _, f := range e.fields {
      if f.start >= start {
        f.end -= diff
        if f.start != start {
          f.start -= diff
        } else {
          end = f.end
        }
      }
    }
  }

  buf := e.buf
  return buf[:start+voffset], buf[end:]
}

func (e *entry) str(k string, kquote bool, v string, vquote bool) *entry {
  pretty := e.logger.cfg.Pretty

  if f, exists := e.fields[k]; exists {
    klen, vlen := len(k), len(v)

    if kquote {
      klen += 2
    }

    if vquote {
      vlen += 2
    }

    if pretty {
      klen += 9 // len("\x1b[90m") + len("\x1b[0m") = 9
    }

    head, tail := e.split(klen, vlen, f.start, f.end)
    e.buf = append(appendString(head, v, vquote), tail...)

    return e
  }

  buf := e.buf
  e.buf = appendString(appendKey(buf, k, kquote, pretty), v, vquote)

  f := fieldPool.Get().(*field)
  f.start = len(buf)
  f.end = len(e.buf)
  e.fields[k] = f

  return e
}

func (e *entry) Str(k, v string) *entry {
  if e == nil {
    return e
  }
  return e.str(k, shouldQuote(k), v, shouldQuote(v))
}

func (e *entry) Trace() *entry {
  if e == nil {
    return e
  }
  return e.Str("stack", captureStackTrace(1))
}

func (e *entry) Err(err error) *entry {
  if err == nil {
    return e.Str("err", "nil")
  }
  return e.Str("err", err.Error())
}

func (e *entry) Int(k string, v int) *entry {
  if e == nil {
    return e
  }
  return e.Str(k, strconv.FormatInt(int64(v), 10))
}

func (e *entry) Uint16(k string, v uint16) *entry {
  if e == nil {
    return e
  }
  return e.Str(k, strconv.FormatUint(uint64(v), 10))
}

func (e *entry) Uint64(k string, v uint64) *entry {
  if e == nil {
    return e
  }
  return e.Str(k, strconv.FormatUint(v, 10))
}