package resp

var asciiSpace = [256]bool{'\t': true, '\n': true, '\v': true, '\f': true, '\r': true, ' ': true}

// 'Inspired' by sdssplitargs from https://github.com/antirez/sds
func appendArgument(dst, src []byte) ([]byte, int) {
	// skip initial blanks
	pos := 0
	for ; pos < len(src) && asciiSpace[src[pos]]; pos++ {
	}

	var inQ, inSQ bool
MainLoop:
	for pos < len(src) {
		p := src[pos]

		if inQ {
			if p == '"' {
				pos++
				break MainLoop
			} else if p == '\\' && pos+3 < len(src) && src[pos+1] == 'x' && isHexChar(src[pos+2]) && isHexChar(src[pos+3]) {
				p = fromHexChar(src[pos+2])<<4 | fromHexChar(src[pos+3])
				pos += 3
			} else if p == '\\' && pos+1 < len(src) {
				pos++
				p = src[pos]
				switch p {
				case 'n':
					p = '\n'
				case 'r':
					p = '\r'
				case 't':
					p = '\t'
				case 'b':
					p = '\b'
				case 'a':
					p = '\a'
				}
			}
			dst = append(dst, p)

		} else if inSQ {
			if p == '\'' {
				pos++
				break MainLoop
			}
			dst = append(dst, p)

		} else {
			switch p {
			case ' ', '\n', '\r', '\t':
				break MainLoop
			case '"':
				if len(dst) != 0 {
					break MainLoop
				}
				inQ = true
			case '\'':
				if len(dst) != 0 {
					break MainLoop
				}
				inSQ = true
			default:
				dst = append(dst, p)
			}
		}
		pos++
	}
	return dst, pos
}

// --------------------------------------------------------------------

func isHexChar(c byte) bool {
	return (c >= '0' && c <= '9') ||
		(c >= 'a' && c <= 'f') ||
		(c >= 'A' && c <= 'F')
}

func fromHexChar(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}
