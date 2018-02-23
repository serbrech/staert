package staert

import (
	"reflect"
	"testing"

	"github.com/containous/flaeg/parse"

	"github.com/stretchr/testify/assert"
)

func getPtrPtrConfig() *struct {
	StringValue string
	NextPointer **basicAppConfig
} {
	configPtr := &basicAppConfig{
		BoolValue:   true,
		IntValue:    1,
		StringValue: "string",
	}
	expectedPtrPtr := &struct {
		StringValue string
		NextPointer **basicAppConfig
	}{
		StringValue: "FOO",
	}
	expectedPtrPtr.NextPointer = &configPtr
	return expectedPtrPtr
}

func TestAssignValues(t *testing.T) {

	parsers, _ := parse.LoadParsers(nil)
	subject := &envSource{
		"",
		"_",
		parsers,
	}

	expectedPtrPtr := getPtrPtrConfig()

	testCases := []struct {
		Label       string
		Value       interface{}
		Values      []*envValue
		Expectation interface{}
	}{
		{
			"BasicStruct",
			&struct {
				StringValue      string
				OtherStringValue string
			}{},
			[]*envValue{
				&envValue{"FOO", path{"StringValue"}},
				&envValue{"BAR", path{"OtherStringValue"}},
			},
			&struct {
				StringValue      string
				OtherStringValue string
			}{"FOO", "BAR"},
		},
		{
			"BasicStructWithParser",
			&struct {
				StringValue string
				IntValue    int
			}{},
			[]*envValue{
				&envValue{"FOO", path{"StringValue"}},
				&envValue{"1", path{"IntValue"}},
			},
			&struct {
				StringValue string
				IntValue    int
			}{"FOO", 1},
		},
		{
			"BasicStructEmbedded",
			&struct {
				StringValue string
				Next        basicAppConfig
			}{},
			[]*envValue{
				&envValue{"FOO", path{"StringValue"}},
				&envValue{"1", path{"Next", "IntValue"}},
				&envValue{"true", path{"Next", "BoolValue"}},
				&envValue{"string", path{"Next", "StringValue"}},
			},
			&struct {
				StringValue string
				Next        basicAppConfig
			}{"FOO", basicAppConfig{
				BoolValue:   true,
				IntValue:    1,
				StringValue: "string",
			}},
		},
		{
			"BasicStructPointer",
			&struct {
				StringValue string
				NextPointer *basicAppConfig
			}{},
			[]*envValue{
				&envValue{"FOO", path{"StringValue"}},
				&envValue{"1", path{"NextPointer", "IntValue"}},
				&envValue{"true", path{"NextPointer", "BoolValue"}},
				&envValue{"string", path{"NextPointer", "StringValue"}},
			},
			&struct {
				StringValue string
				NextPointer *basicAppConfig
			}{"FOO", &basicAppConfig{
				BoolValue:   true,
				IntValue:    1,
				StringValue: "string",
			}},
		},
		{
			"BasicStructPointerPointer",
			&struct {
				StringValue string
				NextPointer **basicAppConfig
			}{},
			[]*envValue{
				&envValue{"FOO", path{"StringValue"}},
				&envValue{"1", path{"NextPointer", "IntValue"}},
				&envValue{"true", path{"NextPointer", "BoolValue"}},
				&envValue{"string", path{"NextPointer", "StringValue"}},
			},
			expectedPtrPtr,
		},
		{
			"WithWrongPath",
			&delegatorType{},
			[]*envValue{
				&envValue{"FOO", path{"WrongPath"}},
			},
			&delegatorType{},
		},
		{
			"WithInterfaceDelegation",
			&delegatorType{},
			[]*envValue{
				&envValue{"FOO", path{"StringValue"}},
				&envValue{"1", path{"IntValue"}},
			},
			&delegatorType{
				IntValue:    1,
				StringValue: "FOO",
			},
		},
		{
			"WithMapOfStringValues",
			&struct {
				Config map[string]string
			}{},
			[]*envValue{
				&envValue{"FOO", path{"Config", "foo"}},
				&envValue{"MEH", path{"Config", "bar"}},
				&envValue{"BAR", path{"Config", "biz"}},
			},
			&struct {
				Config map[string]string
			}{
				Config: map[string]string{
					"foo": "FOO",
					"bar": "MEH",
					"biz": "BAR",
				},
			},
		},
		{
			"CanParseMapValues",
			&struct {
				Config map[string]int
			}{},
			[]*envValue{
				&envValue{"1", path{"Config", "foo"}},
				&envValue{"2", path{"Config", "bar"}},
				&envValue{"3", path{"Config", "biz"}},
			},
			&struct {
				Config map[string]int
			}{
				Config: map[string]int{
					"foo": 1,
					"bar": 2,
					"biz": 3,
				},
			},
		},
		{
			"CanParseMapValuesAndKeys",
			&struct {
				Config map[int]int
			}{},
			[]*envValue{
				&envValue{"1", path{"Config", "1"}},
				&envValue{"2", path{"Config", "2"}},
				&envValue{"3", path{"Config", "3"}},
			},
			&struct {
				Config map[int]int
			}{
				Config: map[int]int{
					1: 1,
					2: 2,
					3: 3,
				},
			},
		},
		{
			"WithMapofStringToStruct",
			&struct {
				Config map[string]basicAppConfig
			}{},
			[]*envValue{
				&envValue{"FOOO", path{"Config", "foo", "StringValue"}},
				&envValue{"10", path{"Config", "foo", "IntValue"}},
			},
			&struct {
				Config map[string]basicAppConfig
			}{
				Config: map[string]basicAppConfig{
					"foo": basicAppConfig{
						StringValue: "FOOO",
						IntValue:    10,
					},
				},
			},
		},
		{
			"WithMapofIntToStruct",
			&struct {
				Config map[int]basicAppConfig
			}{},
			[]*envValue{
				&envValue{"FOOO", path{"Config", "0", "StringValue"}},
				&envValue{"10", path{"Config", "0", "IntValue"}},
			},
			&struct {
				Config map[int]basicAppConfig
			}{
				Config: map[int]basicAppConfig{
					0: basicAppConfig{
						StringValue: "FOOO",
						IntValue:    10,
					},
				},
			},
		},
		{
			"WithArrayofInts",
			&struct {
				Config []int
			}{},
			[]*envValue{
				&envValue{"1", path{"Config", "0"}},
				&envValue{"10", path{"Config", "1"}},
			},
			&struct {
				Config []int
			}{
				Config: []int{1, 10},
			},
		},
		{
			"WithArrayofStructs",
			&struct {
				Config []basicAppConfig
			}{},
			[]*envValue{
				&envValue{"Test", path{"Config", "0", "StringValue"}},
				&envValue{"10", path{"Config", "0", "IntValue"}},
				&envValue{"true", path{"Config", "0", "BoolValue"}},
				&envValue{"Test2", path{"Config", "1", "StringValue"}},
				&envValue{"20", path{"Config", "1", "IntValue"}},
				&envValue{"false", path{"Config", "1", "BoolValue"}},
			},
			&struct {
				Config []basicAppConfig
			}{
				Config: []basicAppConfig{
					{
						BoolValue:   true,
						IntValue:    10,
						StringValue: "Test",
					},
					{
						BoolValue:   false,
						IntValue:    20,
						StringValue: "Test2",
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Label, func(t *testing.T) {
			err := subject.assignValues(reflect.ValueOf(testCase.Value).Elem(), testCase.Values, []string{})
			if err != nil {
				t.Logf("Expected no error, got %s", err.Error())
				t.Fail()
			}

			assert.Exactly(t, testCase.Expectation, testCase.Value)
		})
	}
}

func TestFilterEnvVarWithPrefix(t *testing.T) {
	envSource := []*envValue{
		&envValue{"FOOO", path{"Config", "0", "foo", "StringValue"}},
		&envValue{"10", path{"Config", "0", "foo", "IntValue"}},
		&envValue{"10", path{"Config", "IntValue"}},
		&envValue{"10", path{"Config", "0", "0", "IntValue"}},
	}

	result := filterEnvVarWithPrefix(envSource, []string{"Config", "0", "foo"})

	expected := []*envValue{
		&envValue{"FOOO", path{"StringValue"}},
		&envValue{"10", path{"IntValue"}},
	}
	assert.Exactly(t, expected, result)
}
