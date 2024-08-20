package Tools

import (
	"math/rand"
	"sync"
	"time"
)

const ALPHA_NUMERIC_SPECIAL = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+"
const ALPHA_NUMERIC = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const ALPHA = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const NUMERIC = "0123456789"

func GetSystemTime() int64 {
	return time.Now().UnixNano()
}

type Randomizer struct {
	source rand.Source
	seed   int64
	mutex  *sync.Mutex
}

func NewRandomizer(seed int64) *Randomizer {
	return &Randomizer{
		source: rand.NewSource(seed),
		seed:   seed,
		mutex:  &sync.Mutex{},
	}
}

// Generates a random string of a given length and alphabet based on the randomizer seed.
// Two calls to the same randomizer with the same parameters will usually not generate the same result.
func (rand *Randomizer) GenerateRandomString(length uint32, validChars string) string {
	str := ""
	for i := uint32(0); i < length; i++ {
		rand.mutex.Lock()
		str += string(validChars[rand.source.Int63()%int64(len(validChars))])
		rand.mutex.Unlock()
	}
	return str
}

// Generates a random number between min and max based on the randomizer seed.
// Two calls to the same randomizer with the same parameters will usually not generate the same result.
func (rand *Randomizer) GenerateRandomNumber(min, max int64) int64 {
	rand.mutex.Lock()
	number := rand.source.Int63()%int64(max-min+1) + min
	rand.mutex.Unlock()
	return number
}

// Generates a random string of a given length and alphabet based on a combination of the randomizer seed and the seed provided.
// Two calls to the same randomizer with the same parameters will always generate the same result.
func (randomizer *Randomizer) GenerateRandomStringWithSeed(length uint32, validChars string, seed int64) string {
	str := ""
	for i := uint32(0); i < length; i++ {
		str += string(validChars[(rand.NewSource(randomizer.seed+int64(seed)).Int63())%int64(len(validChars))])
	}
	return str

}

// Uses a combination of the randomizer seed and the seed provided.
// Two calls to the same randomizer with the same parameters will always generate the same result.
func (randomizer *Randomizer) GenerateRandomNumberWithSeed(min, max int64, seed int64) int64 {
	return rand.NewSource(randomizer.seed+int64(seed)).Int63()%int64(max-min+1) + min
}

// GenerateRandomString generates a random string of a given length and alphabet.
// Seed is the current system time.
// Two calls to this function will usually not generate the same result.
func GenerateRandomString(length uint32, alphabet string) string {
	return NewRandomizer(GetSystemTime()).GenerateRandomString(length, alphabet)
}

// GenerateRandomNumber generates a random number between min and max.
// Seed is the current system time.
// Two calls to this function will usually not generate the same result.
func GenerateRandomNumber(min, max int64) int64 {
	return NewRandomizer(GetSystemTime()).GenerateRandomNumber(min, max)
}
