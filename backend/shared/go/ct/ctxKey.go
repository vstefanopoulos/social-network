package ct

// type alias:
// context key type ALIAS in order to help enforcing a single source of truth for key namings
type CtxKey = string

// func (c CtxKey) String() string {
// 	return string(c)
// }

// Holds the keys to values on request context.
// warning! keys that are meant to be propagated through grpc services have strict requirements! They must be ascii, lowercase, and only allowed symbols: "-_."
const (
	ClaimsKey        = "jwt-claims"
	ReqActionDetails = "action-details"
	ReqTimestamp     = "timestamp"
	ReqID            = "request-id"
	UserId           = "user-id"
	IP               = "ip"
)

type ctxKeys struct {
	keys []CtxKey
}

func (ctxk *ctxKeys) GetKeys() []CtxKey {
	return ctxk.keys
}

// Common keys that will by default be propagated by all services
var commonKeys = ctxKeys{
	keys: []CtxKey{UserId, ReqID, IP},
}

// returns the common keys, you can add more if you want
// meant to be used to receive the default keys that will be propagated across all services
func CommonKeys(extraKeys ...CtxKey) *ctxKeys {
	commonKeys.keys = append(commonKeys.keys, extraKeys...)
	return &commonKeys
}
