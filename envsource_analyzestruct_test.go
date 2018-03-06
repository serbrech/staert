package staert

import (
	"os"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/containous/flaeg/parse"
	"github.com/stretchr/testify/require"
)

func setupEnv(env map[string]string) {
	for k, v := range env {
		os.Setenv(k, v)
	}

}
func cleanupEnv(env map[string]string) {
	for k := range env {
		os.Unsetenv(k)
	}
}

type basicAppConfig struct {
	StringValue string
	IntValue    int
	BoolValue   bool
}

type typeInterface interface {
	Foo() string
}

type delegatorType struct {
	typeInterface
	IntValue    int
	StringValue string
}

type sortableEnvValues []*envValue

func (s sortableEnvValues) Len() int {
	return len(s)
}

func (s sortableEnvValues) Less(i, j int) bool {
	return s[i].StrValue < s[j].StrValue
}

func (s sortableEnvValues) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type testAnalyzeStructThenHook func(t *testing.T, expectation, result sortableEnvValues, err error)

func testAnalyzeStructShouldSucceed(t *testing.T, expectation, result sortableEnvValues, err error) {
	require := require.New(t)

	require.NoError(err)
	require.Lenf(result, len(expectation), "Unexpected count of values returned")

	// Sort by value, according to StrValue (which might not be the best
	// idea ever), in order to ensure index based comparison consistency
	sort.Sort(expectation)
	sort.Sort(result)

	for i, v := range expectation {
		require.Equal(v.StrValue, result[i].StrValue)
		require.Exactly(v.Path, result[i].Path)
	}
}

func testAnalyzeStructShouldFail(t *testing.T, expectation, result sortableEnvValues, err error) {
	require.Error(t, err)
}

func TestAnalyzeStruct(t *testing.T) {
	subject := &envSource{"", "_", map[reflect.Type]parse.Parser{}}

	testCases := []struct {
		Label       string
		Source      interface{}
		Expectation []*envValue
		Env         map[string]string
		Then        testAnalyzeStructThenHook
	}{
		{
			Label:  "WithBasicConfiguration",
			Source: &basicAppConfig{},
			Expectation: []*envValue{
				{"FOOO", path{"StringValue"}},
				{"10", path{"IntValue"}},
				{"true", path{"BoolValue"}},
			},
			Env: map[string]string{
				"STRING_VALUE": "FOOO",
				"INT_VALUE":    "10",
				"BOOL_VALUE":   "true",
			},
			Then: testAnalyzeStructShouldSucceed,
		},
		{
			Label: "WithUnexportedFields",
			Source: &struct {
				unexported string
				IntValue   int
			}{},
			Expectation: []*envValue{
				{"10", path{"IntValue"}},
			},
			Env: map[string]string{
				"UNEXPORTED": "FOOO",
				"INT_VALUE":  "10",
			},
			Then: testAnalyzeStructShouldSucceed,
		},
		{
			Label: "WithEmbeddedStruct",
			Source: &struct {
				basicAppConfig
				FloatValue float32
			}{},
			Expectation: []*envValue{
				{"FOOO", path{"StringValue"}},
				{"10", path{"IntValue"}},
				{"true", path{"BoolValue"}},
				{"42.1", path{"FloatValue"}},
			},
			Env: map[string]string{
				"STRING_VALUE": "FOOO",
				"INT_VALUE":    "10",
				"BOOL_VALUE":   "true",
				"FLOAT_VALUE":  "42.1",
			},
			Then: testAnalyzeStructShouldSucceed,
		},
		{
			Label: "WithNestedStructValue",
			Source: &struct {
				Config basicAppConfig
			}{},
			Expectation: []*envValue{
				{"FOOO", path{"Config", "StringValue"}},
				{"10", path{"Config", "IntValue"}},
				{"true", path{"Config", "BoolValue"}},
			},
			Env: map[string]string{
				"CONFIG_STRING_VALUE": "FOOO",
				"CONFIG_INT_VALUE":    "10",
				"CONFIG_BOOL_VALUE":   "true",
			},
			Then: testAnalyzeStructShouldSucceed,
		},
		{
			Label: "WithDoubleNestedStructValue",
			Source: &struct {
				Nested struct {
					Config basicAppConfig
				}
			}{},
			Expectation: []*envValue{
				{"FOOO", path{"Nested", "Config", "StringValue"}},
				{"10", path{"Nested", "Config", "IntValue"}},
				{"true", path{"Nested", "Config", "BoolValue"}},
			},
			Env: map[string]string{
				"NESTED_CONFIG_STRING_VALUE": "FOOO",
				"NESTED_CONFIG_INT_VALUE":    "10",
				"NESTED_CONFIG_BOOL_VALUE":   "true",
			},
			Then: testAnalyzeStructShouldSucceed,
		},
		{
			Label: "WithNestedStructPtr",
			Source: &struct {
				Config *basicAppConfig
			}{},
			Expectation: []*envValue{
				{"FOOO", path{"Config", "StringValue"}},
				{"10", path{"Config", "IntValue"}},
				{"true", path{"Config", "BoolValue"}},
			},
			Env: map[string]string{
				"CONFIG_STRING_VALUE": "FOOO",
				"CONFIG_INT_VALUE":    "10",
				"CONFIG_BOOL_VALUE":   "true",
			},
			Then: testAnalyzeStructShouldSucceed,
		},
		{
			Label: "WithDoubleNestedStructPtr",
			Source: &struct {
				Nested *struct {
					Config *basicAppConfig
				}
			}{},
			Expectation: []*envValue{
				{"FOOO", path{"Nested", "Config", "StringValue"}},
				{"10", path{"Nested", "Config", "IntValue"}},
				{"true", path{"Nested", "Config", "BoolValue"}},
			},
			Env: map[string]string{
				"NESTED_CONFIG_STRING_VALUE": "FOOO",
				"NESTED_CONFIG_INT_VALUE":    "10",
				"NESTED_CONFIG_BOOL_VALUE":   "true",
			},
			Then: testAnalyzeStructShouldSucceed,
		},
		{
			Label: "WithDoubleNestedStructMixed",
			Source: &struct {
				Nested *struct {
					Config basicAppConfig
				}
			}{},
			Expectation: []*envValue{
				{"FOOO", path{"Nested", "Config", "StringValue"}},
				{"10", path{"Nested", "Config", "IntValue"}},
				{"true", path{"Nested", "Config", "BoolValue"}},
			},
			Env: map[string]string{
				"NESTED_CONFIG_STRING_VALUE": "FOOO",
				"NESTED_CONFIG_INT_VALUE":    "10",
				"NESTED_CONFIG_BOOL_VALUE":   "true",
			},
			Then: testAnalyzeStructShouldSucceed,
		},
		{
			Label: "WithPtrValue",
			Source: &struct {
				IntValue *int
			}{},
			Expectation: []*envValue{
				{"10", path{"IntValue"}},
			},
			Env: map[string]string{
				"INT_VALUE": "10",
			},
			Then: testAnalyzeStructShouldSucceed,
		},
		{
			Label: "WithNestedPtrValue",
			Source: &struct {
				Config struct {
					IntValue *int
				}
			}{},
			Expectation: []*envValue{
				{"10", path{"Config", "IntValue"}},
			},
			Env: map[string]string{
				"CONFIG_INT_VALUE": "10",
			},
			Then: testAnalyzeStructShouldSucceed,
		},
		{
			Label: "WithNestedPtrValue",
			Source: &struct {
				Config struct {
					IntValue *int
				}
			}{},
			Expectation: []*envValue{
				{"10", path{"Config", "IntValue"}},
			},
			Env: map[string]string{
				"CONFIG_INT_VALUE": "10",
			},
			Then: testAnalyzeStructShouldSucceed,
		},
		{
			Label: "WithPtrPtrToValue",
			Source: &struct {
				Config **int
			}{},
			Expectation: []*envValue{
				{"10", path{"Config"}},
			},
			Env: map[string]string{
				"CONFIG": "10",
			},
			Then: testAnalyzeStructShouldSucceed,
		},
		{
			Label: "WithPtrPtrToStruct",
			Source: &struct {
				Config **basicAppConfig
			}{},
			Expectation: []*envValue{
				{"FOOO", path{"Config", "StringValue"}},
				{"10", path{"Config", "IntValue"}},
				{"true", path{"Config", "BoolValue"}},
			},
			Env: map[string]string{
				"CONFIG_STRING_VALUE": "FOOO",
				"CONFIG_INT_VALUE":    "10",
				"CONFIG_BOOL_VALUE":   "true",
			},
			Then: testAnalyzeStructShouldSucceed,
		},
		{
			Label:  "WithInterfaceDelegation",
			Source: &delegatorType{},
			Expectation: []*envValue{
				{"FOOO", path{"StringValue"}},
				{"10", path{"IntValue"}},
			},
			Env: map[string]string{
				"STRING_VALUE": "FOOO",
				"INT_VALUE":    "10",
			},
			Then: testAnalyzeStructShouldSucceed,
		},
		{
			Label: "WithMapOfValues",
			Source: &struct {
				Config map[string]string
			}{},
			Expectation: []*envValue{
				{"FOO", path{"Config", "foo"}},
				{"MEH", path{"Config", "bar"}},
				{"BAR", path{"Config", "biz"}},
			},
			Env: map[string]string{
				"CONFIG_FOO": "FOO",
				"CONFIG_BAR": "MEH",
				"CONFIG_BIZ": "BAR",
			},
			Then: testAnalyzeStructShouldSucceed,
		},
		{
			Label: "WithMapOfStructValues",
			Source: &struct {
				Config map[string]basicAppConfig
			}{},
			Expectation: []*envValue{
				{"FOO", path{"Config", "foo", "StringValue"}},
				{"MEH", path{"Config", "bar", "StringValue"}},
				{"BAR", path{"Config", "biz", "StringValue"}},
			},
			Env: map[string]string{
				"CONFIG_FOO_STRING_VALUE": "FOO",
				"CONFIG_BAR_STRING_VALUE": "MEH",
				"CONFIG_BIZ_STRING_VALUE": "BAR",
			},
			Then: testAnalyzeStructShouldSucceed,
		},
		{
			Label: "WithMapOfStructPtr",
			Source: &struct {
				Config map[string]*basicAppConfig
			}{},
			Expectation: []*envValue{
				{"FOO", path{"Config", "foo", "StringValue"}},
				{"MEH", path{"Config", "bar", "StringValue"}},
				{"BAR", path{"Config", "biz", "StringValue"}},
			},
			Env: map[string]string{
				"CONFIG_FOO_STRING_VALUE": "FOO",
				"CONFIG_BAR_STRING_VALUE": "MEH",
				"CONFIG_BIZ_STRING_VALUE": "BAR",
			},
			Then: testAnalyzeStructShouldSucceed,
		},
		{
			Label: "WithMapOfMapOfPtrStruct",
			Source: &struct {
				Config map[int]map[string]*basicAppConfig
			}{},
			Expectation: []*envValue{
				{"FOO", path{"Config", "0", "foo", "StringValue"}},
				{"MEH", path{"Config", "1", "bar", "StringValue"}},
				{"BAR", path{"Config", "0", "biz", "StringValue"}},
			},
			Env: map[string]string{
				"CONFIG_0_FOO_STRING_VALUE": "FOO",
				"CONFIG_1_BAR_STRING_VALUE": "MEH",
				"CONFIG_0_BIZ_STRING_VALUE": "BAR",
			},
			Then: testAnalyzeStructShouldSucceed,
		},
		{
			Label: "WithSliceToValue",
			Source: &struct {
				Config []int
			}{},
			Expectation: []*envValue{
				{"FOOO", path{"Config", "0"}},
				{"10", path{"Config", "1"}},
				{"true", path{"Config", "2"}},
			},
			Env: map[string]string{
				"CONFIG_0": "FOOO",
				"CONFIG_1": "10",
				"CONFIG_2": "true",
			},
			Then: testAnalyzeStructShouldSucceed,
		},
		{
			Label: "WithSliceToValueAndInvalidKey",
			Source: &struct {
				Config []int
			}{},
			Expectation: []*envValue{},
			Env: map[string]string{
				"CONFIG_0":      "FOOO",
				"CONFIG_1":      "10",
				"CONFIG_PATATE": "true",
			},
			Then: testAnalyzeStructShouldFail,
		},
		{
			Label: "WithAnArrayToValue",
			Source: &struct {
				Config [10]int
			}{},
			Expectation: []*envValue{
				{"FOOO", path{"Config", "0"}},
				{"10", path{"Config", "1"}},
				{"true", path{"Config", "2"}},
			},
			Env: map[string]string{
				"CONFIG_0": "FOOO",
				"CONFIG_1": "10",
				"CONFIG_2": "true",
			},
			Then: testAnalyzeStructShouldSucceed,
		},
		{
			Label: "WithAnArrayAndAnOutOfBoundIndex",
			Source: &struct {
				Config [10]int
			}{},
			Expectation: []*envValue{},
			Env: map[string]string{
				"CONFIG_11": "10",
			},
			Then: testAnalyzeStructShouldFail,
		},
		{
			Label: "WithAnArrayToValue",
			Source: &struct {
				Config [10]int
			}{},
			Expectation: []*envValue{
				{"FOOO", path{"Config", "0"}},
				{"10", path{"Config", "1"}},
				{"true", path{"Config", "2"}},
			},
			Env: map[string]string{
				"CONFIG_0": "FOOO",
				"CONFIG_1": "10",
				"CONFIG_2": "true",
			},
			Then: testAnalyzeStructShouldSucceed,
		},
		{
			Label: "WithASliceToStruct",
			Source: &struct {
				Config []basicAppConfig
			}{},
			Expectation: []*envValue{
				{"FOOO", path{"Config", "0", "StringValue"}},
				{"10", path{"Config", "0", "IntValue"}},
				{"MIMI", path{"Config", "1", "StringValue"}},
				{"15", path{"Config", "1", "IntValue"}},
			},
			Env: map[string]string{
				"CONFIG_0_STRING_VALUE": "FOOO",
				"CONFIG_0_INT_VALUE":    "10",
				"CONFIG_1_STRING_VALUE": "MIMI",
				"CONFIG_1_INT_VALUE":    "15",
			},
			Then: testAnalyzeStructShouldSucceed,
		},
		{
			Label: "WithASliceToASliceToStruct",
			Source: &struct {
				Config [][]basicAppConfig
			}{},
			Expectation: []*envValue{
				{"FOOO", path{"Config", "0", "0", "StringValue"}},
				{"10", path{"Config", "0", "0", "IntValue"}},
				{"MIMI", path{"Config", "1", "1", "StringValue"}},
				{"15", path{"Config", "1", "1", "IntValue"}},
			},
			Env: map[string]string{
				"CONFIG_0_0_STRING_VALUE": "FOOO",
				"CONFIG_0_0_INT_VALUE":    "10",
				"CONFIG_1_1_STRING_VALUE": "MIMI",
				"CONFIG_1_1_INT_VALUE":    "15",
			},
			Then: testAnalyzeStructShouldSucceed,
		},
		{
			Label: "WithASliceToAMapToStruct",
			Source: &struct {
				Config []map[string]basicAppConfig
			}{},
			Expectation: []*envValue{
				{"FOOO", path{"Config", "0", "foo", "StringValue"}},
				{"10", path{"Config", "0", "foo", "IntValue"}},
				{"MIMI", path{"Config", "1", "bar", "StringValue"}},
				{"15", path{"Config", "1", "bar", "IntValue"}},
			},
			Env: map[string]string{
				"CONFIG_0_FOO_STRING_VALUE": "FOOO",
				"CONFIG_0_FOO_INT_VALUE":    "10",
				"CONFIG_1_BAR_STRING_VALUE": "MIMI",
				"CONFIG_1_BAR_INT_VALUE":    "15",
			},
			Then: testAnalyzeStructShouldSucceed,
		},
		{
			Label: "WithAFunc",
			Source: &struct {
				Config basicAppConfig
				Time   func() time.Time
			}{}, Expectation: []*envValue{
				{"FOOO", path{"Config", "StringValue"}},
				{"10", path{"Config", "IntValue"}},
				{"true", path{"Config", "BoolValue"}},
			},
			Env: map[string]string{
				"CONFIG_STRING_VALUE": "FOOO",
				"CONFIG_INT_VALUE":    "10",
				"CONFIG_BOOL_VALUE":   "true",
			},
			Then: testAnalyzeStructShouldSucceed,
		},
		{
			Label: "WithAWebBasicAuth",
			Source: &struct {
				Basic     *Basic
				UsersFile string
			}{}, Expectation: []*envValue{
				{"UserZero", path{"Basic", "0"}},
				{"UserOne", path{"Basic", "1"}},
				{"path/to/file", path{"UsersFile"}},
			},
			Env: map[string]string{
				"BASIC_0":    "UserZero",
				"BASIC_1":    "UserOne",
				"USERS_FILE": "path/to/file",
			},
			Then: testAnalyzeStructShouldSucceed,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Label, func(t *testing.T) {
			setupEnv(testCase.Env)
			res, err := subject.analyzeStruct(
				reflect.TypeOf(testCase.Source).Elem(),
				path{},
			)
			testCase.Then(t, testCase.Expectation, res, err)
			cleanupEnv(testCase.Env)
		})
	}

}

func TestEnvVarFromPath(t *testing.T) {
	testCases := []struct {
		Label       string
		Prefix      string
		Separator   string
		Path        []string
		Expectation string
	}{
		{"BlankPrefix", "", "_", []string{"Foo"}, "FOO"},
		{"NonBlankPrefix", "YOUPI", "_", []string{"Foo"}, "YOUPI_FOO"},
		{
			"CamelCasedPathMembers",
			"YOUPI",
			"_",
			[]string{"Foo", "IamGroot", "IAmBatman"},
			"YOUPI_FOO_IAM_GROOT_I_AM_BATMAN",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Label, func(t *testing.T) {
			subject := &envSource{
				testCase.Prefix,
				testCase.Separator,
				map[reflect.Type]parse.Parser{},
			}

			result := subject.envVarFromPath(testCase.Path)
			require.Exactly(t, testCase.Expectation, result)
		})
	}

}

// Users authentication users
type Users []string

type Basic struct {
	Users     `mapstructure:","`
	UsersFile string
}

type testcase struct {
	Source      interface{}
	Expectation []*envValue
	Env         map[string]string
	Then        testAnalyzeStructThenHook
}

func TestWithArray(t *testing.T) {

	sourceConfig := struct {
		StringArray []string
	}{}
	testCase := struct {
		Source      interface{}
		Expectation []*envValue
		Env         map[string]string
	}{
		Source: &sourceConfig,
		Expectation: []*envValue{
			{"one", path{"StringArray", "0"}},
			{"two", path{"StringArray", "1"}},
		},
		Env: map[string]string{
			"STRING_ARRAY_0": "one",
			"STRING_ARRAY_1": "two",
		},
	}

	setupEnv(testCase.Env)
	parsers, _ := parse.LoadParsers(nil)
	subject := &envSource{"", "_", parsers}
	res, _ := subject.analyzeStruct(reflect.TypeOf(testCase.Source).Elem(), path{})
	subject.assignValues(reflect.ValueOf(&sourceConfig), res, []string{})

	require.ElementsMatch(t, []string{"one", "two"}, sourceConfig.StringArray)
	cleanupEnv(testCase.Env)
}

func TestWithSliceToValue(t *testing.T) {
	config := struct {
		Config []int
	}{}
	testCase := testcase{
		Source: &config,
		Expectation: []*envValue{
			{"FOOO", path{"Config", "0"}},
			{"10", path{"Config", "1"}},
			{"true", path{"Config", "2"}},
		},
		Env: map[string]string{
			"CONFIG_0": "FOOO",
			"CONFIG_1": "10",
			"CONFIG_2": "true",
		},
		Then: testAnalyzeStructShouldSucceed,
	}

	setupEnv(testCase.Env)

	parsers, _ := parse.LoadParsers(nil)
	subject := &envSource{"", "_", parsers}
	res, err := subject.analyzeStruct(reflect.TypeOf(testCase.Source).Elem(), path{})
	testCase.Then(t, testCase.Expectation, res, err)
	configVal := reflect.ValueOf(&config)
	subject.assignValues(configVal, res, []string{})

	cleanupEnv(testCase.Env)
}

func TestNextLevelKeys(t *testing.T) {
	subject := &envSource{"", "_", map[reflect.Type]parse.Parser{}}
	testCases := []struct {
		Label       string
		Prefix      string
		Env         []string
		Expectation []string
	}{
		{
			Label:  "WithPrefix",
			Prefix: "CONFIG_APP",
			Env: []string{
				"CONFIG_APP_BATMAN_FOO",
				"CONFIG_APP_ROBIN_FOO",
				"CONFIG_APP_JOCKER_FOO",
			},
			Expectation: []string{
				"CONFIG_APP_BATMAN",
				"CONFIG_APP_ROBIN",
				"CONFIG_APP_JOCKER",
			},
		},
		{
			Label:  "WithDuplicates",
			Prefix: "CONFIG_APP",
			Env: []string{
				"CONFIG_APP_BATMAN_FOO",
				"CONFIG_APP_ROBIN_FOO",
				"CONFIG_APP_JOCKER_FOO",
				"CONFIG_APP_BATMAN_BAR",
			},
			Expectation: []string{
				"CONFIG_APP_BATMAN",
				"CONFIG_APP_ROBIN",
				"CONFIG_APP_JOCKER",
				"CONFIG_APP_BATMAN",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Label, func(t *testing.T) {
			res := subject.nextLevelKeys(testCase.Prefix, testCase.Env)
			require.ElementsMatch(t, testCase.Expectation, res)
		})
	}
}

func TestEnvVarsWithPrefix(t *testing.T) {

	subject := &envSource{"", "_", map[reflect.Type]parse.Parser{}}

	testCases := []struct {
		Label       string
		Prefix      string
		Env         map[string]string
		Expectation []string
	}{
		{
			Label:  "WithPrefix",
			Prefix: "STAERT_APP",
			Env: map[string]string{
				"STRING_VALUE":          "FOOO",
				"INT_VALUE":             "10",
				"BOOL_VALUE":            "true",
				"STAERT_APP_BOOL_VALUE": "true",
				"STAERT_APP_BAR_VALUE":  "true",
			},
			Expectation: []string{"STAERT_APP_BAR_VALUE", "STAERT_APP_BOOL_VALUE"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Label, func(t *testing.T) {
			setupEnv(testCase.Env)
			res := subject.envVarsWithPrefix(testCase.Prefix)
			require.ElementsMatch(t, testCase.Expectation, res)
			cleanupEnv(testCase.Env)
		})
	}
}

func TestUnique(t *testing.T) {
	testCases := []struct {
		Label       string
		In          []string
		Expectation []string
	}{
		{
			Label:       "WithDuplicates",
			In:          []string{"FOO", "BAR", "BIZ", "FOO", "BIZ"},
			Expectation: []string{"FOO", "BAR", "BIZ"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Label, func(t *testing.T) {
			res := unique(testCase.In)
			require.ElementsMatch(t, testCase.Expectation, res)
		})
	}
}

func TestKeyFromEnvVar(t *testing.T) {
	subject := &envSource{"", "_", map[reflect.Type]parse.Parser{}}
	testCases := []struct {
		Label       string
		Prefix      string
		EnvVar      string
		Expectation string
	}{
		{"WithPrefix", "CONFIG_APP", "CONFIG_APP_BATMAN", "batman"},
		{"WithPrefixAndSuffix", "CONFIG_APP", "CONFIG_APP_BATMAN_FOO", "batman"},
		{"WithoutPrefix", "", "BATMAN", "batman"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Label, func(t *testing.T) {
			res := subject.keyFromEnvVar(testCase.EnvVar, testCase.Prefix)
			require.Equal(t, testCase.Expectation, res)
		})
	}
}

// Dummy string parser, to enable test writing
type testStringParser string

func (s testStringParser) String() string {
	return string(s)
}

func (s *testStringParser) Set(val string) error {
	*s = testStringParser(val)
	return nil
}

func (s testStringParser) SetValue(val interface{}) {}

func (s testStringParser) Get() interface{} {
	return nil
}
