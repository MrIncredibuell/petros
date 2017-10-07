package petros

type contextKey string

const requestArgsKey = contextKey("requestArgs")
const currentUserKey = contextKey("currentUser")
