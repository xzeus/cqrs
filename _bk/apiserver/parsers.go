package apiserver

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
)

var ErrEmptyValue = errors.New("cannot parse empty value")

func Hex32(v int32) string {
	return fmt.Sprintf("%X", uint32(v))
}

func Hex64(v int64) string {
	return fmt.Sprintf("%X", uint64(v))
}

// Bit sizes 0, 8, 16, 32, and 64 correspond to int, int8, int16, int32, and int64
func ParseIntFromMap(valueMap map[string]string, base int, bitSize int, key string) (int64, error) {
	value, ok := valueMap[key]
	if !ok {
		return 0, errors.New(fmt.Sprintf("Unable to find key [ %s ]", key))
	}
	result, err := strconv.ParseInt(value, base, bitSize)
	if err != nil { // Try to parse with base 64 encoding
		base64_value, err := base64.StdEncoding.DecodeString(value)
		if err != nil { // Data wasn't numeric or encoded
			return 0, errors.New(fmt.Sprintf("Unable to convert value [ %d ] to base64", value))
		}
		result, err = strconv.ParseInt(string(base64_value), base, bitSize)
		if err != nil { // Decoded value wasn't parsable
			return 0, errors.New(fmt.Sprintf("Unable to convert decoded value [ %s ] to int64", base64_value))
		}
	}
	return result, nil
}

func ParseUintFromMap(valueMap map[string]string, base int, bitSize int, key string) (uint64, error) {
	value, ok := valueMap[key]
	if !ok {
		return 0, errors.New(fmt.Sprintf("Unable to find key [ %s ] ", key))
	}
	result, err := strconv.ParseUint(value, base, bitSize)
	if err != nil { // Try to parse with base 64 encoding
		base64_value, err := base64.StdEncoding.DecodeString(value)
		if err != nil { // Data wasn't numeric or encoded
			return 0, errors.New(fmt.Sprintf("Unable to convert value [ %d ] to base64", value))
		}
		result, err = strconv.ParseUint(string(base64_value), base, bitSize)
		if err != nil { // Decoded value wasn't parsable
			return 0, errors.New(fmt.Sprintf("Unable to convert decoded value [ %s ] to uint64", base64_value))
		}
	}
	return result, nil
}

func ParseInt(value string, base int, bitSize int) (int64, error) {
	if value == "" {
		return 0, ErrEmptyValue
	}
	result, err := strconv.ParseInt(value, base, bitSize)
	if err != nil { // Try to parse with base 64 encoding
		base64_value, err := base64.StdEncoding.DecodeString(value)
		if err != nil { // Data wasn't numeric or encoded
			return 0, errors.New(fmt.Sprintf("Unable to convert value [ %d ] to base64", value))
		}
		result, err = strconv.ParseInt(string(base64_value), base, bitSize)
		if err != nil { // Decoded value wasn't parsable
			return 0, errors.New(fmt.Sprintf("Unable to convert decoded value [ %s ] to int64", base64_value))
		}
	}
	return result, nil
}

func ParseUint(v string, base int, bitSize int) (r uint64, err error) {
	if v == "" {
		return 0, ErrEmptyValue
	}
	if r, err = strconv.ParseUint(v, base, bitSize); err == nil {
		return
	} // Not a straight conversion
	var b []byte
	if b, err = base64.StdEncoding.DecodeString(v); err != nil {
		err = errors.New(fmt.Sprintf("Unable to convert value [ %d ] to base64", v))
	} else {
		if r, err = strconv.ParseUint(string(b), base, bitSize); err != nil {
			err = errors.New(fmt.Sprintf("Unable to convert decoded value [ %s ] to uint64", b))
		}
	}
	return
}

func Int32(value string) (int32, error) {
	result, err := ParseUint(value, 16, 32)
	if err != nil {
		return 0, err
	}
	return int32(result), nil
}

func Int64(value string) (int64, error) {
	result, err := ParseUint(value, 16, 64)
	if err != nil {
		return 0, err
	}
	return int64(result), nil
}

func Uint32(value string) (uint32, error) {
	result, err := ParseUint(value, 16, 32)
	if err != nil {
		return 0, err
	}
	return uint32(result), nil
}

func Uint64(value string) (uint64, error) {
	result, err := ParseUint(value, 16, 64)
	if err != nil {
		return 0, err
	}
	return uint64(result), nil
}
