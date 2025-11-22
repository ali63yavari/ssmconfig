package ssmconfig

func ToPointerValue[TValue interface{}](value TValue) *TValue {
	v := value
	return &v
}
