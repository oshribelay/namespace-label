package resources

import (
	"fmt"

	"github.com/oshribelay/namespace-label/internal/controller/utils"
)

func ValidateNamespaceLabel(labels, protectedPrefixes map[string]string) error {
	for key := range labels {
		if utils.IsReservedLabel(key, protectedPrefixes) {
			return fmt.Errorf("Invalid label: reserved label cannot be modified: %s", key)
		}
	}
	return nil
}
