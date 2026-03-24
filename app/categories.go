package app

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Category string

// Special Case Categories
const (
	CategorySettings    Category = "settings"
	CategorySecurity    Category = "security"
	CategoryDevelopment Category = "development"
)

const (
	CategoryCustomers     Category = "customers"
	CategoryBilling       Category = "billing"
	CategoryProduct       Category = "product"
	CategoryMarketing     Category = "marketing"
	CategoryCommunication Category = "communication"
	CategorySupport       Category = "support"
	CategoryReporting     Category = "reporting"
	CategoryDistribution  Category = "distribution"
	CategoryMonitoring    Category = "monitoring"
	CategoryManagement    Category = "management"
	CategoryOperations    Category = "operations"
	CategoryUtilities     Category = "utilities"
	CategoryIntegrations  Category = "integration"
	CategoryGeneral       Category = "general"
)

func (c Category) Name() string {
	return cases.Title(language.English, cases.Compact).String(strings.ReplaceAll(string(c), "-", " "))
}
