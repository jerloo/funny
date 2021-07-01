package funny

import (
	"fmt"
)

// P panic
func P(keyword string, pos Position) error {
	return fmt.Errorf("funny error [%s] at position %s", keyword, pos.String())
}
