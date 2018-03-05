package staert

import (
	"os"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/containous/flaeg/parse"
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
	if err != nil {
		t.Logf("Weren't expecting an error, got [%v]", err)
		t.FailNow()
	}

	if len(expectation) != len(result) {
		t.Logf("Unexpected count of values returned: Expected [%d] got [%d]", len(expectation), len(result))
		t.FailNow()
	}

	// Sort by value, according to StrValue (which might not be the best
	// idea ever), in order to ensure index based comparison consistency
	sort.Sort(expectation)
	sort.Sort(result)

	for i, v := range expectation {
		if v.StrValue != result[i].StrValue {
			t.Logf("Expected [%v] got [%v]", *v, *result[i])
			t.Fail()
		}

		if len(v.Path) != len(result[i].Path) {
			t.Logf("Expected Path length of [%v] got [%v]", len(v.Path), len(result[i].Path))
			t.FailNow()
		}

		for j, p := range v.Path {
			if p != result[i].Path[j] {
				t.Logf("Expected path term [%v] got [%v]", p, result[i].Path[j])
				t.Fail()
			}
		}

	}
}

func testAnalyzeStructShouldFail(t *testing.T, expectation, result sortableEnvValues, err error) {
	if err == nil {
		t.Logf("Expected an error, got nothing")
		t.Fail()
	}
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
			"WithBasicConfiguration",
			&basicAppConfig{},
			[]*envValue{
				{"FOOO", path{"StringValue"}},
				{"10", path{"IntValue"}},
				{"true", path{"BoolValue"}},
			},
			map[string]string{
				"STRING_VALUE": "FOOO",
				"INT_VALUE":    "10",
				"BOOL_VALUE":   "true",
			},
			testAnalyzeStructShouldSucceed,
		},
		{
			"WithUnexportedFields",
			&struct {
				unexported string
				IntValue   int
			}{},
			[]*envValue{
				{"10", path{"IntValue"}},
			},
			map[string]string{
				"UNEXPORTED": "FOOO",
				"INT_VALUE":  "10",
			},
			testAnalyzeStructShouldSucceed,
		},
		{
			"WithEmbeddedStruct",
			&struct {
				basicAppConfig
				FloatValue float32
			}{},
			[]*envValue{
				{"FOOO", path{"StringValue"}},
				{"10", path{"IntValue"}},
				{"true", path{"BoolValue"}},
				{"42.1", path{"FloatValue"}},
			},
			map[string]string{
				"STRING_VALUE": "FOOO",
				"INT_VALUE":    "10",
				"BOOL_VALUE":   "true",
				"FLOAT_VALUE":  "42.1",
			},
			testAnalyzeStructShouldSucceed,
		},
		{
			"WithNestedStructValue",
			&struct {
				Config basicAppConfig
			}{},
			[]*envValue{
				{"FOOO", path{"Config", "StringValue"}},
				{"10", path{"Config", "IntValue"}},
				{"true", path{"Config", "BoolValue"}},
			},
			map[string]string{
				"CONFIG_STRING_VALUE": "FOOO",
				"CONFIG_INT_VALUE":    "10",
				"CONFIG_BOOL_VALUE":   "true",
			},
			testAnalyzeStructShouldSucceed,
		},
		{
			"WithDoubleNestedStructValue",
			&struct {
				Nested struct {
					Config basicAppConfig
				}
			}{},
			[]*envValue{
				{"FOOO", path{"Nested", "Config", "StringValue"}},
				{"10", path{"Nested", "Config", "IntValue"}},
				{"true", path{"Nested", "Config", "BoolValue"}},
			},
			map[string]string{
				"NESTED_CONFIG_STRING_VALUE": "FOOO",
				"NESTED_CONFIG_INT_VALUE":    "10",
				"NESTED_CONFIG_BOOL_VALUE":   "true",
			},
			testAnalyzeStructShouldSucceed,
		},
		{
			"WithNestedStructPtr",
			&struct {
				Config *basicAppConfig
			}{},
			[]*envValue{
				{"FOOO", path{"Config", "StringValue"}},
				{"10", path{"Config", "IntValue"}},
				{"true", path{"Config", "BoolValue"}},
			},
			map[string]string{
				"CONFIG_STRING_VALUE": "FOOO",
				"CONFIG_INT_VALUE":    "10",
				"CONFIG_BOOL_VALUE":   "true",
			},
			testAnalyzeStructShouldSucceed,
		},
		{
			"WithDoubleNestedStructPtr",
			&struct {
				Nested *struct {
					Config *basicAppConfig
				}
			}{},
			[]*envValue{
				{"FOOO", path{"Nested", "Config", "StringValue"}},
				{"10", path{"Nested", "Config", "IntValue"}},
				{"true", path{"Nested", "Config", "BoolValue"}},
			},
			map[string]string{
				"NESTED_CONFIG_STRING_VALUE": "FOOO",
				"NESTED_CONFIG_INT_VALUE":    "10",
				"NESTED_CONFIG_BOOL_VALUE":   "true",
			},
			testAnalyzeStructShouldSucceed,
		},
		{
			"WithDoubleNestedStructMixed",
			&struct {
				Nested *struct {
					Config basicAppConfig
				}
			}{},
			[]*envValue{
				{"FOOO", path{"Nested", "Config", "StringValue"}},
				{"10", path{"Nested", "Config", "IntValue"}},
				{"true", path{"Nested", "Config", "BoolValue"}},
			},
			map[string]string{
				"NESTED_CONFIG_STRING_VALUE": "FOOO",
				"NESTED_CONFIG_INT_VALUE":    "10",
				"NESTED_CONFIG_BOOL_VALUE":   "true",
			},
			testAnalyzeStructShouldSucceed,
		},
		{
			"WithPtrValue",
			&struct {
				IntValue *int
			}{},
			[]*envValue{
				{"10", path{"IntValue"}},
			},
			map[string]string{
				"INT_VALUE": "10",
			},
			testAnalyzeStructShouldSucceed,
		},
		{
			"WithNestedPtrValue",
			&struct {
				Config struct {
					IntValue *int
				}
			}{},
			[]*envValue{
				{"10", path{"Config", "IntValue"}},
			},
			map[string]string{
				"CONFIG_INT_VALUE": "10",
			},
			testAnalyzeStructShouldSucceed,
		},
		{
			"WithNestedPtrValue",
			&struct {
				Config struct {
					IntValue *int
				}
			}{},
			[]*envValue{
				{"10", path{"Config", "IntValue"}},
			},
			map[string]string{
				"CONFIG_INT_VALUE": "10",
			},
			testAnalyzeStructShouldSucceed,
		},
		{
			"WithPtrPtrToValue",
			&struct {
				Config **int
			}{},
			[]*envValue{
				{"10", path{"Config"}},
			},
			map[string]string{
				"CONFIG": "10",
			},
			testAnalyzeStructShouldSucceed,
		},
		{
			"WithPtrPtrToStruct",
			&struct {
				Config **basicAppConfig
			}{},
			[]*envValue{
				{"FOOO", path{"Config", "StringValue"}},
				{"10", path{"Config", "IntValue"}},
				{"true", path{"Config", "BoolValue"}},
			},
			map[string]string{
				"CONFIG_STRING_VALUE": "FOOO",
				"CONFIG_INT_VALUE":    "10",
				"CONFIG_BOOL_VALUE":   "true",
			},
			testAnalyzeStructShouldSucceed,
		},
		{
			"WithInterfaceDelegation",
			&delegatorType{},
			[]*envValue{
				{"FOOO", path{"StringValue"}},
				{"10", path{"IntValue"}},
			},
			map[string]string{
				"STRING_VALUE": "FOOO",
				"INT_VALUE":    "10",
			},
			testAnalyzeStructShouldSucceed,
		},
		{
			"WithMapOfValues",
			&struct {
				Config map[string]string
			}{},
			[]*envValue{
				{"FOO", path{"Config", "foo"}},
				{"MEH", path{"Config", "bar"}},
				{"BAR", path{"Config", "biz"}},
			},
			map[string]string{
				"CONFIG_FOO": "FOO",
				"CONFIG_BAR": "MEH",
				"CONFIG_BIZ": "BAR",
			},
			testAnalyzeStructShouldSucceed,
		},
		{
			"WithMapOfStructValues",
			&struct {
				Config map[string]basicAppConfig
			}{},
			[]*envValue{
				{"FOO", path{"Config", "foo", "StringValue"}},
				{"MEH", path{"Config", "bar", "StringValue"}},
				{"BAR", path{"Config", "biz", "StringValue"}},
			},
			map[string]string{
				"CONFIG_FOO_STRING_VALUE": "FOO",
				"CONFIG_BAR_STRING_VALUE": "MEH",
				"CONFIG_BIZ_STRING_VALUE": "BAR",
			},
			testAnalyzeStructShouldSucceed,
		},
		{
			"WithMapOfStructPtr",
			&struct {
				Config map[string]*basicAppConfig
			}{},
			[]*envValue{
				{"FOO", path{"Config", "foo", "StringValue"}},
				{"MEH", path{"Config", "bar", "StringValue"}},
				{"BAR", path{"Config", "biz", "StringValue"}},
			},
			map[string]string{
				"CONFIG_FOO_STRING_VALUE": "FOO",
				"CONFIG_BAR_STRING_VALUE": "MEH",
				"CONFIG_BIZ_STRING_VALUE": "BAR",
			},
			testAnalyzeStructShouldSucceed,
		},
		{
			"WithMapOfMapOfPtrStruct",
			&struct {
				Config map[int]map[string]*basicAppConfig
			}{},
			[]*envValue{
				{"FOO", path{"Config", "0", "foo", "StringValue"}},
				{"MEH", path{"Config", "1", "bar", "StringValue"}},
				{"BAR", path{"Config", "0", "biz", "StringValue"}},
			},
			map[string]string{
				"CONFIG_0_FOO_STRING_VALUE": "FOO",
				"CONFIG_1_BAR_STRING_VALUE": "MEH",
				"CONFIG_0_BIZ_STRING_VALUE": "BAR",
			},
			testAnalyzeStructShouldSucceed,
		},
		{
			"WithSliceToValue",
			&struct {
				Config []int
			}{},
			[]*envValue{
				{"FOOO", path{"Config", "0"}},
				{"10", path{"Config", "1"}},
				{"true", path{"Config", "2"}},
			},
			map[string]string{
				"CONFIG_0": "FOOO",
				"CONFIG_1": "10",
				"CONFIG_2": "true",
			},
			testAnalyzeStructShouldSucceed,
		},
		{
			"WithSliceToValueAndInvalidKey",
			&struct {
				Config []int
			}{},
			[]*envValue{},
			map[string]string{
				"CONFIG_0":      "FOOO",
				"CONFIG_1":      "10",
				"CONFIG_PATATE": "true",
			},
			testAnalyzeStructShouldFail,
		},
		{
			"WithAnArrayToValue",
			&struct {
				Config [10]int
			}{},
			[]*envValue{
				{"FOOO", path{"Config", "0"}},
				{"10", path{"Config", "1"}},
				{"true", path{"Config", "2"}},
			},
			map[string]string{
				"CONFIG_0": "FOOO",
				"CONFIG_1": "10",
				"CONFIG_2": "true",
			},
			testAnalyzeStructShouldSucceed,
		},
		{
			"WithAnArrayAndAnOutOfBoundIndex",
			&struct {
				Config [10]int
			}{},
			[]*envValue{},
			map[string]string{
				"CONFIG_11": "10",
			},
			testAnalyzeStructShouldFail,
		},
		{
			"WithAnArrayToValue",
			&struct {
				Config [10]int
			}{},
			[]*envValue{
				{"FOOO", path{"Config", "0"}},
				{"10", path{"Config", "1"}},
				{"true", path{"Config", "2"}},
			},
			map[string]string{
				"CONFIG_0": "FOOO",
				"CONFIG_1": "10",
				"CONFIG_2": "true",
			},
			testAnalyzeStructShouldSucceed,
		},
		{
			"WithASliceToStruct",
			&struct {
				Config []basicAppConfig
			}{},
			[]*envValue{
				{"FOOO", path{"Config", "0", "StringValue"}},
				{"10", path{"Config", "0", "IntValue"}},
				{"MIMI", path{"Config", "1", "StringValue"}},
				{"15", path{"Config", "1", "IntValue"}},
			},
			map[string]string{
				"CONFIG_0_STRING_VALUE": "FOOO",
				"CONFIG_0_INT_VALUE":    "10",
				"CONFIG_1_STRING_VALUE": "MIMI",
				"CONFIG_1_INT_VALUE":    "15",
			},
			testAnalyzeStructShouldSucceed,
		},
		{
			"WithASliceToASliceToStruct",
			&struct {
				Config [][]basicAppConfig
			}{},
			[]*envValue{
				{"FOOO", path{"Config", "0", "0", "StringValue"}},
				{"10", path{"Config", "0", "0", "IntValue"}},
				{"MIMI", path{"Config", "1", "1", "StringValue"}},
				{"15", path{"Config", "1", "1", "IntValue"}},
			},
			map[string]string{
				"CONFIG_0_0_STRING_VALUE": "FOOO",
				"CONFIG_0_0_INT_VALUE":    "10",
				"CONFIG_1_1_STRING_VALUE": "MIMI",
				"CONFIG_1_1_INT_VALUE":    "15",
			},
			testAnalyzeStructShouldSucceed,
		},
		{
			"WithASliceToAMapToStruct",
			&struct {
				Config []map[string]basicAppConfig
			}{},
			[]*envValue{
				{"FOOO", path{"Config", "0", "foo", "StringValue"}},
				{"10", path{"Config", "0", "foo", "IntValue"}},
				{"MIMI", path{"Config", "1", "bar", "StringValue"}},
				{"15", path{"Config", "1", "bar", "IntValue"}},
			},
			map[string]string{
				"CONFIG_0_FOO_STRING_VALUE": "FOOO",
				"CONFIG_0_FOO_INT_VALUE":    "10",
				"CONFIG_1_BAR_STRING_VALUE": "MIMI",
				"CONFIG_1_BAR_INT_VALUE":    "15",
			},
			testAnalyzeStructShouldSucceed,
		},
		{
			"WithAFunc",
			&struct {
				Config basicAppConfig
				Time   func() time.Time
			}{}, []*envValue{
				{"FOOO", path{"Config", "StringValue"}},
				{"10", path{"Config", "IntValue"}},
				{"true", path{"Config", "BoolValue"}},
			},
			map[string]string{
				"CONFIG_STRING_VALUE": "FOOO",
				"CONFIG_INT_VALUE":    "10",
				"CONFIG_BOOL_VALUE":   "true",
			},
			testAnalyzeStructShouldSucceed,
		},
		{
			"WithAWebBasicAuth",
			&struct {
				Basic     *Basic
				UsersFile string
			}{}, []*envValue{
				{"UserZero", path{"Basic", "0"}},
				{"UserOne", path{"Basic", "1"}},
				{"path/to/file", path{"UsersFile"}},
			},
			map[string]string{
				"BASIC_0":    "UserZero",
				"BASIC_1":    "UserOne",
				"USERS_FILE": "path/to/file",
			},
			testAnalyzeStructShouldSucceed,
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

			if result != testCase.Expectation {
				t.Logf("Expected [%s] got [%s]\n", testCase.Expectation, result)
				t.Fail()
			}
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

	config := struct {
		StringArray []string
	}{}
	testCase := struct {
		Source      interface{}
		Expectation []*envValue
		Env         map[string]string
		Then        testAnalyzeStructThenHook
	}{
		&config,
		[]*envValue{
			{"one", path{"StringArray", "0"}},
			{"two", path{"StringArray", "1"}},
		},
		map[string]string{
			"STRING_ARRAY_0": "one",
			"STRING_ARRAY_1": "two",
		},
		testAnalyzeStructShouldSucceed,
	}

	setupEnv(testCase.Env)
	parsers, _ := parse.LoadParsers(nil)
	subject := &envSource{"", "_", parsers}
	res, err := subject.analyzeStruct(
		reflect.TypeOf(testCase.Source).Elem(),
		path{},
	)
	testCase.Then(t, testCase.Expectation, res, err)
	subject.assignValues(reflect.ValueOf(&config), res, []string{})
	if config.StringArray[0] != "one" || config.StringArray[1] != "two" {
		t.Logf("Expected [[one two]] got [%v]\n", config.StringArray)
		t.FailNow()
	}
	cleanupEnv(testCase.Env)
}

func TestWithSliceToValue(t *testing.T) {
	config := struct {
		Config []int
	}{}
	testCase := testcase{
		&config,
		[]*envValue{
			{"FOOO", path{"Config", "0"}},
			{"10", path{"Config", "1"}},
			{"true", path{"Config", "2"}},
		},
		map[string]string{
			"CONFIG_0": "FOOO",
			"CONFIG_1": "10",
			"CONFIG_2": "true",
		},
		testAnalyzeStructShouldSucceed,
	}

	setupEnv(testCase.Env)
	parsers, _ := parse.LoadParsers(nil)
	subject := &envSource{"", "_", parsers}
	res, err := subject.analyzeStruct(
		reflect.TypeOf(testCase.Source).Elem(),
		path{},
	)
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
		EnvVars     []string
		Expectation []string
	}{
		{
			"WithPrefix",
			"CONFIG_APP",
			[]string{
				"CONFIG_APP_BATMAN_FOO",
				"CONFIG_APP_ROBIN_FOO",
				"CONFIG_APP_JOCKER_FOO",
			},
			[]string{
				"CONFIG_APP_BATMAN",
				"CONFIG_APP_ROBIN",
				"CONFIG_APP_JOCKER",
			},
		},
		{
			"WithDuplicates",
			"CONFIG_APP",
			[]string{
				"CONFIG_APP_BATMAN_FOO",
				"CONFIG_APP_ROBIN_FOO",
				"CONFIG_APP_JOCKER_FOO",
				"CONFIG_APP_BATMAN_BAR",
			},
			[]string{
				"CONFIG_APP_BATMAN",
				"CONFIG_APP_ROBIN",
				"CONFIG_APP_JOCKER",
				"CONFIG_APP_BATMAN",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Label, func(t *testing.T) {
			res := subject.nextLevelKeys(testCase.Prefix, testCase.EnvVars)
			for i, exp := range testCase.Expectation {
				if exp != res[i] {
					t.Logf("Unexpected value, expected [%s] got [%s]", exp, res[i])
					t.Fail()
				}
			}
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
			"WithPrefix",
			"STAERT_APP",
			map[string]string{
				"STRING_VALUE":          "FOOO",
				"INT_VALUE":             "10",
				"BOOL_VALUE":            "true",
				"STAERT_APP_BOOL_VALUE": "true",
				"STAERT_APP_BAR_VALUE":  "true",
			},
			[]string{"STAERT_APP_BAR_VALUE", "STAERT_APP_BOOL_VALUE"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Label, func(t *testing.T) {
			setupEnv(testCase.Env)
			res := subject.envVarsWithPrefix(testCase.Prefix)
			sort.Strings(res)
			for i, envVar := range testCase.Expectation {
				if res != nil && envVar != res[i] {
					t.Logf("Invalid env variableName, expected [%s] got [%s]", envVar, res[i])
					t.Fail()
				}
			}
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
			"WithDuplicates",
			[]string{"FOO", "BAR", "BIZ", "FOO", "BIZ"},
			[]string{"FOO", "BAR", "BIZ"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Label, func(t *testing.T) {
			res := unique(testCase.In)
			for i, val := range testCase.Expectation {
				if res[i] != val {
					t.Logf("Invalid result: expected [%s] got [%s]\n", val, res[i])
					t.Fail()
				}
			}
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
			if res := subject.keyFromEnvVar(testCase.EnvVar, testCase.Prefix); res != testCase.Expectation {
				t.Logf("Unexpected value, expected [%s] got [%s]", testCase.Expectation, res)
				t.Fail()
			}
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
