package resources

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"
)

// matchAllRegexes is the test oracle: s must match every pattern (AND).
func matchAllRegexes(patterns []string, s string) (matched bool, err error) {
	if len(patterns) == 0 {
		return false, fmt.Errorf("no patterns")
	}
	for i, p := range patterns {
		re, compileErr := regexp.Compile(strings.TrimSpace(p))
		if compileErr != nil {
			return false, fmt.Errorf("regexes[%d]: %w", i, compileErr)
		}
		if !re.MatchString(s) {
			return false, nil
		}
	}
	return true, nil
}

func TestComposeANDRegexes_SingleReturnsAsIs(t *testing.T) {
	t.Parallel()
	got, err := composeANDRegexes([]string{"^.{1,40}$"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "^.{1,40}$" {
		t.Fatalf("got %q", got)
	}
}

func TestComposeANDRegexes_FirstNameLikePatterns_Equivalent(t *testing.T) {
	t.Parallel()

	const letters = `A-Za-zÀÁÂÃÄÅÆÇÈÉÊËÌÍÎÏÐÑÒÓÔÕÖÙÚÛÜÝÞßàáâãäåæçèéêëìíîïðñòóôõöùúûüýþÿØøĂăĆćČčĎďđĘęĚěĳĹĺĽľŁŃńŇňŐőŒœŔŕŘřŚŞşŠšŢţŤťŮůŰűŹŻŽžȘșȚț`
	const special = `/,` + "`" + `´\-'.&,\(\)`
	minOneLetter := `^.*[` + letters + `]+.*$`
	min1Max40 := `^.{1,40}$`
	spaceChars := `^[ ` + letters + special + `]*$`
	noPeriodFirst := `^[^.].*$`

	patterns := []string{minOneLetter, min1Max40, spaceChars, noPeriodFirst}
	composed, err := composeANDRegexes(patterns)
	if err != nil {
		t.Fatal(err)
	}
	re, err := regexp.Compile(composed)
	if err != nil {
		t.Fatalf("composed not RE2-valid: %v\npattern=%s", err, composed)
	}

	cases := []struct {
		in   string
		want bool
	}{
		{"Anna", true},
		{"José", true},
		{"A", true},
		{"", false},
		{".Anna", false},
		{"123", false},
		{strings.Repeat("A", 41), false},
		{"Ann@", false},
		{"Mary Jane", true},
	}
	for _, tc := range cases {
		ref, err := matchAllRegexes(patterns, tc.in)
		if err != nil {
			t.Fatalf("matchAllRegexes(%q): %v", tc.in, err)
		}
		if ref != tc.want {
			t.Fatalf("reference matchAllRegexes(%q)=%v want %v", tc.in, ref, tc.want)
		}
		got := re.MatchString(tc.in)
		if got != ref {
			t.Errorf("composed vs AND for %q: composed=%v and=%v\ncomposed=%s", tc.in, got, ref, composed)
		}
	}
}

func TestComposeANDRegexes_FirstNameLike_FuzzEquivalence(t *testing.T) {
	t.Parallel()

	const letters = `A-Za-z`
	const special = `/',.\-`
	minOneLetter := `^.*[` + letters + `]+.*$`
	min1Max40 := `^.{1,40}$`
	spaceChars := `^[ ` + letters + special + `]*$`
	noPeriodFirst := `^[^.].*$`
	patterns := []string{minOneLetter, min1Max40, spaceChars, noPeriodFirst}

	composed, err := composeANDRegexes(patterns)
	if err != nil {
		t.Fatal(err)
	}
	re, err := regexp.Compile(composed)
	if err != nil {
		t.Fatal(err)
	}

	alphabet := []rune(" ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz/',.\\-@#")
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 300; i++ {
		n := rng.Intn(45)
		var sb strings.Builder
		for j := 0; j < n; j++ {
			sb.WriteRune(alphabet[rng.Intn(len(alphabet))])
		}
		s := sb.String()
		ref, err := matchAllRegexes(patterns, s)
		if err != nil {
			t.Fatal(err)
		}
		got := re.MatchString(s)
		if got != ref {
			t.Fatalf("fuzz mismatch for %q (len=%d runes=%d): composed=%v and=%v\ncomposed=%s",
				s, len(s), utf8.RuneCountInString(s), got, ref, composed)
		}
	}
}

func TestComposeANDRegexes_CharsetAndLength(t *testing.T) {
	t.Parallel()
	composed, err := composeANDRegexes([]string{`^[A-Za-z]*$`, `^.{2,5}$`})
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(composed)
	patterns := []string{`^[A-Za-z]*$`, `^.{2,5}$`}
	for _, s := range []string{"", "A", "Ab", "Abcde", "Abcdef", "A1"} {
		ref, _ := matchAllRegexes(patterns, s)
		if re.MatchString(s) != ref {
			t.Errorf("%q: composed=%v and=%v pattern=%s", s, re.MatchString(s), ref, composed)
		}
	}
}

func TestComposeANDRegexes_EmptyLengthIntersection(t *testing.T) {
	t.Parallel()
	if _, err := composeANDRegexes([]string{`^.{1,3}$`, `^.{5,9}$`}); err == nil {
		t.Fatal("expected empty length intersection error")
	}
}

func TestComposeANDRegexes_RejectsPerlLookaheads(t *testing.T) {
	t.Parallel()
	if _, err := composeANDRegexes([]string{`^(?=.*[a-z]).*$`}); err == nil {
		t.Fatal("expected error for lookahead (unsupported in RE2)")
	}
}

func TestComposeANDRegexes_RejectsMultipleContains(t *testing.T) {
	t.Parallel()
	if _, err := composeANDRegexes([]string{`^.*[a-z].*$`, `^.*[A-Z].*$`, `^.{8,20}$`}); err == nil {
		t.Fatal("expected error for multiple contains")
	}
}

func TestComposeANDRegexes_RejectsUnknownShape(t *testing.T) {
	t.Parallel()
	if _, err := composeANDRegexes([]string{`^foo$`, `^bar$`}); err == nil {
		t.Fatal("expected error for unsupported shapes")
	}
}

func TestComposeANDRegexes_EmptyAndInvalid(t *testing.T) {
	t.Parallel()
	if _, err := composeANDRegexes(nil); err == nil {
		t.Fatal("expected error for empty list")
	}
	if _, err := composeANDRegexes([]string{"", "^a$"}); err == nil {
		t.Fatal("expected error for empty pattern")
	}
	if _, err := composeANDRegexes([]string{"["}); err == nil {
		t.Fatal("expected error for invalid pattern")
	}
}
