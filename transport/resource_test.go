package transport

import (
	"reflect"
	"testing"
)

func TestGetResources(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Resource
	}{
		{"id only", "7oPPnV1", []Resource{{ID: "7oPPnV1"}}},
		{"id and level", "abc.2", []Resource{{ID: "abc", Level: 2}}},
		{"id and display", "abc:Name", []Resource{{ID: "abc", Display: "Name"}}},
		{"id level display", "abc.3:Name", []Resource{{ID: "abc", Level: 3, Display: "Name"}}},
		{"multiple", "a.1:One,b.2:Two,c", []Resource{{ID: "a", Level: 1, Display: "One"}, {ID: "b", Level: 2, Display: "Two"}, {ID: "c"}}},
		{"empty input", "", []Resource{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetResources(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Fatalf("GetResources(%q) = %#v, expected %#v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestResource_ToHeader(t *testing.T) {
	tests := []struct {
		r   Resource
		exp string
	}{
		{Resource{ID: "x"}, "x"},
		{Resource{ID: "x", Level: 9}, "x.9"},
		{Resource{ID: "x", Display: "Disp"}, "x:Disp"},
		{Resource{ID: "x", Level: 4, Display: "Disp"}, "x.4:Disp"},
	}

	for _, tt := range tests {
		if got := tt.r.ToHeader(); got != tt.exp {
			t.Fatalf("Resource(%#v).ToHeader() = %q, expected %q", tt.r, got, tt.exp)
		}
	}
}

func TestChannelDepartments(t *testing.T) {
	in := map[string]string{"chanB": "dept1", "chanA": "dept2"}
	header := ChannelDepartmentsHeader(in)
	if header != "chanA=dept2,chanB=dept1" {
		t.Fatalf("ChannelDepartmentsHeader(%v) = %q", in, header)
	}
	if got := GetChannelDepartments(header); !reflect.DeepEqual(got, in) {
		t.Fatalf("GetChannelDepartments(%q) = %v, expected %v", header, got, in)
	}

	tests := []struct {
		input    string
		expected map[string]string
	}{
		{"", map[string]string{}},
		{"a=1,b,=2,c=", map[string]string{"a": "1"}},
	}
	for _, tt := range tests {
		if got := GetChannelDepartments(tt.input); !reflect.DeepEqual(got, tt.expected) {
			t.Fatalf("GetChannelDepartments(%q) = %v, expected %v", tt.input, got, tt.expected)
		}
	}
}
