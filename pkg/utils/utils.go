package utils

import "github.com/PharmaKart/order-svc/internal/proto"

func ConvertMapToKeyValuePairs(m map[string]string) []*proto.KeyValuePair {
	if m == nil {
		return nil
	}

	result := make([]*proto.KeyValuePair, 0, len(m))
	for k, v := range m {
		result = append(result, &proto.KeyValuePair{
			Key:   k,
			Value: v,
		})
	}
	return result
}
