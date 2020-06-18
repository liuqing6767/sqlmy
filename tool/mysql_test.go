package dao

import (
	"os"
	"testing"
)

func Test_underlineConvert(t *testing.T) {
	tests := []struct {
		name      string
		val       string
		wantShort string
		wantMid   string
		wantLog   string
	}{
		{
			name:      "case1",
			val:       "",
			wantShort: "",
			wantMid:   "",
			wantLog:   "",
		},
		{
			name:      "case2",
			val:       "_",
			wantShort: "",
			wantMid:   "",
			wantLog:   "",
		},
		{
			name:      "case3",
			val:       "_a_b_c",
			wantShort: "abc",
			wantMid:   "aBC",
			wantLog:   "ABC",
		},
		{
			name:      "case4",
			val:       "a_b_c",
			wantShort: "abc",
			wantMid:   "aBC",
			wantLog:   "ABC",
		},
		{
			name:      "case5",
			val:       "AAA_BB_CC",
			wantShort: "abc",
			wantMid:   "aAABBCC",
			wantLog:   "AAABBCC",
		},
		{
			name:      "case6",
			val:       "Aaa_Bb_Cc",
			wantShort: "abc",
			wantMid:   "aaaBbCc",
			wantLog:   "AaaBbCc",
		},
		{
			name:      "case7",
			val:       "aa_bb_cc",
			wantShort: "abc",
			wantMid:   "aaBbCc",
			wantLog:   "AaBbCc",
		},
		{
			name:      "case12",
			val:       "__",
			wantShort: "",
			wantMid:   "",
			wantLog:   "",
		},
		{
			name:      "case13",
			val:       "_a_b_c_",
			wantShort: "abc",
			wantMid:   "aBC",
			wantLog:   "ABC",
		},
		{
			name:      "case14",
			val:       "a_b_c_",
			wantShort: "abc",
			wantMid:   "aBC",
			wantLog:   "ABC",
		},
		{
			name:      "case15",
			val:       "AAA_BB_CC_",
			wantShort: "abc",
			wantMid:   "aAABBCC",
			wantLog:   "AAABBCC",
		},
		{
			name:      "case16",
			val:       "Aaa_Bb_Cc_",
			wantShort: "abc",
			wantMid:   "aaaBbCc",
			wantLog:   "AaaBbCc",
		},
		{
			name:      "case17",
			val:       "aa_bb_cc_",
			wantShort: "abc",
			wantMid:   "aaBbCc",
			wantLog:   "AaBbCc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotShort, gotMid, gotLog := underlineConvert(tt.val)
			if gotShort != tt.wantShort {
				t.Errorf("underlineConvert() gotShort = %v, want %v", gotShort, tt.wantShort)
			}
			if gotMid != tt.wantMid {
				t.Errorf("underlineConvert() gotMid = %v, want %v", gotMid, tt.wantMid)
			}
			if gotLog != tt.wantLog {
				t.Errorf("underlineConvert() gotLog = %v, want %v", gotLog, tt.wantLog)
			}
		})
	}
}

func Test_first2Lower(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want string
	}{
		{"case1", "", ""},
		{"case2", "a", "a"},
		{"case1", "aa", "aa"},
		{"case1", "Aa", "aa"},
		{"case1", "zz", "zz"},
		{"case1", "ZZ", "zZ"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := first2Lower(tt.str); got != tt.want {
				t.Errorf("first2Lower() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_first2Upper(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want string
	}{
		{"case1", "", ""},
		{"case2", "a", "A"},
		{"case1", "aa", "Aa"},
		{"case1", "Aa", "Aa"},
		{"case1", "zz", "Zz"},
		{"case1", "ZZ", "ZZ"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := first2Upper(tt.str); got != tt.want {
				t.Errorf("first2Lower() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMysqlDao_Create(t *testing.T) {
	tests := []struct {
		name string
		md   *MysqlDao
		want int
	}{
		{
			name: "case1",
			md: &MysqlDao{
				Table:  "",
				DSN:    "root:12345678@127.0.0.1:3306/demo",
				Output: os.Stderr,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.md.Create(); got != tt.want {
				t.Errorf("Create() = %v, want %v", got, tt.want)
			}
		})
	}
}
