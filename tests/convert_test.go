package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	type testCase struct {
		dir    string
		errors []string
	}
	tests := []testCase{
		{
			dir:    "convert_over_func",
			errors: []string{"convert should be defined over func"},
		},
		{
			dir:    "at_least_one_param",
			errors: []string{"function should have at least 1 parameter"},
		},
		{
			dir:    "at_least_one_result",
			errors: []string{"function should return at least 1 result"},
		},
		{
			dir:    "basic_types_not_supported",
			errors: []string{"basic type not supported", "int32"},
		},
		{
			dir:    "interface_not_allowed",
			errors: []string{"interface{ a() }", "unsupported type. supported: MyType, *MyType, []MyType, []*MyType"},
		},
		{
			dir:    "interface_not_allowed_2",
			errors: []string{"a Iface", "not struct"},
		},
		{
			dir:    "unnamed_parameter",
			errors: []string{"A", "parameter should have name"},
		},
		{
			dir:    "param_type_not_found",
			errors: []string{"a XYZ", "type not found"},
		},
		{
			dir:    "result_type_not_found",
			errors: []string{"a.B", "type not found"},
		},
		{
			dir:    "not_allowed_type",
			errors: []string{"a [][]A", "unsupported type. supported: MyType, *MyType, []MyType, []*MyType"},
		},
		{
			dir:    "not_allowed_type_2",
			errors: []string{"*[]B", "unsupported type. supported: MyType, *MyType, []MyType, []*MyType"},
		},
		{
			dir:    "not_allowed_type_3",
			errors: []string{"a []*[]A", "unsupported type. supported: MyType, *MyType, []MyType, []*MyType"},
		},
		{
			dir:    "not_allowed_type_4",
			errors: []string{"a []string", "basic type not supported"},
		},
		{
			dir:    "inconsistent_types",
			errors: []string{"a []A, *B", "parameter and return types should be both slices or both not slices"},
		},
		{
			dir:    "inconsistent_types_2",
			errors: []string{"a *A, []*B", "parameter and return types should be both slices or both not slices"},
		},
		{dir: "in_field_not_exists"},
		{dir: "not_exported_field"},
		{dir: "fields_added"},
		{dir: "fields_added_same_file"},
		{
			dir:    "out_struct_not_created_if_package",
			errors: []string{"*a.Out", "type not found"},
		},
		{dir: "out_struct_created"},
		{dir: "array"},
		{
			dir:    "bad_ast",
			errors: []string{"expected '}', found 'type'"},
		},
		{
			dir: "multiple_converts",
			errors: []string{
				"no //go:generate convert",
				"possible reason is multiple converts in same file. in this case rerun `go generate ..`",
			},
		},
	}

	wd, err := os.Getwd()
	require.NoError(t, err)

	_, err = exec.Command("go", "build", "-o", filepath.Join(wd, "convert"), filepath.Dir(wd)).Output()
	require.NoError(t, err)

	for _, test := range tests {
		t.Run(test.dir, func(t *testing.T) {
			err = RunTest(t, filepath.Join(wd, test.dir))

			if len(test.errors) == 0 {
				require.NoError(t, err)
			}

			for _, expectedErr := range test.errors {
				require.Contains(t, err.Error(), expectedErr)
			}
		})
	}

	err = os.Remove(filepath.Join(wd, "convert"))
	require.NoError(t, err)
}
