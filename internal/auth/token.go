package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	HospitalID int64 `json:"hospital_id"`
	jwt.RegisteredClaims
}

type Manager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewManager(secret string, ttl time.Duration) *Manager {
	return &Manager{secret: []byte(secret), ttl: ttl, now: time.Now}
}

func (m *Manager) Generate(staffID, hospitalID int64) (string, int64, error) {
	now := m.now()
	claims := Claims{
		HospitalID: hospitalID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(staffID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	return token, int64(m.ttl.Seconds()), err
}

func (m *Manager) Parse(raw string) (int64, int64, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}
		return m.secret, nil
	}, jwt.WithExpirationRequired())
	if err != nil || !token.Valid {
		return 0, 0, errors.New("invalid token")
	}
	staffID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || staffID < 1 || claims.HospitalID < 1 {
		return 0, 0, errors.New("invalid token claims")
	}
	return staffID, claims.HospitalID, nil
}
