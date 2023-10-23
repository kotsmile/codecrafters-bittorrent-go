package torrent

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type List []interface{}
type Dict map[string]interface{}

const (
	IntegerPrefix   = 'i'
	ListPrefix      = 'l'
	DictPrefix      = 'd'
	EndPostfix      = 'e'
	StringSeparator = ':'
)

func DecodeNextInteger(value string) (int, int, error) {
	firstEnd := strings.Index(value, string(EndPostfix))
	integer, err := strconv.Atoi(value[1:firstEnd])
	if err != nil {
		return 0, 0, err
	}

	return integer, firstEnd + 1, nil
}

func DecodeNextStr(value string) (string, int, error) {
	indexColon := strings.Index(value, string(StringSeparator))
	length, err := strconv.Atoi(value[:indexColon])
	if err != nil {
		return "", 0, err
	}

	content := value[indexColon+1 : indexColon+1+length]
	return content, indexColon + 1 + length, err
}

func DecodeNextList(value string) (List, int, error) {
	list := make(List, 0, 10)
	index := 1
	for value[index] != EndPostfix {
		element, endIndex, err := DecodeBencode(value[index:])
		if err != nil {
			return nil, 0, err
		}

		index += endIndex
		list = append(list, element)
	}

	return list, index + 1, nil
}

func DecodeNextDict(value string) (Dict, int, error) {
	dict := make(Dict)
	index := 1
	for value[index] != EndPostfix {
		element, endIndex, err := DecodeBencode(value[index:])
		if err != nil {
			return nil, 0, err
		}

		key, ok := element.(string)
		if !ok {
			return nil, 0, errors.New("key is not string")
		}

		value, endIndexValue, err := DecodeBencode(value[index+endIndex:])
		if err != nil {
			return nil, 0, err
		}

		dict[key] = value
		index += endIndex + endIndexValue
	}

	return dict, index + 1, nil
}

func DecodeBencode(bencodedStr string) (interface{}, int, error) {
	switch bencodedStr[0] {
	case DictPrefix:
		return DecodeNextDict(bencodedStr)
	case ListPrefix:
		return DecodeNextList(bencodedStr)
	case IntegerPrefix:
		return DecodeNextInteger(bencodedStr)
	default:
		return DecodeNextStr(bencodedStr)
	}
}

func EncodeBencode(benocodedData interface{}) (string, error) {
	switch data := benocodedData.(type) {
	case string:
		return fmt.Sprintf("%d:%s", len(data), data), nil
	case int:
		return fmt.Sprintf("%c%d%c", IntegerPrefix, data, EndPostfix), nil
	case List:
		elementsString := ""
		for _, elem := range data {
			encodeElement, err := EncodeBencode(elem)
			if err != nil {
				return "", err
			}

			elementsString += encodeElement
		}

		return fmt.Sprintf("%c%s%c", ListPrefix, elementsString, EndPostfix), nil
	case Dict:
		elementsString := ""

		keys := make([]string, 0)
		for k := range data {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			value, ok := data[key]
			if !ok {
				return "", errors.New("unreacheable")
			}

			enocodedKey, err := EncodeBencode(key)
			if err != nil {
				return "", err
			}

			encodedValue, err := EncodeBencode(value)
			if err != nil {
				return "", err
			}

			elementsString += fmt.Sprintf("%s%s", enocodedKey, encodedValue)

		}

		return fmt.Sprintf("%c%s%c", DictPrefix, elementsString, EndPostfix), nil
	default:
		return "", errors.New("wrong type")
	}
}
