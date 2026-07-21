package api

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
)

// RawMessage maps custom json/jsonb GraphQL scalars.
type RawMessage = json.RawMessage

// MarshalRawMessage marshals a RawMessage scalar for gqlgen.
func MarshalRawMessage(r RawMessage) graphqlMarshaler {
	return graphqlMarshaler{r: r}
}

// UnmarshalRawMessage unmarshals a RawMessage scalar for gqlgen.
func UnmarshalRawMessage(v interface{}) (RawMessage, error) {
	switch val := v.(type) {
	case string:
		return RawMessage(val), nil
	case []byte:
		return RawMessage(val), nil
	default:
		b, err := json.Marshal(v)
		return RawMessage(b), err
	}
}

type graphqlMarshaler struct {
	r RawMessage
}

func (m graphqlMarshaler) MarshalGQL(w io.Writer) {
	if len(m.r) == 0 {
		w.Write([]byte("null"))
		return
	}
	w.Write(m.r)
}

// Numeric is a custom GraphQL scalar that maps to Hardcover's "numeric" type.
type Numeric float64

func (n Numeric) GetGraphQLType() string { return "numeric" }

func MarshalNumeric(n float64) graphqlMarshaler {
	return graphqlMarshaler{r: RawMessage(fmt.Sprintf("%f", n))}
}

func UnmarshalNumeric(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case string:
		return strconv.ParseFloat(val, 64)
	default:
		return 0, fmt.Errorf("invalid numeric %v", v)
	}
}

// Date is a custom GraphQL scalar.
type Date string

func (d Date) GetGraphQLType() string { return "date" }
