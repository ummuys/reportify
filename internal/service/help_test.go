package service

type mockErr string

func (e mockErr) Error() string { return string(e) }

func anyErr(s string) error { return mockErr(s) }
