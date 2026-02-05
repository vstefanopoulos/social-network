package ct

import "fmt"

func UserKey(receiverId any) string {
	return fmt.Sprintf("user.%v", receiverId)
}

func GroupKey(groupId any) string {
	return fmt.Sprintf("grm.%v", groupId)
}
