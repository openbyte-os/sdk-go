package app

import "testing"

func TestCategoryName(t *testing.T) {
	tests := []struct {
		name     string
		category Category
		want     string
	}{
		{name: "single word", category: CategoryCustomers, want: "Customers"},
		{name: "single word settings", category: CategorySettings, want: "Settings"},
		{name: "single word billing", category: CategoryBilling, want: "Billing"},
		{name: "hyphenated category", category: Category("my-category"), want: "My Category"},
		{name: "multiple hyphens", category: Category("a-b-c"), want: "A B C"},
		{name: "empty string", category: Category(""), want: ""},
		{name: "already capitalized input", category: Category("UPPER"), want: "Upper"},
		{name: "mixed case with hyphens", category: Category("my-Custom-app"), want: "My Custom App"},
		{name: "single character", category: Category("x"), want: "X"},
		{name: "trailing hyphen", category: Category("test-"), want: "Test "},
		{name: "leading hyphen", category: Category("-test"), want: " Test"},
		{name: "integration constant", category: CategoryIntegrations, want: "Integration"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.category.Name()
			if got != tt.want {
				t.Errorf("Category(%q).Name() = %q, want %q", string(tt.category), got, tt.want)
			}
		})
	}
}
