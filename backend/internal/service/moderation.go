package service

import (
	"regexp"
	"strings"
)

// ModerationResult is the outcome of checking one child's free-text
// PromptSubmission before it may reach any IntentClassifier.
type ModerationResult struct {
	Allowed bool
	Reason  string // human-readable, empty when Allowed
}

// Moderator gates free text before classification, per the HLD's
// "Модерация и безопасность" requirement: this runs before any LLM
// call, never after.
type Moderator interface {
	Check(text string) ModerationResult
}

// maxPromptLength is the Phase 0 length cap — abuse/DoS protection on
// the downstream LLM call, not a game-design number.
const maxPromptLength = 500

// bannedSubstrings is a minimal, non-exhaustive starting list for
// development. Per the spec's open questions, this needs a real content
// moderation / legal review before Phase 0 goes in front of children —
// it is not production-ready as-is.
var bannedSubstrings = []string{
	"урод",
	"идиот",
	"idiot",
	"дурак",
	"stupid",
}

var (
	emailLikePattern = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	phoneLikePattern = regexp.MustCompile(`(\+?\d[\d\s\-()]{8,}\d)`)
)

// BasicModerator is the Phase 0 ModerationModule: rule-based, no
// external calls. See the spec's open questions for its limits.
type BasicModerator struct{}

func (BasicModerator) Check(text string) ModerationResult {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return ModerationResult{Allowed: false, Reason: "напиши тактику словами"}
	}
	if len([]rune(trimmed)) > maxPromptLength {
		return ModerationResult{Allowed: false, Reason: "слишком длинный текст"}
	}

	lower := strings.ToLower(trimmed)
	for _, banned := range bannedSubstrings {
		if strings.Contains(lower, banned) {
			return ModerationResult{Allowed: false, Reason: "недопустимая формулировка"}
		}
	}

	if emailLikePattern.MatchString(trimmed) {
		return ModerationResult{Allowed: false, Reason: "не указывай контактные данные в тактике"}
	}
	if phoneLikePattern.MatchString(trimmed) {
		return ModerationResult{Allowed: false, Reason: "не указывай контактные данные в тактике"}
	}

	return ModerationResult{Allowed: true}
}
