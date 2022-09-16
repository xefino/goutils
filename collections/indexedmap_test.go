package collections

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Indexed Map Tests", func() {

	// Tests that, when calling the Add function with the overwrite
	// flag set to true, if a collision occurs then the item being
	// added will take precedence over the existing item
	It("Add - Overwrite true - Collisions overwritten", func() {

		// First, create a new indexed map
		dict := NewIndexedMap[string, int]()

		// Next, add two new entries associated with the same key
		// with overwrite set to true
		dict.Add("derp", 1, true)
		dict.Add("derp", 2, true)

		// Finally, verify that the newer entry is in the dictionary
		// and that all the data exists in the dictionary
		Expect(dict.data).Should(HaveLen(1))
		Expect(dict.data[0]).Should(Equal(2))
		Expect(dict.indexes).Should(HaveLen(1))
		Expect(dict.indexes["derp"]).Should(Equal(0))
	})

	// Tests that, when calling the Add function with the overwrite
	// flag set to false, if a collision occurs then the existing item
	// will take precedence over the item being added
	It("Add - Overwrite false - Collisions ignored", func() {

		// First, create a new indexed map
		dict := NewIndexedMap[string, int]()

		// Next, add two new entries associated with the same key
		// with overwrite set to false
		dict.Add("derp", 1, false)
		dict.Add("derp", 2, false)

		// Finally, verify that the newer entry is in the dictionary
		// and that all the data exists in the dictionary
		Expect(dict.data).Should(HaveLen(1))
		Expect(dict.data[0]).Should(Equal(1))
		Expect(dict.indexes).Should(HaveLen(1))
		Expect(dict.indexes["derp"]).Should(Equal(0))
	})

	// Tests the conditions determining how the Exists function will return;
	// true if the key is associated with an entry in the dictionary, false otherwise
	DescribeTable("Exists - Conditions",
		func(key string, exists bool) {

			// First, create a new indexed map
			dict := NewIndexedMap[string, int]()

			// Next, add some entries to the map
			dict.Add("derp", 1, false)
			dict.Add("herp", 2, false)
			dict.Add("sherbert", 3, false)

			// Finally, check if the key is in the mapping
			Expect(dict.Exists(key)).Should(Equal(exists))
		},
		Entry("Key 1 in dictionary - True", "derp", true),
		Entry("Key 2 in dictionary - True", "herp", true),
		Entry("Key 3 in dictionary - True", "sherbert", true),
		Entry("Key not in dictionary - False", "rainbow", false))

	// Tests that, if the Get function is called with a key that is
	// not associated with any item in the list, then the zero value
	// for that type and false will be returned
	It("Get - Not found - False returned", func() {

		// First, create a new indexed map
		dict := NewIndexedMap[string, int]()

		// Next, add some entries to the map
		dict.Add("derp", 1, false)
		dict.Add("herp", 2, false)
		dict.Add("sherbert", 3, false)

		// Now, attempt to retrieve an item from the map that
		// is not associated with any key in the map
		item, found := dict.Get("rainbow")

		// Verify that the item wasn't found
		Expect(item).Should(BeZero())
		Expect(found).Should(BeFalse())
	})

	// Tests that, if the Get function is called with a key that is
	// associated with an item in the list, then the value of that
	// item and true will be returned
	It("Get - Found - Returned", func() {

		// First, create a new indexed map
		dict := NewIndexedMap[string, int]()

		// Next, add some entries to the map
		dict.Add("derp", 1, false)
		dict.Add("herp", 2, false)
		dict.Add("sherbert", 3, false)

		// Now, attempt to retrieve an item from the map that
		// is associated with one of the keys in the map
		item, found := dict.Get("herp")

		// Verify that the item was found
		Expect(item).Should(Equal(2))
		Expect(found).Should(BeTrue())
	})

	// Tests that, if the Remove function is called with a key that is
	// not associated with any item in the list, then false will be returned
	It("Remove - Not found - False returned", func() {

		// First, create a new indexed map
		dict := NewIndexedMap[string, int]()

		// Next, add some entries to the map
		dict.Add("derp", 1, false)
		dict.Add("herp", 2, false)
		dict.Add("sherbert", 3, false)

		// Now, attempt to remove an item from the map that
		// is not associated with any key in the map
		found := dict.Remove("rainbow")

		// Finally, verify that the item wasn't found
		Expect(found).Should(BeFalse())
	})

	// Tests that, if the Remove function is called with a key that is
	// associated with an item in the list, then true will be returned and
	// the item will be removed from the list
	It("Remove - Found - Removed", func() {

		// First, create a new indexed map
		dict := NewIndexedMap[string, int]()

		// Next, add some entries to the map
		dict.Add("derp", 1, false)
		dict.Add("herp", 2, false)
		dict.Add("sherbert", 3, false)

		// Now, attempt to remove an item from the map that
		// is associated with one of the keys in the map
		found := dict.Remove("herp")
		Expect(found).Should(BeTrue())

		// Finally, verify the data in the indexed dictionary
		Expect(dict.indexes).Should(HaveLen(2))
		Expect(dict.indexes).Should(HaveKey("derp"))
		Expect(dict.indexes["derp"]).Should(Equal(0))
		Expect(dict.indexes).Should(HaveKey("sherbert"))
		Expect(dict.indexes["sherbert"]).Should(Equal(1))
		Expect(dict.data).Should(Equal([]int{1, 3}))
	})

	// Tests that calling the Keys function will return all the keys
	// associated with entries in the dictionary
	It("Keys - Works", func() {

		// First, create a new indexed map
		dict := NewIndexedMap[string, int]()

		// Next, add some entries to the map
		dict.Add("derp", 1, false)
		dict.Add("herp", 2, false)
		dict.Add("sherbert", 3, false)

		// Now, get the keys from the dictionary
		keys := dict.Keys()

		// Finally, verify the list of keys
		Expect(keys).Should(HaveLen(3))
		Expect(keys).Should(ContainElements("derp", "herp", "sherbert"))
	})

	// Tests that calling the Data function will return all the data
	// associated with entries in the dictionary
	It("Data - Works", func() {

		// First, create a new indexed map
		dict := NewIndexedMap[string, int]()

		// Next, add some entries to the map
		dict.Add("derp", 1, false)
		dict.Add("herp", 2, false)
		dict.Add("sherbert", 3, false)

		// Now, get the data from the dictionary
		data := dict.Data()

		// Finally, verify the list of keys
		Expect(data).Should(HaveLen(3))
		Expect(data).Should(ContainElements(1, 2, 3))
	})

	// Tests that calling the Length function will return the number
	// of entries in the dictionary
	It("Length - Works", func() {

		// First, create a new indexed map
		dict := NewIndexedMap[string, int]()

		// Next, add some entries to the map
		dict.Add("derp", 1, false)
		dict.Add("herp", 2, false)
		dict.Add("sherbert", 3, false)

		// Finally, verify the length of the dictionary
		Expect(dict.Length()).Should(Equal(3))
	})

	// Tests that, if the ForEach function is called and the inner
	// function never returns false, then it will iterate over all
	// items in the indexed map
	It("ForEach - Inner returns true - All elements returned", func() {

		// First, create a new indexed map
		dict := NewIndexedMap[string, int]()

		// Next, add some entries to the map
		dict.Add("derp", 1, false)
		dict.Add("herp", 2, false)
		dict.Add("sherbert", 3, false)

		// Now, iterate over all the items in the indexed map
		// and do some function on each
		keys := make([]string, 3)
		total := 0
		dict.ForEach(func(key string, value int) bool {
			keys[value-1] = key
			total += value
			return true
		})

		// Finally, verify the output we would expect if all the
		// elements had been operated on
		Expect(keys).Should(Equal([]string{"derp", "herp", "sherbert"}))
		Expect(total).Should(Equal(6))
	})

	// Tests that, if the ForEach function is called and the inner
	// function returns false, then it will stop iterating at that point
	It("ForEach - Inner returns false - Not all elements returned", func() {

		// First, create a new indexed map
		dict := NewIndexedMap[string, int]()

		// Next, add some entries to the map
		dict.Add("derp", 1, false)
		dict.Add("herp", 2, false)
		dict.Add("sherbert", 3, false)

		// Now, iterate over all the items in the indexed map
		// and do some function on each
		keys := make([]string, 3)
		total := 0
		dict.ForEach(func(key string, value int) bool {
			if value == 2 {
				return false
			}

			keys[value-1] = key
			total += value
			return true
		})

		// Finally, verify the output we would expect if the
		// function short-circuited
		possibilities := map[int][]string{
			0: {"", "", ""},
			1: {"derp", "", ""},
			4: {"derp", "", "sherbert"}}
		Expect(possibilities).Should(HaveKey(total))
		Expect(possibilities[total]).Should(Equal(keys))
	})
})
