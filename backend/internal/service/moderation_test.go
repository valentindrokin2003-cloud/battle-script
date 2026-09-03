package service

import "testing"

func TestBasicModerator_Check(t *testing.T) {
	m := BasicModerator{}
	cases := []struct {
		name        string
		text        string
		wantAllowed bool
	}{
		{"valid tactic text", "Провоцируй босса, когда он целится в целителя, иначе атакуй слабейшего врага", true},
		{"empty text", "", false},
		{"whitespace only", "   \t\n  ", false},
		{"too long", stringOfLength(600), false},
		{"exactly at length limit", stringOfLength(500), true},
		{"contains profanity ru", "атакуй этого урода в первую очередь", false},
		{"contains profanity en", "attack that idiot boss first", false},
		{"contains email-like text", "напиши мне на test@example.com если что", false},
		{"contains phone-like text", "мой номер +7 999 123 45 67 запомни", false},
		{"normal text mentioning numbers not a phone", "отступи если щит ниже 30 процентов", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := m.Check(tc.text)
			if got.Allowed != tc.wantAllowed {
				t.Errorf("Check(%q) = Allowed:%v Reason:%q, want Allowed:%v", tc.text, got.Allowed, got.Reason, tc.wantAllowed)
			}
			if !got.Allowed && got.Reason == "" {
				t.Errorf("Check(%q) rejected with empty Reason", tc.text)
			}
		})
	}
}

func stringOfLength(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}
