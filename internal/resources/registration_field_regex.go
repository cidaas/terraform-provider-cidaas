package resources

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// composeANDRegexes validates patterns with Go's regexp (RE2) and returns one
// RE2 pattern with AND semantics for supported validation shapes (length, charset,
// contains, no_leading). Patterns are not concatenated and lookaheads are not used.
// Unsupported or ambiguous combinations fail closed.
func composeANDRegexes(patterns []string) (string, error) {
	if len(patterns) == 0 {
		return "", fmt.Errorf("regexes must contain at least one pattern")
	}

	cleaned := make([]string, 0, len(patterns))
	for i, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			return "", fmt.Errorf("regexes[%d] is empty", i)
		}
		if _, err := regexp.Compile(p); err != nil {
			return "", fmt.Errorf("regexes[%d] is not a valid Go regexp: %w", i, err)
		}
		cleaned = append(cleaned, p)
	}

	if len(cleaned) == 1 {
		return cleaned[0], nil
	}

	var c mergedConstraints
	for i, p := range cleaned {
		sh, err := classifyRegexShape(p)
		if err != nil {
			return "", fmt.Errorf("regexes[%d]: %w", i, err)
		}
		if err := c.apply(sh); err != nil {
			return "", fmt.Errorf("regexes[%d]: %w", i, err)
		}
	}

	composed, err := c.build()
	if err != nil {
		return "", err
	}
	if _, err := regexp.Compile(composed); err != nil {
		return "", fmt.Errorf("composed regex is not a valid Go regexp: %w", err)
	}
	return composed, nil
}

type shapeKind int

const (
	shapeLength shapeKind = iota
	shapeCharset
	shapeContains
	shapeNoLeading
)

type regexShape struct {
	kind          shapeKind
	minLen        int
	maxLen        int // -1 = unbounded
	classBody     string
	quant         string // "*" or "+"
	forbidLeading rune
	hasForbid     bool
}

var (
	reLengthExact   = regexp.MustCompile(`^\^\.\{(\d+)\}\$$`)
	reLengthBounded = regexp.MustCompile(`^\^\.\{(\d+),(\d+)\}\$$`)
	reLengthMin     = regexp.MustCompile(`^\^\.\{(\d+),\}\$$`)
	reCharset       = regexp.MustCompile(`^\^\[((?:[^\\\[\]]|\\.)+)\]([*+])\$$`)
	reContains      = regexp.MustCompile(`^\^\.\*\[((?:[^\\\[\]]|\\.)+)\]\+?\.\*\$$`)
	reNoLeading     = regexp.MustCompile(`^\^\[\^((?:[^\\\[\]]|\\.))\]\.\*\$$`)
)

func classifyRegexShape(p string) (regexShape, error) {
	if m := reLengthExact.FindStringSubmatch(p); m != nil {
		n, _ := strconv.Atoi(m[1])
		return regexShape{kind: shapeLength, minLen: n, maxLen: n}, nil
	}
	if m := reLengthBounded.FindStringSubmatch(p); m != nil {
		minV, _ := strconv.Atoi(m[1])
		maxV, _ := strconv.Atoi(m[2])
		return regexShape{kind: shapeLength, minLen: minV, maxLen: maxV}, nil
	}
	if m := reLengthMin.FindStringSubmatch(p); m != nil {
		minV, _ := strconv.Atoi(m[1])
		return regexShape{kind: shapeLength, minLen: minV, maxLen: -1}, nil
	}
	if m := reCharset.FindStringSubmatch(p); m != nil {
		return regexShape{kind: shapeCharset, classBody: m[1], quant: m[2]}, nil
	}
	if m := reContains.FindStringSubmatch(p); m != nil {
		return regexShape{kind: shapeContains, classBody: m[1]}, nil
	}
	if m := reNoLeading.FindStringSubmatch(p); m != nil {
		r, ok := parseSingleClassRune(m[1])
		if !ok {
			return regexShape{}, fmt.Errorf("unsupported no_leading pattern %q (only a single forbidden first character is supported)", p)
		}
		return regexShape{kind: shapeNoLeading, forbidLeading: r, hasForbid: true}, nil
	}
	return regexShape{}, fmt.Errorf("unsupported pattern shape %q (supported: length ^.{m,n}$, charset ^[…]*$/^+$, contains ^.*[…].*$, no_leading ^[^x].*$); provide a single field_definition.regex instead", p)
}

func parseSingleClassRune(body string) (rune, bool) {
	if body == "" {
		return 0, false
	}
	if strings.HasPrefix(body, `\`) && len(body) >= 2 {
		r, size := utf8.DecodeRuneInString(body[1:])
		if r == utf8.RuneError && size == 1 {
			return 0, false
		}
		if body[1+size:] != "" {
			return 0, false
		}
		return r, true
	}
	r, size := utf8.DecodeRuneInString(body)
	if r == utf8.RuneError && size == 1 {
		return 0, false
	}
	if body[size:] != "" {
		return 0, false
	}
	return r, true
}

type mergedConstraints struct {
	hasLength bool
	minLen    int
	maxLen    int // -1 unbounded

	hasCharset bool
	charset    string
	quant      string

	contains []string

	hasNoLeading bool
	forbid       rune
}

func (c *mergedConstraints) apply(sh regexShape) error {
	switch sh.kind {
	case shapeLength:
		if !c.hasLength {
			c.hasLength = true
			c.minLen = sh.minLen
			c.maxLen = sh.maxLen
			return nil
		}
		if sh.minLen > c.minLen {
			c.minLen = sh.minLen
		}
		switch {
		case c.maxLen < 0 && sh.maxLen >= 0:
			c.maxLen = sh.maxLen
		case c.maxLen >= 0 && sh.maxLen >= 0 && sh.maxLen < c.maxLen:
			c.maxLen = sh.maxLen
		}
		if c.maxLen >= 0 && c.minLen > c.maxLen {
			return fmt.Errorf("length constraints have empty intersection (min=%d max=%d)", c.minLen, c.maxLen)
		}
		return nil
	case shapeCharset:
		if c.hasCharset {
			if c.charset != sh.classBody {
				return fmt.Errorf("cannot merge distinct charset patterns (fail-closed); provide a single field_definition.regex")
			}
			if sh.quant == "+" || c.quant == "+" {
				c.quant = "+"
			}
			return nil
		}
		c.hasCharset = true
		c.charset = sh.classBody
		c.quant = sh.quant
		return nil
	case shapeContains:
		c.contains = append(c.contains, sh.classBody)
		return nil
	case shapeNoLeading:
		if c.hasNoLeading && c.forbid != sh.forbidLeading {
			return fmt.Errorf("cannot merge distinct no_leading constraints")
		}
		c.hasNoLeading = true
		c.forbid = sh.forbidLeading
		return nil
	default:
		return fmt.Errorf("internal: unknown shape")
	}
}

func (c *mergedConstraints) build() (string, error) {
	if len(c.contains) > 1 {
		return "", fmt.Errorf("regexes: multiple contains constraints cannot be AND-merged into one RE2 pattern (no lookaheads); provide a single field_definition.regex")
	}

	minLen := 0
	maxLen := -1
	if c.hasLength {
		minLen = c.minLen
		maxLen = c.maxLen
	}
	if c.hasCharset && c.quant == "+" && minLen < 1 {
		minLen = 1
	}
	if c.hasNoLeading && minLen < 1 {
		minLen = 1
	}
	if len(c.contains) == 1 && minLen < 1 {
		minLen = 1
	}
	if maxLen >= 0 && minLen > maxLen {
		return "", fmt.Errorf("length constraints have empty intersection (min=%d max=%d)", minLen, maxLen)
	}

	switch {
	case !c.hasCharset && len(c.contains) == 0 && !c.hasNoLeading && c.hasLength:
		return buildLengthOnly(minLen, maxLen), nil
	case c.hasCharset && len(c.contains) == 0:
		return buildCharsetLengthNoLeading(c.charset, minLen, maxLen, c.hasNoLeading, c.forbid)
	case c.hasCharset && len(c.contains) == 1:
		if maxLen < 0 {
			return "", fmt.Errorf("regexes: contains+charset merge requires a bounded max length (e.g. ^.{1,40}$); provide a single field_definition.regex")
		}
		if maxLen > 64 {
			return "", fmt.Errorf("regexes: max length %d too large for contains merge (limit 64 to keep composed regex size reasonable)", maxLen)
		}
		return buildContainsCharset(c.charset, c.contains[0], minLen, maxLen, c.hasNoLeading, c.forbid)
	default:
		return "", fmt.Errorf("regexes: unsupported combination of shapes for RE2 merge; provide a single field_definition.regex")
	}
}

func buildLengthOnly(minLen, maxLen int) string {
	if maxLen < 0 {
		return fmt.Sprintf("^.{%d,}$", minLen)
	}
	if minLen == maxLen {
		return fmt.Sprintf("^.{%d}$", minLen)
	}
	return fmt.Sprintf("^.{%d,%d}$", minLen, maxLen)
}

func buildCharsetLengthNoLeading(charset string, minLen, maxLen int, noLeading bool, forbid rune) (string, error) {
	if minLen < 0 {
		minLen = 0
	}
	a := "[" + charset + "]"
	if !noLeading {
		if maxLen < 0 {
			if minLen == 0 {
				return "^" + a + "*$", nil
			}
			return fmt.Sprintf("^%s{%d,}$", a, minLen), nil
		}
		if minLen == maxLen {
			return fmt.Sprintf("^%s{%d}$", a, minLen), nil
		}
		return fmt.Sprintf("^%s{%d,%d}$", a, minLen, maxLen), nil
	}

	firstBody, ok := removeLiteralRuneFromClassBody(charset, forbid)
	if !ok {
		return "", fmt.Errorf("cannot apply no_leading to charset (fail-closed)")
	}
	if firstBody == "" {
		return "", fmt.Errorf("no_leading removes all charset characters for the first position")
	}
	first := "[" + firstBody + "]"
	if minLen < 1 {
		minLen = 1
	}
	if maxLen < 0 {
		return fmt.Sprintf("^%s%s*$", first, a), nil
	}
	lo := minLen - 1
	hi := maxLen - 1
	if lo == hi {
		return fmt.Sprintf("^%s%s{%d}$", first, a, lo), nil
	}
	return fmt.Sprintf("^%s%s{%d,%d}$", first, a, lo, hi), nil
}

func buildContainsCharset(charset, need string, minLen, maxLen int, noLeading bool, forbid rune) (string, error) {
	if minLen < 1 {
		minLen = 1
	}
	if maxLen < minLen {
		return "", fmt.Errorf("length constraints have empty intersection")
	}

	a := "[" + charset + "]"
	n := "[" + need + "]"

	firstBody := charset
	if noLeading {
		var ok bool
		firstBody, ok = removeLiteralRuneFromClassBody(charset, forbid)
		if !ok || firstBody == "" {
			return "", fmt.Errorf("cannot apply no_leading to charset (fail-closed)")
		}
	}
	first := "[" + firstBody + "]"

	needFirstBody := need
	if noLeading {
		if b, ok := removeLiteralRuneFromClassBody(need, forbid); ok {
			needFirstBody = b
		}
	}

	var alts []string
	for k := minLen; k <= maxLen; k++ {
		if k == 1 {
			if needFirstBody == "" {
				continue
			}
			alts = append(alts, "["+needFirstBody+"]")
			continue
		}
		if needFirstBody != "" {
			alts = append(alts, fmt.Sprintf("%s%s{%d}", "["+needFirstBody+"]", a, k-1))
		}
		for i := 1; i < k; i++ {
			before := i - 1
			after := k - i - 1
			alts = append(alts, fmt.Sprintf("%s%s{%d}%s%s{%d}", first, a, before, n, a, after))
		}
	}
	if len(alts) == 0 {
		return "", fmt.Errorf("contains+charset+length merge produced no alternatives")
	}
	return "^(?:" + strings.Join(alts, "|") + ")$", nil
}

func removeLiteralRuneFromClassBody(body string, forbid rune) (string, bool) {
	var b strings.Builder
	i := 0
	for i < len(body) {
		if body[i] == '\\' && i+1 < len(body) {
			r, size := utf8.DecodeRuneInString(body[i+1:])
			if r == forbid {
				i += 1 + size
				continue
			}
			b.WriteByte('\\')
			b.WriteString(body[i+1 : i+1+size])
			i += 1 + size
			continue
		}
		if i+2 < len(body) && body[i+1] == '-' {
			// Keep ranges intact (do not attempt to subtract from ranges).
			b.WriteByte(body[i])
			b.WriteByte('-')
			_, size2 := utf8.DecodeRuneInString(body[i+2:])
			b.WriteString(body[i+2 : i+2+size2])
			i += 2 + size2
			continue
		}
		r, size := utf8.DecodeRuneInString(body[i:])
		if r == forbid {
			i += size
			continue
		}
		b.WriteString(body[i : i+size])
		i += size
	}
	return b.String(), true
}
