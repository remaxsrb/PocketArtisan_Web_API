package validators

import "regexp"

var emailRegexp = regexp.MustCompile(
	`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9-]+(?:\.[a-zA-Z0-9-]{2,})+$`,
)

func IsValidEmail(email string) bool {
	return emailRegexp.MatchString(email)
}
