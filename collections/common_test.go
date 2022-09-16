package collections

import (
	"strconv"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Create a new test runner we'll use to test all the
// modules in the collections package
func TestCollections(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Collections Suite")
}

var _ = Describe("Common Tests", func() {

	// Tests that the Keys function returns all the keys in the mapping
	It("Keys - Works", func() {

		// Create our test mapping
		mapping := map[string]int{"A": 0, "B": 42, "C": -5}

		// Get the keys from the mapping
		keys := Keys(mapping)

		// Verify the list of keys that was returned
		Expect(keys).Should(HaveLen(3))
		Expect(keys).Should(ConsistOf("A", "B", "C"))
	})

	// Tests that the Values function returns all the values in the mapping
	It("Values - Works", func() {

		// Create our test mapping
		mapping := map[string]int{"A": 0, "B": 42, "C": -5}

		// Get the values from the mapping
		values := Values(mapping)

		// Verify the list of values that was returned
		Expect(values).Should(HaveLen(3))
		Expect(values).Should(ConsistOf(0, 42, -5))
	})

	// Tests that calling ToDictionary with an empty map will result in a panic
	It("ToDictionary - Map is nil - Panic", func() {

		// Attempt to convert the list to a map; this should panic
		list := []int{1, 24, 3, 5}
		Expect(func() {
			ToDictionary(nil, list, func(index int, i int) string {
				return strconv.FormatInt(int64(i), 10)
			}, true)
		}).Should(Panic())
	})

	// Tests that calling ToDictionary with a nil list will
	// result in no change to the map
	It("ToDictionary - List is nil - No work done", func() {

		// Attempt to convert a nil list to a map; nothing should happen
		mapping := make(map[string]int)
		ToDictionary(mapping, nil, func(index int, i int) string {
			return strconv.FormatInt(int64(i), 10)
		}, true)

		// Verify that the map is still empty
		Expect(mapping).Should(BeEmpty())
	})

	// Tests that, if the ToDictionary function is called with a list that
	// contains elements resulting in collisions and that the overwrite
	// flag is true, then the map will contain the newer value
	It("ToDictionary - Overwrite true - Collisions overwritten", func() {

		// First, create our test list
		list := []int{1, 24, 3, 5, 16}

		// Next, convert the list to a map
		mapping := make(map[string]int)
		ToDictionary(mapping, list, func(index int, i int) string {
			if i == 5 {
				return "3"
			}

			return strconv.FormatInt(int64(i), 10)
		}, true)

		// Finally, verify the values in the map
		Expect(mapping).Should(HaveLen(4))
		Expect(mapping["1"]).Should(Equal(1))
		Expect(mapping["3"]).Should(Equal(5))
		Expect(mapping["16"]).Should(Equal(16))
		Expect(mapping["24"]).Should(Equal(24))
	})

	// Tests that, if the ToDictionary function is called with a list that
	// contains elements resulting in collisions and that the overwrite
	// flag is false, then the map will contain the older value
	It("ToDictionary - Overwrite false - Collisions ignored", func() {

		// First, create our test list
		list := []int{1, 24, 3, 5, 16}

		// Next, convert the list to a map
		mapping := make(map[string]int)
		ToDictionary(mapping, list, func(index int, i int) string {
			if i == 5 {
				return "3"
			}

			return strconv.FormatInt(int64(i), 10)
		}, false)

		// Finally, verify the values in the map
		Expect(mapping).Should(HaveLen(4))
		Expect(mapping["1"]).Should(Equal(1))
		Expect(mapping["3"]).Should(Equal(3))
		Expect(mapping["16"]).Should(Equal(16))
		Expect(mapping["24"]).Should(Equal(24))
	})

	// Tests that calling ToDictionaryKeys with an empty map will result in a panic
	It("ToDictionaryKeys - Map is nil - Panic", func() {

		// Attempt to convert the list to a map; this should panic
		list := []int{1, 24, 3, 5}
		Expect(func() {
			ToDictionaryKeys(nil, list, func(index int, i int) string {
				return strconv.FormatInt(int64(i), 10)
			}, true)
		}).Should(Panic())
	})

	// Tests that calling ToDictionaryKeys with a nil list will
	// result in no change to the map
	It("ToDictionaryKeys - List is nil - No work done", func() {

		// Attempt to convert a nil list to a map; nothing should happen
		mapping := make(map[int]string)
		ToDictionaryKeys(mapping, nil, func(index int, i int) string {
			return strconv.FormatInt(int64(i), 10)
		}, true)

		// Verify that the map is still empty
		Expect(mapping).Should(BeEmpty())
	})

	// Tests that, if the ToDictionaryKeys function is called with a list that
	// contains elements resulting in collisions and that the overwrite
	// flag is true, then the map will contain the newer value
	It("ToDictionaryKeys - Overwrite true - Collisions overwritten", func() {

		// First, create our test list
		list := []int{1, 24, 3, 3, 16}

		// Next, convert the list to a map
		mapping := make(map[int]string)
		ToDictionaryKeys(mapping, list, func(index int, i int) string {
			if index == 3 {
				return "5"
			}

			return strconv.FormatInt(int64(i), 10)
		}, true)

		// Finally, verify the values in the map
		Expect(mapping).Should(HaveLen(4))
		Expect(mapping[1]).Should(Equal("1"))
		Expect(mapping[3]).Should(Equal("5"))
		Expect(mapping[16]).Should(Equal("16"))
		Expect(mapping[24]).Should(Equal("24"))
	})

	// Tests that, if the ToDictionaryKeys function is called with a list that
	// contains elements resulting in collisions and that the overwrite
	// flag is false, then the map will contain the older value
	It("ToDictionaryKeys - Overwrite false - Collisions ignored", func() {

		// First, create our test list
		list := []int{1, 24, 3, 3, 16}

		// Next, convert the list to a map
		mapping := make(map[int]string)
		ToDictionaryKeys(mapping, list, func(index int, i int) string {
			if index == 3 {
				return "5"
			}

			return strconv.FormatInt(int64(i), 10)
		}, false)

		// Finally, verify the values in the map
		Expect(mapping).Should(HaveLen(4))
		Expect(mapping[1]).Should(Equal("1"))
		Expect(mapping[3]).Should(Equal("3"))
		Expect(mapping[16]).Should(Equal("16"))
		Expect(mapping[24]).Should(Equal("24"))
	})

	// Tests that calling Index with a nil list will result in no change to the map
	It("Index - List is nil - No work done", func() {
		mapping := Index[string](nil)
		Expect(mapping).Should(BeEmpty())
	})

	// Tests that, if the Index function is called with a list that contains elements
	// will result in an index mapping being returned for that list
	It("Index - List not empty - Converted", func() {

		// First, create our test list
		list := []int{1, 24, 3, 16}

		// Next, convert the list to a map
		mapping := Index(list)

		// Finally, verify the values in the map
		Expect(mapping).Should(HaveLen(4))
		for i, item := range list {
			Expect(mapping[item]).Should(Equal(i))
		}
	})

	// Tests that calling IndexWithFunction with a nil list will result in an emtpy map
	It("IndexWithFunction - List is nil - No work done", func() {
		mapping := IndexWithFunction(nil, func(i int) string {
			return strconv.FormatInt(int64(i), 10)
		})

		Expect(mapping).Should(BeEmpty())
	})

	// Tests that, if the IndexWithFunction is called with a list that contains elements
	// then a mapping will be returned with the converted items as keys and the index of
	// the original items as values
	It("IndexWithFunction - List not empty - Converted", func() {

		// First, create our test list
		list := []int{1, 24, 3, 16}

		// Next, convert the list to a map
		mapping := IndexWithFunction(list, func(i int) string {
			return strconv.FormatInt(int64(i), 16)
		})

		// Verify the state of the mapping
		Expect(mapping).Should(HaveLen(4))
		Expect(mapping["1"]).Should(Equal(0))
		Expect(mapping["18"]).Should(Equal(1))
		Expect(mapping["3"]).Should(Equal(2))
		Expect(mapping["10"]).Should(Equal(3))
	})

	// Tests that, if the AsSlice function is called with no data, then an
	// empty list will be returned
	It("AsSlice - No data provided - Empty list returned", func() {
		list := AsSlice[int]()
		Expect(list).Should(BeEmpty())
	})

	// Tests that, if the AsSlice function is called with data, then that data
	// will be added to a new slice of the same length that respects the ordering
	// of the data provided
	It("AsSlice - Data provided - Returned as list", func() {
		list := AsSlice(1, 2, 3, 10)
		Expect(list).Should(HaveLen(4))
		Expect(list).Should(Equal([]int{1, 2, 3, 10}))
	})

	// Tests that the Convert function will produce an empty list if called with no data
	It("Convert - No data provided - Empty list returned", func() {

		// Attempt to do the conversion with an empty list
		list := Convert(func(i int) string {
			return strconv.FormatInt(int64(i), 10)
		})

		// Verify that the list is empty
		Expect(list).Should(BeEmpty())
	})

	// Tests that the Convert function will produce a list of data where each item is the
	// result of an input to the convert function provided
	It("Convert - Data provided - Converted", func() {

		// Attempt to do the conversion with a non-empty list
		list := Convert(func(i int) string {
			return strconv.FormatInt(int64(i), 10)
		}, 1, 2, 3, 10)

		// Verify the converted data
		Expect(list).Should(HaveLen(4))
		Expect(list).Should(Equal([]string{"1", "2", "3", "10"}))
	})

	// Tests that, if the ConvertDictionary function is called with a value of nil for the
	// list, then an empty map will be returned
	It("ConvertDictionary - Nil - Works", func() {
		mapping := ConvertDictionary(nil, func(item int) (string, int) {
			return strconv.FormatInt(int64(item), 10), item
		})

		Expect(mapping).Should(BeEmpty())
	})

	// Tests that, if the ConvertDictionary function is called with a non-nil list, then
	// the conversion function will be called for each item and the result will be added
	// to the mappping that is then returned
	It("ConvertDictionary - List not nil - Works", func() {
		mapping := ConvertDictionary([]int{1, 2, 42}, func(item int) (string, int) {
			return strconv.FormatInt(int64(item), 10), item
		})

		Expect(mapping).Should(HaveLen(3))
		Expect(mapping).Should(HaveKey("1"))
		Expect(mapping["1"]).Should(Equal(1))
		Expect(mapping).Should(HaveKey("2"))
		Expect(mapping["2"]).Should(Equal(2))
		Expect(mapping).Should(HaveKey("42"))
		Expect(mapping["42"]).Should(Equal(42))
	})

	// Tests the conditions describing how the Contains function will operate
	DescribeTable("Contains - Conditions",
		func(list []int, value int, found bool) {
			Expect(Contains(list, value)).Should(Equal(found))
		},
		Entry("List is nil - False", nil, 42, false),
		Entry("Value not in list - False", []int{1, 2, 4, 8, 16}, 42, false),
		Entry("Value in list - True", []int{2, 10, 42, 99, 420}, 42, true))

	// Tests the conditions describing how the ContainsFunc function will operate
	DescribeTable("ContainsFunc - Conditions",
		func(list []int, found bool) {
			Expect(ContainsFunc(list, func(i int) bool {
				return i%42 == 1
			})).Should(Equal(found))
		},
		Entry("List is nil - False", nil, false),
		Entry("Checker returns false for all items - False", []int{0, 2, 41, 42, 83}, false),
		Entry("Checker returns true for one item - True", []int{0, 2, 41, 42, 43, 83}, true))
})
