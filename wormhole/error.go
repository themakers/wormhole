package wormhole

//import "encoding/json"
//
//type ErrorCode string
//
//const (
//	//ErrorUnclassified         ErrorCode = "unclassified"
//	ErrorInternal         ErrorCode = "internal"
//	ErrorNotFound         ErrorCode = "not_found"
//	ErrorPermissionDenied ErrorCode = "permission_denied"
//)
//
//var _ error = new(Error)
//
//type Error struct {
//	Code    ErrorCode         `json:"Code"`
//	Type    string            `json:"Type"`
//	Message string            `json:"Message"`
//	Fields  map[string]string `json:"Fields"`
//  // TODO IsRemote bool ????????????????????
//}
//
//func NewError(code ErrorCode, tpe, message string, fields ...string) error {
//	e := &Error{
//		Code:    code,
//		Type:    tpe,
//		Message: message,
//	}
//
//	for i := 0; i < len(fields); i += 2 {
//		k := fields[i]
//		v := ""
//		if i+1 < len(fields) {
//			v = fields[i+1]
//		}
//		e.Fields[k] = v
//	}
//
//	return e
//}
//
//func FromError(err error) *Error {
//	if err == nil {
//		return nil
//	} else if e, ok := err.(*Error); ok {
//		return e
//	} else {
//		return &Error{
//			Code:    ErrorInternal,
//			Message: err.Error(),
//		}
//	}
//}
//
//func IsError(err error) (*Error, bool) {
//	if e, ok := err.(*Error); ok {
//		return e, true
//	} else {
//		return nil, false
//	}
//}
//
//func (e *Error) Error() string {
//	data, err := json.Marshal(e)
//	if err != nil {
//		panic(err)
//	}
//	return string(data)
//}
