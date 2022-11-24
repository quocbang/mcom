package mcom

// Request definition.
type Request interface {
	// CheckInsufficiency returns Code_INSUFFICIENT_REQUEST while missing required parameters.
	//
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	// todo: rename this method.
	CheckInsufficiency() error
}
