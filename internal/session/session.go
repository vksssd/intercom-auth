package session

import (
	"context"
	"crypto/rand"
	"encoding/gob"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/nacl/secretbox"
)

var (
	store          *sessions.CookieStore
	sessionName    = "auth-session"
	sessionMaxAge  = 30 * time.Minute // 30 minutes
	cleanupInterval = 1 * time.Hour
	redisClient    *redis.Client
	secretKey      = []byte("your-secret-key")
	mutex          sync.Mutex
)

func init() {
	store = sessions.NewCookieStore(secretKey)
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   int(sessionMaxAge.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteDefaultMode,
	}
	gob.Register(map[string]interface{}{})
	gob.Register(time.Time{})

	// Initialize Redis Client
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Update with your Redis server address
	})
}

func Get(r *http.Request, sessionName string) (*sessions.Session, error) {
	session, err := store.Get(r, sessionName)
	if err != nil {
		log.Printf("Error getting session: %v", err)
		return nil, err
	}

	// if err = Decrypt(session); err != nil {
	// 	log.Printf("Error decrypting session: %v", err)
	// 	return nil, err
	// }

	// log.Println(session)

	return session, nil
}

func Save(w http.ResponseWriter, r *http.Request, session *sessions.Session) error {
	mutex.Lock()
	defer mutex.Unlock()

	// if err := Encrypt(session); err != nil {
	// 	log.Printf("Error encrypting session: %v", err)
	// 	return err
	// }

	// log.Printf("Session Values before saving: %+v", session.Values)

	err := session.Save(r, w)
	if err != nil {
		log.Printf("Error saving session: %v", err)
	}


	// log.Printf("Session Values after saving: %+v", session.Values)


	return err
}

func Encrypt(session *sessions.Session) error {
	for key, value := range session.Values {
		encryptedValue, err := encryptValue(value)
		if err != nil {
			log.Printf("Error encrypting session value: %v", err)
			return err
		}
		session.Values[key] = encryptedValue
	}
	return nil
}

func Decrypt(session *sessions.Session) error {
	for key, value := range session.Values {
		decryptedValue, err := decryptValue(value)
		if err != nil {
			log.Printf("Error decrypting session value: %v", err)
			return err
		}
		session.Values[key] = decryptedValue
	}
	return nil
}

func encryptValue(value interface{}) ([]byte, error) {
	plaintext, ok := value.(string)
	if !ok {
		return nil, errors.New("invalid value type")
	}

	var nonce [24]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, err
	}

	encrypted := secretbox.Seal(nonce[:], []byte(plaintext), &nonce, convertTo32ByteSlice(secretKey))
	return encrypted, nil
}

func decryptValue(value interface{}) ([]byte, error) {
	encrypted, ok := value.([]byte)
	if !ok {
		return nil, errors.New("invalid value type")
	}

	var nonce [24]byte
	copy(nonce[:], encrypted[:24])
	decrypted, ok := secretbox.Open(nil, encrypted[24:], &nonce, convertTo32ByteSlice(secretKey))
	if !ok {
		return nil, errors.New("decryption error")
	}

	return decrypted, nil
}

func CleanupExpiredSessions() {
	log.Println("Cleaning up expired sessions")
	// Implement cleanup logic here
}

func ScheduleSessionCleanup(ctx context.Context) {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			CleanupExpiredSessions()
		case <-ctx.Done():
			log.Println("Stopping session cleanup")
			return
		}
	}
}

func Configure(name string, maxAge time.Duration, secure bool, secret []byte) {
	mutex.Lock()
	defer mutex.Unlock()

	sessionName = name
	sessionMaxAge = maxAge
	secretKey = secret

	store = sessions.NewCookieStore(secretKey)
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   int(maxAge.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	}

	log.Println("Session configuration updated")
}

func convertTo32ByteSlice(key []byte) *[32]byte {
	var convertedKey [32]byte
	copy(convertedKey[:], key)
	return &convertedKey
}

