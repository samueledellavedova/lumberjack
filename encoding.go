package lumberjack

import "time"

const timestampFormat = time.RFC3339

func appendTimestamp(dst []byte, pretty bool) []byte {
  if pretty {
    return append(now().AppendFormat(append(dst, "\x1b[90m"...), timestampFormat), "\x1b[0m "...)
  }
  return append(now().AppendFormat(dst, timestampFormat), ' ')
}

func shouldQuote(s string) bool {
  for _, r := range s {
    if r == ' ' {
      return true
    }
  }
  return s == ""
}

func appendString(dst []byte, s string, quote bool) []byte {
  if quote {
    return append(append(append(dst, '"'), s...), '"')
  }
  return append(dst, s...)
}

func appendKey(dst []byte, k string, quote bool, pretty bool) []byte {
  if pretty {
    return append(append(appendString(append(append(dst, ' '), "\x1b[90m"...), k, quote), '='), "\x1b[0m"...)
  }
  return append(appendString(append(dst, ' '), k, quote), '=')
}