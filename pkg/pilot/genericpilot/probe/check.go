package probe

// Check is a function that can return an error to signal failure
type Check func() error

// CombineChecks performs a logical AND on a list of Checks
func CombineChecks(checks ...Check) Check {
	return func() error {
		for _, check := range checks {
			if err := check(); err != nil {
				return err
			}
		}
		return nil
	}
}
