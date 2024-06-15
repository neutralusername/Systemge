package Randomizer

import (
	"math/rand"
	"sync"
	"time"
)

type Randomizer struct {
	source rand.Source
	seed   int64
	mutex  *sync.Mutex
}

const ALPHA_NUMERIC_SPECIAL = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+"
const ALPHA_NUMERIC = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const ALPHA = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const NUMERIC = "0123456789"

func New(seed int64) *Randomizer {
	return &Randomizer{
		source: rand.NewSource(seed),
		seed:   seed,
		mutex:  &sync.Mutex{},
	}
}

func (rand *Randomizer) GenerateRandomString(length int, validChars string) string {
	str := ""
	for i := 0; i < length; i++ {
		rand.mutex.Lock()
		str += string(validChars[rand.source.Int63()%int64(len(validChars))])
		rand.mutex.Unlock()
	}
	return str
}

func (rand *Randomizer) GenerateRandomNumber(min, max int) int {
	rand.mutex.Lock()
	number := int(rand.source.Int63()%int64(max-min+1)) + min
	rand.mutex.Unlock()
	return number
}

func (randomizer *Randomizer) GenerateRandomStringWithSeed(length int, validChars string, seed int) string {
	str := ""
	for i := 0; i < length; i++ {
		str += string(validChars[(rand.NewSource(randomizer.seed+int64(seed)).Int63())%int64(len(validChars))])
	}
	return str

}

func (randomizer *Randomizer) GenerateRandomNumberWithSeed(min, max int, seed int) int {
	return int(rand.NewSource(randomizer.seed+int64(seed)).Int63()%int64(max-min+1)) + min
}

func GetSystemTime() int64 {
	return time.Now().UnixNano()
}
