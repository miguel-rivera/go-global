package globalpayments

import "testing"

type MockTimeFormatter struct {
	counter int
	layout  string
}

func (mock *MockTimeFormatter) Format(layout string) string {
	mock.layout = layout
	mock.counter++
	return "20180614095601"
}

func Test_formatTimer(t *testing.T) {
	mock := &MockTimeFormatter{}

	returnedString := formatTime(mock, "20060102150405")

	if got, want := mock.counter, 1; got != want {
		t.Errorf("formatTime called  %v time(s), want %v time(s)", got, want)
	}

	if got, want := returnedString, "20180614095601"; got != want {
		t.Errorf("CardStorage formatTime is %v, want %v", got, want)
	}
}
