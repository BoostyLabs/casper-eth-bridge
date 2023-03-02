package hexutils

import "strings"

const (
	// accountHashPrefix defines account hash prefix.
	accountHashPrefix = "account-hash"
	// hashPrefix defiles hash prefix.
	hashPrefix = "hash"
)

// isHexCharacter returns bool of c being a valid hexadecimal.
func isHexCharacter(c byte) bool {
	return ('0' <= c && c <= '9') || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F')
}

// Has0xPrefix validates str begins with '0x' or '0X'.
func Has0xPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X')
}

// Has00Prefix validates str begins with '00'.
func Has00Prefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && str[1] == '0'
}

// HasAccountHashPrefix validates str begins with 'account-hash'.
func HasAccountHashPrefix(str string) bool {
	return strings.Contains(str, accountHashPrefix) || strings.Contains(strings.ToLower(str), accountHashPrefix)
}

// HasHashPrefix validates str begins with 'hash-'.
func HasHashPrefix(str string) bool {
	return strings.Contains(str, hashPrefix) || strings.Contains(strings.ToLower(str), hashPrefix)
}

// HasTagPrefix validates str begins with '01' or '02'.
func HasTagPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && (str[1] == '1' || str[1] == '2')
}

// IsHex validates whether each byte is valid hexadecimal string, ignores '0x' Or '0X' prefix.
func IsHex(str string) bool {
	if Has0xPrefix(str) {
		str = str[2:]
	}

	if len(str) == 0 || len(str)%2 != 0 {
		return false
	}
	for _, c := range []byte(str) {
		if !isHexCharacter(c) {
			return false
		}
	}
	return true
}

// ToHexString converts to an even-length string, appending a leading zero if necessary.
func ToHexString(str string) string {
	if len(str)%2 == 0 {
		return str
	}

	return "0" + str
}
