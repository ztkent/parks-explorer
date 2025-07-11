package dashboard

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/google/uuid"
)

const (
	VisitorCookieName = "nps_visitor"
	CookieMaxAge      = 30 * 24 * 60 * 60 // 1 month
)

// generateGUID creates a new GUID for visitors
func generateGUID() string {
	return uuid.New().String()
}

// hashEmail creates a SHA256 hash of the email
func hashEmail(email string) string {
	hasher := sha256.New()
	hasher.Write([]byte(email))
	return hex.EncodeToString(hasher.Sum(nil))
}

// getOrCreateVisitorID gets the visitor ID from cookie or creates a new one
func (s *Dashboard) getOrCreateVisitorID(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie(VisitorCookieName)
	var visitorID string
	if err != nil || cookie.Value == "" {
		visitorID = generateGUID()
	} else {
		visitorID = cookie.Value
	}
	s.setVisitorCookie(w, visitorID)
	return visitorID
}

// setVisitorCookie sets the visitor tracking cookie
func (s *Dashboard) setVisitorCookie(w http.ResponseWriter, visitorID string) {
	cookie := &http.Cookie{
		Name:     VisitorCookieName,
		Value:    visitorID,
		Path:     "/",
		MaxAge:   CookieMaxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, cookie)
}

// updateVisitorIDToEmail updates the visitor ID to email hash when user logs in
func (s *Dashboard) updateVisitorIDToEmail(w http.ResponseWriter, email string) {
	emailHash := hashEmail(email)
	s.setVisitorCookie(w, emailHash)
}

// TagVistorsMiddleware ensures every visitor has a tracking cookie
func (s *Dashboard) TagVistorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.getOrCreateVisitorID(w, r)
		next.ServeHTTP(w, r)
	})
}
