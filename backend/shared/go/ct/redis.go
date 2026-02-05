package ct

import (
	"fmt"
)

type BasicUserInfoKey struct {
	Id Id
}

func (k BasicUserInfoKey) GenKey() (string, error) {
	if err := k.Id.Validate(); err != nil {
		return "", err
	}
	return fmt.Sprintf("basic_user_info:%d", k.Id), nil
}

func (k BasicUserInfoKey) String() string {
	return fmt.Sprintf("basic_user_info:%d", k.Id)
}

type ImageKey struct {
	Variant FileVariant
	Id      Id
}

func (k ImageKey) GenKey() (string, error) {
	if err := ValidateStruct(k); err != nil {
		return "", err
	}
	return fmt.Sprintf("img_%s:%d", k.Variant, k.Id), nil
}

func (k ImageKey) String() string {
	return fmt.Sprintf("img_%s:%d", k.Variant, k.Id)
}

type IsGroupMemberKey struct {
	GroupId Id
	UserId  Id
}

func (k IsGroupMemberKey) GenKey() (string, error) {
	if err := ValidateStruct(k); err != nil {
		return "", err
	}
	return fmt.Sprintf("is_group:%d.member:%d", k.GroupId.Int64(), k.UserId.Int64()), nil
}

func (k IsGroupMemberKey) String() string {
	return fmt.Sprintf("is_group:%d.member:%d", k.GroupId.Int64(), k.UserId.Int64())
}
