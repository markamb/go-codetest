package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"testing"
)

type hashTestCase struct {
	hashVal uint32
	str     string
}

var SamplesHashCodes = []hashTestCase{
	{2166136261, ""},
	{3826002220, "a"},
	{3289118412, "A"},
	{735387639, "AA"},
	{2240196237, "ACB"},
	{1552166763, "ABC"},
	{1552166763, "ABC"}, // duplicate entry
	{19800803, "http://bbc.co.uk/"},
	{3042750490, "https://bbc.co.uk/"},
	{1671006656, "http://bbc.co.uk/index.html"},
	{2956717082, "春节"}, // Unicode characters
	{1569850263, "http://bbc.co.uk/path?data=春节"},
}

func TestMyHash32(t *testing.T) {
	// Test the sample URLS without using the wrapper function
	hasher := CreateMyHash32()
	if hasher.Size() != 4 {
		t.Errorf("Incorrect hash code size, expected %v, got %v", 4, hasher.Size())
	}
	if hasher.BlockSize() != 1 {
		t.Errorf("Incorrect hash code block size, expected %v, got %v", 1, hasher.BlockSize())
	}

	for _, test := range SamplesHashCodes {
		hasher.Reset()
		hasher.Write([]byte(test.str))
		result := hasher.Sum(nil)
		num := binary.BigEndian.Uint32(result)
		if num != test.hashVal {
			t.Errorf("Incorrect hash code for string (%s), expected %v, got %v", test.str, test.hashVal, num)
		}
		if hasher.Sum32() != test.hashVal {
			t.Errorf("Incorrect hash code for string (%s), expected %v, got %v", test.str, test.hashVal, num)
		}
	}

}
func TestHashString(t *testing.T) {

	for _, test := range SamplesHashCodes {
		hc := HashString(test.str)
		if hc != test.hashVal {
			t.Errorf("Incorrect hash code for string (%s), expected %v, got %v", test.str, test.hashVal, hc)
		}
	}
}

// TestURLHashing loads a list of 10K (or so) URLS from file and calculates their hash codes.
// Checks for no errors, "excessive" number of clashes and for a "reasonable" distribution of values
func TestURLHashing(t *testing.T) {

	var out *os.File
	// Uncomment the following line to print some stats on the distribution of the has values
	// out = os.Stdout

	// use a hash map to keep count of how many times each hash code is returned
	hashCodes := make(map[uint32]int) // count of occurrences for each code
	const bucketCount = 20
	var hashBuckets [bucketCount]int // count in (large) buckets
	const bucketSize = math.MaxUint32 / bucketCount

	urlFile, err := os.Open(path.Join("testdata", "urls.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer urlFile.Close()

	collisions := 0 // count of number of collisions
	count := 0
	scanner := bufio.NewScanner(urlFile)
	for scanner.Scan() {
		count++
		url := scanner.Text()
		hashVal := HashString(url)
		hashCodes[hashVal]++
		bucket := hashVal / bucketSize
		hashBuckets[bucket]++
		if hashCodes[hashVal] > 1 {
			collisions++
			// this isn't really an error as such - clashes can occur but they should be very rare
			// (and don't happen in the chosen 10,000 test URLs)
			t.Errorf("Duplicate hash value for string (%s): expected %v", url, hashVal)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// check distribution of codes across the buckets
	maxCount := math.MinInt32
	minCount := math.MaxInt32
	fmt.Fprintf(out, "Bucket Sizes: [")
	for _, n := range hashBuckets {
		if n > maxCount {
			maxCount = n
		}
		if n < minCount {
			minCount = n
		}
		fmt.Fprintf(out, "%d ", n)

	}
	fmt.Fprintf(out, "]\n")
	fmt.Fprintf(out, "Min/Max: (%d,%d)\n", minCount, maxCount)
	fmt.Fprintf(out, "Number of collisions from %d URLs: %d\n", count, collisions)
}
