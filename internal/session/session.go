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
	"github.com/vksssd/intercom-auth/config"
	"golang.org/x/crypto/nacl/secretbox"
)

// var (
// 	store          *sessions.CookieStore
// 	sessionName    = "auth-session"
// 	sessionMaxAge  = 30 * time.Minute // 30 minutes
// 	cleanupInterval = 1 * time.Hour
// 	redisClient    *redis.Client
// 	secretKey      = []byte("your-secret-key")
// 	mutex          sync.Mutex
// )

type SessionService struct {
	Store *sessions.CookieStore
	redisClient *redis.Client
	SessionConfig config.SessionConfig
	mutex sync.Mutex
}

func NewSessionService(redisClient redis.Client, cfg config.SessionConfig)(*SessionService, error){
	store := sessions.NewCookieStore([]byte(cfg.Secret))
	store.Options = &sessions.Options{
		Path: "/",
		MaxAge: int( 30 * time.Minute), //update these all according to cfg
		HttpOnly: true,
		Secure: true,
		SameSite: http.SameSiteStrictMode,
	}

	//register for session ecoding/decoding
	gob.Register(map[string]interface{}{})
	gob.Register(time.Time{})

	return &SessionService{
		Store: store,
		SessionConfig: cfg,
		redisClient: &redisClient,
	}, nil

}


func (s *SessionService) Get(r *http.Request, sessionName string) (*sessions.Session, error) {
	session, err := s.Store.Get(r, sessionName)
	if err != nil {
		log.Printf("Error getting session: %v", err)
		return nil, err
	}

	// if err = s.Decrypt(session); err != nil {
	// 	log.Printf("Error decrypting session: %v", err)
	// 	return nil, err
	// }

	// log.Println(session)

	return session, nil
}

func (s *SessionService)Save(w http.ResponseWriter, r *http.Request, session *sessions.Session) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// if err := s.Encrypt(session); err != nil {
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

func (s *SessionService) Encrypt(session *sessions.Session) error {
	for key, value := range session.Values {
		encryptedValue, err := s.encryptValue(value)
		if err != nil {
			log.Printf("Error encrypting session value: %v", err)
			return err
		}
		session.Values[key] = encryptedValue
	}
	return nil
}

func(s *SessionService) Decrypt(session *sessions.Session) error {
	for key, value := range session.Values {
		decryptedValue, err := s.decryptValue(value)
		if err != nil {
			log.Printf("Error decrypting session value: %v", err)
			return err
		}
		session.Values[key] = decryptedValue
	}
	return nil
}

func (s *SessionService)encryptValue(value interface{}) ([]byte, error) {
	plaintext, ok := value.(string)
	if !ok {
		return nil, errors.New("invalid value type")
	}

	var nonce [24]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, err
	}

	encrypted := secretbox.Seal(nonce[:], []byte(plaintext), &nonce, s.convertTo32ByteSlice([]byte(s.SessionConfig.Secret)))
	return encrypted, nil
}

func (s *SessionService)decryptValue(value interface{}) ([]byte, error) {
	encrypted, ok := value.([]byte)
	if !ok {
		return nil, errors.New("invalid value type")
	}

	var nonce [24]byte
	copy(nonce[:], encrypted[:24])
	decrypted, ok := secretbox.Open(nil, encrypted[24:], &nonce, s.convertTo32ByteSlice([]byte(s.SessionConfig.Secret)))
	if !ok {
		return nil, errors.New("decryption error")
	}

	return decrypted, nil
}

func (s *SessionService)CleanupExpiredSessions() {
	log.Println("Cleaning up expired sessions")
	// Implement cleanup logic here
}

func (s *SessionService)ScheduleSessionCleanup(ctx context.Context) {
	ticker := time.NewTicker(s.SessionConfig.CleanUpInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.CleanupExpiredSessions()
		case <-ctx.Done():
			log.Println("Stopping session cleanup")
			return
		}
	}
}

// func (s *SessionService)Configure(name string, maxAge time.Duration, secure bool, secret []byte) {
// 	mutex.Lock()
// 	defer mutex.Unlock()

// 	sessionName = name
// 	sessionMaxAge = maxAge
// 	secretKey = secret

// 	store = sessions.NewCookieStore(secretKey)
// 	store.Options = &sessions.Options{
// 		Path:     "/",
// 		MaxAge:   int(maxAge.Seconds()),
// 		HttpOnly: true,
// 		Secure:   secure,
// 		SameSite: http.SameSiteStrictMode,
// 	}

// 	log.Println("Session configuration updated")
// }

func (s *SessionService)convertTo32ByteSlice(key []byte) *[32]byte {
	var convertedKey [32]byte
	copy(convertedKey[:], key)
	return &convertedKey
}

