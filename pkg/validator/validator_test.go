package validator

import "testing"

func TestValidateStruct(t *testing.T) {
	type tData struct {
		Email string `validate:"email"`
	}
	type tDataSlice []struct {
		Name []string `validate:"required"`
	}

	tests := []struct {
		name       string
		mock       func() interface{}
		wantResult bool
		wantErr    bool
	}{
		{
			name: "test success",
			mock: func() interface{} {
				return tData{
					"email@email.com",
				}
			},
			wantResult: true,
			wantErr:    false,
		}, {
			name: "test success slice",
			mock: func() interface{} {
				return tDataSlice{struct {
					Name []string "validate:\"required\""
				}{[]string{"A", "B", "C"}}}
			},
			wantResult: true,
			wantErr:    false,
		}, {
			name: "test fail",
			mock: func() interface{} {
				return tData{
					"email@.com",
				}
			},
			wantResult: false,
			wantErr:    true,
		}, {
			name: "test success slice empty",
			mock: func() interface{} {
				return tDataSlice{}
			},
			wantResult: false,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := ValidateStruct(tt.mock())
			if (err != nil) != tt.wantErr {
				t.Errorf("pv.ValidateStruct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotResult != tt.wantResult {
				t.Errorf("pv.ValidateStruct() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}
