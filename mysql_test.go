package sqlmy

import (
	"reflect"
	"testing"
)

func Test_tagSplitter(t *testing.T) {
	tests := []struct {
		name    string
		dbTag   string
		wantKey string
		wantOpt string
	}{
		{
			name:    "case1",
			dbTag:   "",
			wantKey: "",
			wantOpt: "=",
		},
		{
			name:    "case2",
			dbTag:   "a",
			wantKey: "a",
			wantOpt: "=",
		},
		{
			name:    "case3",
			dbTag:   "a,",
			wantKey: "a",
			wantOpt: "",
		},
		{
			name:    "case4",
			dbTag:   "a, not in",
			wantKey: "a",
			wantOpt: "not in",
		},
		{
			name:    "case5",
			dbTag:   "a, not in ",
			wantKey: "a",
			wantOpt: "not in",
		},
		{
			name:    "case6",
			dbTag:   "a, = ",
			wantKey: "a",
			wantOpt: "=",
		},
		{
			name:    "case7",
			dbTag:   " a , not in",
			wantKey: "a",
			wantOpt: "not in",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKey, gotOpt := tagSplitter(tt.dbTag)
			if gotKey != tt.wantKey {
				t.Errorf("tagSplitter() gotKey = %v, want %v", gotKey, tt.wantKey)
			}
			if gotOpt != tt.wantOpt {
				t.Errorf("tagSplitter() gotOpt = %v, want %v", gotOpt, tt.wantOpt)
			}
		})
	}
}

type Person struct {
	ID      int64  `ddb:"id,="`
	Name    string `ddb:"name,!="`
	IsMan   bool   `ddb:"is_man"`
	Nation  string
	City    string  `ddb:"-"`
	Age     *int    `ddb:"age,"`
	Company *string `ddb:"company"`
	Nums    []int   `ddb:"nums,in"`
}

var age = 6
var company = "company"

func TestStruct2Wheres(t *testing.T) {
	tests := []struct {
		name        string
		structValue interface{}
		want        map[string]interface{}
	}{
		{
			name:        "case1",
			structValue: &Person{},
			want: map[string]interface{}{
				"id":      int64(0),
				"name !=": "",
				"is_man":  false,
				"Nation":  "",
			},
		},
		{
			name: "case2",
			structValue: &Person{
				ID: 5,
			},
			want: map[string]interface{}{
				"id":      int64(5),
				"name !=": "",
				"is_man":  false,
				"Nation":  "",
			},
		},
		{
			name: "case3",
			structValue: &Person{
				ID:    5,
				IsMan: true,
				Nums:  []int{},
			},
			want: map[string]interface{}{
				"id":      int64(5),
				"name !=": "",
				"is_man":  true,
				"Nation":  "",
				"nums in": []int{},
			},
		},
		{
			name: "case4",
			structValue: &Person{
				ID:      5,
				IsMan:   true,
				Nums:    []int{1},
				Age:     &age,
				Company: &company,
			},
			want: map[string]interface{}{
				"id":      int64(5),
				"name !=": "",
				"is_man":  true,
				"Nation":  "",
				"nums in": []int{1},
				"age":     6,
				"company": company,
			},
		},
		{
			name:        "case6",
			structValue: nil,
			want:        map[string]interface{}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Struct2Where(tt.structValue); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Struct2Wheres() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}
