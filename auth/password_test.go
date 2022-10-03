package auth

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/bcrypt"
)

var _ = Describe("Password Tests", func() {

	// Tests that, if the GenerateFromPassword function returns an error, then calling the
	// HashPassword function will return an error
	It("HashPassword - GenerateFromPassword fails - Error", func() {

		// First, ensure that the cost is set to an invalid value and then reset it to the default
		// cost when the test finishes
		Cost = 60
		defer func() {
			Cost = bcrypt.DefaultCost
		}()

		// Next, attempt to hash the cleartext password; this should fail
		hashed, err := HashPassword("test_password_ftw")

		// Finally, verify the details of the error and password
		Expect(hashed).Should(BeEmpty())
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).Should(Equal("crypto/bcrypt: cost 60 is outside allowed range (4,31)"))
	})

	// Tests that, if no error occurs, then calling the HashPassword function will return the
	//  hashed and salted version of the password as a string
	It("HashPassword - No errors - Hashed value returned", func() {

		// Attempt to hash the cleartext password; this should fail
		hashed, err := HashPassword("test_password_ftw")
		Expect(err).ShouldNot(HaveOccurred())

		// Verify the hashed password
		fmt.Printf("Length: %d\n", len(hashed))
		Expect(hashed).ShouldNot(BeEmpty())
		Expect(hashed).Should(HaveLen(61))
		Expect(hashed[60]).Should(Equal(uint8(0)))
	})

	// Tests that, if the CompareHashAndPassword function returns an error, then calling the
	// VerifyPassword function will return false and an error indicating why the failure occurred
	It("VerifyPassword - CompareHashAndPassword failed - False, error returned", func() {

		// Attempt to match the cleartext password against the hashed value; this should fail
		matched, err := VerifyPassword("derp", "test_password_ftw")

		// Verify the match value and the error
		Expect(matched).Should(BeFalse())
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).Should(Equal("crypto/bcrypt: hashedSecret too short to be a bcrypted password"))
	})

	// Tests that, if the hashed value of the password does not match the hash value submitted to the
	// VerifyPassword function, then an error and false will be returned
	It("VerifyPassword - Hash value mismatches - False, error returned", func() {

		// Attempt to match the cleartext password against the hashed value; this should fail
		matched, err := VerifyPassword("$2a$10$HnehWrqzxX9RKjrVYzhNkuoZqaWRmlwxk/vjh9BpiUDvtvEQ.mq6u\x00", "test_password_wtf")

		// Verify the match value and the error
		Expect(matched).Should(BeFalse())
		Expect(err).ShouldNot(HaveOccurred())
	})

	// Tests that, if the hash of the cleartext password matches the hashed value, but that the
	// version bit does not match, then an error and true will be returned
	It("VerifyPassword - Version not matched - True, error returned", func() {

		// Attempt to match the cleartext password against the hashed value; this should fail
		matched, err := VerifyPassword("$2a$10$HnehWrqzxX9RKjrVYzhNkuoZqaWRmlwxk/vjh9BpiUDvtvEQ.mq6u\x01", "test_password_ftw")

		// Verify the match value and the error
		Expect(matched).Should(BeTrue())
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).Should(Equal("Version mismatch"))
	})
})
