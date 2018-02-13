package net

type FatalError struct {
	e error
}

func (f *FatalError) Error() string {
	return f.e.Error()
}
