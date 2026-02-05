package application

//---------------------------------------------------------------------
// low priority
//---------------------------------------------------------------------

func DeleteUser() {}

//called with: user_id
//returns success or error
//request needs to come from same user or admin (not implemented)
//---------------------------------------------------------------------
//softDeleteUser(id)

func BanUser() {}

//called with: user_id
//returns success or error
//request needs to come from admin (not implemented)
//---------------------------------------------------------------------
//BanUser(id, liftBanDate)
//TODO logic to automatically unban after liftBanDate

func UnbanUser() {}

//called with: user_id
//returns success or error
//request needs to come from admin (not implemented) or automatic on expiration
//---------------------------------------------------------------------
//UnbanUser(id)
