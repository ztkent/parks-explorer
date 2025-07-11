package dashboard

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Implements Google OAuth 2.0 authentication flow:
//
// OAuth Flow:
// - GoogleLoginHandler: Initiates OAuth flow with Google, generates CSRF-protected state token
// - GoogleCallbackHandler: Handles OAuth callback, exchanges code for tokens, creates user session
// - Uses secure HTTP-only cookies for state verification and session management
//
// User Management:
// - createOrUpdateUser: Creates new users or updates existing user info from Google profile
// - getUserInfoFromGoogle: Fetches user profile data from Google's userinfo API
//
// Session Management:
// - createSession: Generates secure session tokens and stores them with OAuth tokens
// - GetCurrentUser: Validates session cookies and retrieves authenticated user
// - LogoutHandler: Cleans up sessions and cookies on logout

type User struct {
	ID        int       `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Username  string    `json:"username" db:"username"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	GoogleID  string    `json:"google_id" db:"google_id"`
	AvatarURL string    `json:"avatar_url" db:"avatar_url"`
}
type GoogleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

var googleOAuthConfig = &oauth2.Config{
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	// Protocol: Use http:// for local development, https:// for production
	RedirectURL: os.Getenv("GOOGLE_REDIRECT_URI"),
	Scopes:      []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	Endpoint:    google.Endpoint,
}

func (s *Dashboard) GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	// Generate random state token
	b := make([]byte, 32)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)

	// Store state in session cookie
	secure := true
	if os.Getenv("ENV") == "dev" {
		secure = false
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		MaxAge:   300, // 5 minutes
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})

	url := googleOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (s *Dashboard) GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	// Verify state parameter
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Clear state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
	})

	// Exchange code for token
	code := r.URL.Query().Get("code")
	token, err := googleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Failed to exchange token: %v", err)
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	// Get user info from Google
	userInfo, err := s.getUserInfoFromGoogle(token.AccessToken)
	if err != nil {
		log.Printf("Failed to get user info: %v", err)
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	// Create or update user in database
	user, err := s.createOrUpdateUser(userInfo)
	if err != nil {
		log.Printf("Failed to create/update user: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Create session
	sessionToken, err := s.createSession(user.ID, token.AccessToken, token.RefreshToken)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Set session cookie
	secure := true
	if os.Getenv("ENV") == "dev" {
		secure = false
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	// Update visitor tracking cookie id to email hash
	s.updateVisitorIDToEmail(w, user.Email)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Dashboard) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Get session token from cookie
	cookie, err := r.Cookie("session_token")
	if err == nil {
		// Delete session from database
		s.db.Exec("DELETE FROM sessions WHERE session_token = ?", cookie.Value)
	}

	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/",
	})

	// Redirect to main page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Dashboard) getUserInfoFromGoogle(accessToken string) (*GoogleUserInfo, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

func (s *Dashboard) createOrUpdateUser(userInfo *GoogleUserInfo) (*User, error) {
	// Check if user exists
	var user User
	err := s.db.QueryRow(`
		SELECT id, email, username, google_id, avatar_url, created_at, updated_at, is_active
		FROM users WHERE google_id = ?
	`, userInfo.ID).Scan(
		&user.ID, &user.Email, &user.Username, &user.GoogleID,
		&user.AvatarURL, &user.CreatedAt, &user.UpdatedAt, &user.IsActive,
	)

	if err != nil {
		// User doesn't exist, create new one
		username := userInfo.Name
		if username == "" {
			username = userInfo.Email
		}

		result, err := s.db.Exec(`
			INSERT INTO users (email, username, google_id, avatar_url)
			VALUES (?, ?, ?, ?)
		`, userInfo.Email, username, userInfo.ID, userInfo.Picture)

		if err != nil {
			return nil, err
		}

		userID, _ := result.LastInsertId()
		user.ID = int(userID)
		user.Email = userInfo.Email
		user.Username = username
		user.GoogleID = userInfo.ID
		user.AvatarURL = userInfo.Picture
		user.IsActive = true
	} else {
		// User exists, update info
		_, err = s.db.Exec(`
			UPDATE users SET email = ?, avatar_url = ?, updated_at = CURRENT_TIMESTAMP
			WHERE google_id = ?
		`, userInfo.Email, userInfo.Picture, userInfo.ID)

		if err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func (s *Dashboard) createSession(userID int, accessToken, refreshToken string) (string, error) {
	// Generate session token
	b := make([]byte, 32)
	rand.Read(b)
	sessionToken := base64.URLEncoding.EncodeToString(b)

	// Insert session into database
	_, err := s.db.Exec(`
		INSERT INTO sessions (user_id, session_token, access_token, refresh_token, expires_at)
		VALUES (?, ?, ?, ?, ?)
	`, userID, sessionToken, accessToken, refreshToken, time.Now().Add(7*24*time.Hour))

	return sessionToken, err
}

func (s *Dashboard) GetCurrentUser(r *http.Request) (*User, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return nil, fmt.Errorf("no session cookie")
	}

	var user User
	err = s.db.QueryRow(`
		SELECT u.id, u.email, u.username, u.google_id, u.avatar_url, u.created_at, u.updated_at, u.is_active
		FROM users u
		JOIN sessions s ON u.id = s.user_id
		WHERE s.session_token = ? AND s.expires_at > CURRENT_TIMESTAMP
	`, cookie.Value).Scan(
		&user.ID, &user.Email, &user.Username, &user.GoogleID,
		&user.AvatarURL, &user.CreatedAt, &user.UpdatedAt, &user.IsActive,
	)

	if err != nil {
		return nil, fmt.Errorf("invalid session")
	}

	return &user, nil
}

func (s *Dashboard) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := s.GetCurrentUser(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Add user to request context
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AdminMiddleware restricts access to admin-only endpoints to ztkent@gmail.com
func (s *Dashboard) AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := s.GetCurrentUser(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Check if user is the admin
		if user.Email != "ztkent@gmail.com" {
			http.Error(w, "Forbidden - Admin access required", http.StatusForbidden)
			return
		}

		// Add user to request context
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Dashboard) UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.GetCurrentUser(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"authenticated": false,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"authenticated": true,
		"user": map[string]interface{}{
			"id":         user.ID,
			"email":      user.Email,
			"username":   user.Username,
			"avatar_url": user.AvatarURL,
		},
	})
}
