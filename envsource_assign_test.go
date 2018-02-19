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
	}

	for _, testCase := range testCases {
		t.Run(testCase.Label, func(t *testing.T) {
			err := subject.assignValues(reflect.ValueOf(testCase.Value).Elem(), testCase.Values)
			if err != nil {
				t.Logf("Expected no error, got %s", err.Error())
				t.Fail()
			}

			assert.Exactly(t, testCase.Expectation, testCase.Value)
		})
	}
}
