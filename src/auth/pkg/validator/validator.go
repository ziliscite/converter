package validator

import "sync"

type Validator struct {
	errs map[string]string
	mu   sync.Mutex
}

func New() *Validator {
	return &Validator{
		errs: make(map[string]string),
		mu:   sync.Mutex{},
	}
}

func (v *Validator) Valid() bool {
	return len(v.errs) == 0
}

func (v *Validator) AddError(key, message string) {
	v.mu.Lock()
	if _, exists := v.errs[key]; !exists {
		v.errs[key] = message
	}
	v.mu.Unlock()
}

// Errors return the map of errors and clear the internal map
func (v *Validator) Errors() map[string]string {
	defer func() {
		v.mu.Lock()
		v.errs = make(map[string]string)
		v.mu.Unlock()
	}()

	return v.errs
}
