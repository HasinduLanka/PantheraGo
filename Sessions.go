package main

import (
	"crypto/rand"
	"strconv"
	"time"
)

type SessionManager struct {
	Sessions             map[string]*Session
	DefaultSessionStash  StashedSession
	ComponentSrcs        map[string]*ComponentSrc
	ComponentSrcProvider func(string) *ComponentSrc
}

func (sm *SessionManager) Make() {
	sm.ComponentSrcProvider = func(src string) *ComponentSrc {
		C, found := sm.ComponentSrcs[src]
		if !found {
			return nil
		}
		return C
	}
}

// Returns Session, IsSessionNew
func (sm *SessionManager) Get(ID string, pass string) *Session {
	S, found := sm.Sessions[ID]
	if found {
		if S.Pass == pass {
			return S
		}
	}
	return nil
}

func (sm *SessionManager) NewSession() *Session {
	S := sm.DefaultSessionStash.Summon(sm.ComponentSrcProvider)
	S.ID, S.Pass = sm.CreateSessionID()
	sm.Sessions[S.ID] = S
	return S
}

func (sm *SessionManager) CreateSessionID() (string, string) {
	ID := secureRandomAlphaString(16) + "-" + strconv.FormatInt(time.Now().Unix(), 36)
	Pass := secureRandomAlphaString(48)

	return ID, Pass
}

// func componentSrcProvider(src string) *ComponentSrc {

// }

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ" // 52 possibilities
	letterIdxBits = 6                                                      // 6 bits to represent 64 possibilities / indexes
	letterIdxMask = 1<<letterIdxBits - 1                                   // All 1-bits, as many as letterIdxBits
)

func secureRandomAlphaString(length int) string {

	const lenletterBytes int = len(letterBytes)

	result := make([]byte, length)
	bufferSize := int(float64(length) * 1.3)
	for i, j, randomBytes := 0, 0, []byte{}; i < length; j++ {
		if j%bufferSize == 0 {
			randomBytes = secureRandomBytes(bufferSize)
		}
		if idx := int(randomBytes[j%length] & letterIdxMask); idx < lenletterBytes {
			result[i] = letterBytes[idx]
			i++
		}
	}

	return string(result)
}

// SecureRandomBytes returns the requested number of bytes using crypto/rand
func secureRandomBytes(length int) []byte {
	var randomBytes = make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		println("Unable to generate random bytes")
		return []byte{}
	}
	return randomBytes
}

// var rndsrc = rand.NewSource(time.Now().UnixNano())

// const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
// const (
// 	letterIdxBits = 6                    // 6 bits to represent a letter index
// 	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
// 	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
// )

// func randomString(n int) string {
// 	b := make([]byte, n)
// 	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
// 	for i, cache, remain := n-1, rndsrc.Int63(), letterIdxMax; i >= 0; {
// 		if remain == 0 {
// 			cache, remain = rndsrc.Int63(), letterIdxMax
// 		}
// 		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
// 			b[i] = letterBytes[idx]
// 			i--
// 		}
// 		cache >>= letterIdxBits
// 		remain--
// 	}

// 	return *(*string)(unsafe.Pointer(&b))
// }
