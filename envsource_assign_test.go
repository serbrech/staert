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
		Source      interface{}
		Values      []*envValue
		Expectation interface{}
	}{
		{
			Label: "BasicStruct",
			Source: &struct {
				StringValue      string
				OtherStringValue string
			}{},
			Values: []*envValue{
				{"FOO", path{"StringValue"}},
				{"BAR", path{"OtherStringValue"}},
			},
			Expectation: &struct {
				StringValue      string
				OtherStringValue string
			}{"FOO", "BAR"},
		},
		{
			Label: "BasicStructWithParser",
			Source: &struct {
				StringValue string
				IntValue    int
			}{},
			Values: []*envValue{
				{"FOO", path{"StringValue"}},
				{"1", path{"IntValue"}},
			},
			Expectation: &struct {
				StringValue string
				IntValue    int
			}{"FOO", 1},
		},
		{
			Label: "BasicStructEmbedded",
			Source: &struct {
				StringValue string
				Next        basicAppConfig
			}{},
			Values: []*envValue{
				{"FOO", path{"StringValue"}},
				{"1", path{"Next", "IntValue"}},
				{"true", path{"Next", "BoolValue"}},
				{"string", path{"Next", "StringValue"}},
			},
			Expectation: &struct {
				StringValue string
				Next        basicAppConfig
			}{"FOO", basicAppConfig{
				BoolValue:   true,
				IntValue:    1,
				StringValue: "string",
			}},
		},
		{
			Label: "BasicStructPointer",
			Source: &struct {
				StringValue string
				NextPointer *basicAppConfig
			}{},
			Values: []*envValue{
				{"FOO", path{"StringValue"}},
				{"1", path{"NextPointer", "IntValue"}},
				{"true", path{"NextPointer", "BoolValue"}},
				{"string", path{"NextPointer", "StringValue"}},
			},
			Expectation: &struct {
				StringValue string
				NextPointer *basicAppConfig
			}{"FOO", &basicAppConfig{
				BoolValue:   true,
				IntValue:    1,
				StringValue: "string",
			}},
		},
		{
			Label: "BasicStructPointerPointer",
			Source: &struct {
				StringValue string
				NextPointer **basicAppConfig
			}{},
			Values: []*envValue{
				{"FOO", path{"StringValue"}},
				{"1", path{"NextPointer", "IntValue"}},
				{"true", path{"NextPointer", "BoolValue"}},
				{"string", path{"NextPointer", "StringValue"}},
			},
			Expectation: expectedPtrPtr,
		},
		{
			Label:  "WithWrongPath",
			Source: &delegatorType{},
			Values: []*envValue{
				{"FOO", path{"WrongPath"}},
			},
			Expectation: &delegatorType{},
		},
		{
			Label:  "WithInterfaceDelegation",
			Source: &delegatorType{},
			Values: []*envValue{
				{"FOO", path{"StringValue"}},
				{"1", path{"IntValue"}},
			},
			Expectation: &delegatorType{
				IntValue:    1,
				StringValue: "FOO",
			},
		},
		{
			Label: "WithMapOfStringValues",
			Source: &struct {
				Config map[string]string
			}{},
			Values: []*envValue{
				{"FOO", path{"Config", "foo"}},
				{"MEH", path{"Config", "bar"}},
				{"BAR", path{"Config", "biz"}},
			},
			Expectation: &struct {
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
			Label: "CanParseMapValues",
			Source: &struct {
				Config map[string]int
			}{},
			Values: []*envValue{
				{"1", path{"Config", "foo"}},
				{"2", path{"Config", "bar"}},
				{"3", path{"Config", "biz"}},
			},
			Expectation: &struct {
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
			Label: "CanParseMapValuesAndKeys",
			Source: &struct {
				Config map[int]int
			}{},
			Values: []*envValue{
				{"1", path{"Config", "1"}},
				{"2", path{"Config", "2"}},
				{"3", path{"Config", "3"}},
			},
			Expectation: &struct {
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
			Label: "WithMapofStringToStruct",
			Source: &struct {
				Config map[string]basicAppConfig
			}{},
			Values: []*envValue{
				{"FOOO", path{"Config", "foo", "StringValue"}},
				{"10", path{"Config", "foo", "IntValue"}},
			},
			Expectation: &struct {
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
			Label: "WithMapofIntToStruct",
			Source: &struct {
				Config map[int]basicAppConfig
			}{},
			Values: []*envValue{
				{"FOOO", path{"Config", "0", "StringValue"}},
				{"10", path{"Config", "0", "IntValue"}},
			},
			Expectation: &struct {
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
			Label: "WithArrayofInts",
			Source: &struct {
				Config []int
			}{},
			Values: []*envValue{
				{"1", path{"Config", "0"}},
				{"10", path{"Config", "1"}},
			},
			Expectation: &struct {
				Config []int
			}{
				Config: []int{1, 10},
			},
		},
		{
			Label: "WithArrayofStructs",
			Source: &struct {
				Config []basicAppConfig
			}{},
			Values: []*envValue{
				{"Test", path{"Config", "0", "StringValue"}},
				{"10", path{"Config", "0", "IntValue"}},
				{"true", path{"Config", "0", "BoolValue"}},
				{"Test2", path{"Config", "1", "StringValue"}},
				{"20", path{"Config", "1", "IntValue"}},
				{"false", path{"Config", "1", "BoolValue"}},
			},
			Expectation: &struct {
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
		{
			Label: "WithArrayofPointerToStructs",
			Source: &struct {
				Config []*basicAppConfig
			}{},
			Values: []*envValue{
				{"Test", path{"Config", "0", "StringValue"}},
				{"10", path{"Config", "0", "IntValue"}},
				{"true", path{"Config", "0", "BoolValue"}},
				{"Test2", path{"Config", "1", "StringValue"}},
				{"20", path{"Config", "1", "IntValue"}},
				{"false", path{"Config", "1", "BoolValue"}},
			},
			Expectation: &struct {
				Config []*basicAppConfig
			}{
				Config: []*basicAppConfig{
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
			err := subject.assignValues(reflect.ValueOf(testCase.Source).Elem(), testCase.Values, []string{})
			if err != nil {
				t.Logf("Expected no error, got %s", err.Error())
				t.Fail()
			}

			assert.Exactly(t, testCase.Expectation, testCase.Source)
		})
	}
}

func TestFilterEnvVarWithPrefix(t *testing.T) {
	envSource := []*envValue{
		{"FOOO", path{"Config", "0", "foo", "StringValue"}},
		{"10", path{"Config", "0", "foo", "IntValue"}},
		{"10", path{"Config", "IntValue"}},
		{"10", path{"Config", "0", "0", "IntValue"}},
	}

	result := filterEnvVarWithPrefix(envSource, []string{"Config", "0", "foo"})

	expected := []*envValue{
		{"FOOO", path{"StringValue"}},
		{"10", path{"IntValue"}},
	}
	assert.Exactly(t, expected, result)
}
