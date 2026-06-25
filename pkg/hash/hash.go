// Package hash menyediakan utilitas hashing password menggunakan bcrypt.
package hash

import "golang.org/x/crypto/bcrypt"

// Hash menghasilkan bcrypt hash dari password plaintext.
func Hash(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Check membandingkan password plaintext dengan hash. Mengembalikan true bila cocok.
func Check(password, hashed string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)) == nil
}
