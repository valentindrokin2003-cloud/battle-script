package service

import (
	"context"
	"errors"
	"testing"
)

type mockClassifier struct {
	calls     int
	responses []func() (IntentClassification, error)
}

func (m *mockClassifier) Classify(_ context.Context, _ ClassificationContext) (IntentClassification, error) {
	i := m.calls
	m.calls++
	if i >= len(m.responses) {
		return IntentClassification{}, errors.New("mock exhausted")
	}
	return m.responses[i]()
}

func validClassification(class HeroClass) IntentClassification {
	return IntentClassification{
		HeroClass:      class,
		SchemaVersion:  CurrentSchemaVersion,
		Rules:          []Rule{{Priority: 0, Condition: Condition{Type: ConditionAlways}, Action: DefaultFallback(class)}},
		FallbackAction: DefaultFallback(class),
		Confidence:     ConfidenceHigh,
	}
}

func TestClassifyWithFallback_SucceedsOnFirstTry(t *testing.T) {
	want := validClassification(HeroClassMage)
	mock := &mockClassifier{responses: []func() (IntentClassification, error){
		func() (IntentClassification, error) { return want, nil },
	}}
	input := ClassificationContext{HeroClass: HeroClassMage, Boss: frostWarden()}

	got := ClassifyWithFallback(context.Background(), mock, input)

	if got.Confidence != ConfidenceHigh {
		t.Errorf("Confidence = %v, want high", got.Confidence)
	}
	if mock.calls != 1 {
		t.Errorf("classifier called %d times, want 1", mock.calls)
	}
}

func TestClassifyWithFallback_RetriesOnceThenSucceeds(t *testing.T) {
	want := validClassification(HeroClassTank)
	mock := &mockClassifier{responses: []func() (IntentClassification, error){
		func() (IntentClassification, error) { return IntentClassification{}, errors.New("boom") },
		func() (IntentClassification, error) { return want, nil },
	}}
	input := ClassificationContext{HeroClass: HeroClassTank, Boss: frostWarden()}

	got := ClassifyWithFallback(context.Background(), mock, input)

	if got.Confidence != ConfidenceHigh {
		t.Errorf("Confidence = %v, want high (retry succeeded)", got.Confidence)
	}
	if mock.calls != 2 {
		t.Errorf("classifier called %d times, want 2", mock.calls)
	}
}

func TestClassifyWithFallback_InvalidResultTreatedAsFailure(t *testing.T) {
	invalid := IntentClassification{HeroClass: "not_a_class"}
	mock := &mockClassifier{responses: []func() (IntentClassification, error){
		func() (IntentClassification, error) { return invalid, nil }, // err nil but fails ValidateIntentClassification
		func() (IntentClassification, error) { return validClassification(HeroClassHealer), nil },
	}}
	input := ClassificationContext{HeroClass: HeroClassHealer, Boss: frostWarden()}

	got := ClassifyWithFallback(context.Background(), mock, input)

	if got.Confidence != ConfidenceHigh {
		t.Errorf("Confidence = %v, want high (retry produced a valid result)", got.Confidence)
	}
	if mock.calls != 2 {
		t.Errorf("classifier called %d times, want 2", mock.calls)
	}
}

func TestClassifyWithFallback_FallsBackAfterOneRetry(t *testing.T) {
	mock := &mockClassifier{responses: []func() (IntentClassification, error){
		func() (IntentClassification, error) { return IntentClassification{}, errors.New("boom 1") },
		func() (IntentClassification, error) { return IntentClassification{}, errors.New("boom 2") },
	}}
	input := ClassificationContext{HeroClass: HeroClassArcher, Boss: frostWarden()}

	got := ClassifyWithFallback(context.Background(), mock, input)

	if got.Confidence != ConfidenceLowFallbackUsed {
		t.Errorf("Confidence = %v, want low_fallback_used", got.Confidence)
	}
	if got.FallbackAction != DefaultFallback(HeroClassArcher) {
		t.Errorf("FallbackAction = %v, want class default", got.FallbackAction)
	}
	if mock.calls != 2 {
		t.Errorf("classifier called %d times, want exactly 2 (one retry, no more)", mock.calls)
	}
	if err := ValidateIntentClassification(got, frostWarden()); err != nil {
		t.Errorf("fallback classification failed its own validation: %v", err)
	}
}
