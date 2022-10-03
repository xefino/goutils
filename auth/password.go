package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// Current version of the hashing algorithm; update this when the algorithm is changed
var Current byte = 0

// Cost defines the cost to use when hashing passwords
var Cost int = bcrypt.DefaultCost

// VersionMismatch is the error that is returned when a password matches but the version of the
// hashing algorithm is different from the one that generated the password
var VersionMismatch = errors.New("Version mismatch")

// HashPassword hashes a cleartext password to its hashed and salted value
func HashPassword(cleartext string) (string, error) {

	// Use GenerateFromPassword to hash & salt cleartext.
	// MinCost is just an integer constant provided by the bcrypt
	// package along with DefaultCost & MaxCost.
	// The cost can be any value you want provided it isn't lower
	// than the MinCost (4)
	hash, err := bcrypt.GenerateFromPassword([]byte(cleartext), Cost)
	if err != nil {
		return "", err
	}

	// Append a version bit to our hashed password to make the result harder
	// to brute force and to allow us to version our passwords
	hash = append(hash, Current)

	// GenerateFromPassword returns a byte slice so we need to
	// convert the bytes to a string and return it
	return string(hash), nil
}

// VerifyPassword checks that the hashed version of a password matches its cleartext value. In the case
// of a match, true will be returned. In the case of a mismatch, false will be returned. If false is
// returned, an error is guaranteed. However, even if true is returned, an error may be present if the
// algorithm version that generated the password is different from the current version
func VerifyPassword(hashed string, cleartext string) (bool, error) {

	// First, get the actual hash of the password and the version byte from the hash value
	hashedAsBytes := []byte(hashed)
	hashedActual, version := hashedAsBytes[:len(hashedAsBytes)-1], hashedAsBytes[len(hashedAsBytes)-1]

	// Next, compare the hashed password to the cleartext; if this returns an error then the
	// password does not match the value provided so return false and the resulting error
	if err := bcrypt.CompareHashAndPassword(hashedActual, []byte(cleartext)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}

		return false, err
	}

	// Now, check if the version on the password matches the current version. If it doesn't then
	// we'll return a version mismatch error which will be used to let the user know that they need
	// to reset their password
	if version != Current {
		return true, VersionMismatch
	}

	// Finally, since the password value and version match, we can return true and no error
	return true, nil
}
