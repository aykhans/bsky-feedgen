package middleware

import (
	"context"
	"crypto"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aykhans/bsky-feedgen/pkg/logger"
	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/golang-jwt/jwt/v5"
)

const UserDIDKey ContextKey = "user_did"

func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// No auth header, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		did, err := ValidateAuth(r.Context(), r)
		if err != nil {
			logger.Log.Error(err.Error())
		}
		ctx := context.WithValue(r.Context(), UserDIDKey, did)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

///////////////////////////////////////////////////

const (
	authorizationHeaderName        = "Authorization"
	authorizationHeaderValuePrefix = "Bearer "
)

// AuthorizationError is a custom error type for authorization failures.
type AuthorizationError struct {
	Message string
	Err     error // Wrapped error
}

// Error returns the formatted error message.
func (e *AuthorizationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the wrapped error, allowing for errors.Is and errors.As.
func (e *AuthorizationError) Unwrap() error {
	return e.Err
}

// Global (or dependency-injected) DID resolver with caching.
var didResolver *identity.CacheDirectory

// init function is called once when the package is initialized.
func init() {
	// Initialize the base directory for actual DID resolution.
	baseDir := identity.BaseDirectory{} // Zero value is usable.

	// Configure cache with appropriate TTLs.
	// Capacity 0 means unlimited cache size.
	// hitTTL: 24 hours for successful resolutions.
	// errTTL: 5 minutes for failed resolutions.
	// invalidHandleTTL: also 5 minutes for invalid handles.
	resolver := identity.NewCacheDirectory(
		&baseDir,
		0,             // Unlimited capacity
		24*time.Hour,  // hitTTL
		5*time.Minute, // errTTL
		5*time.Minute, // invalidHandleTTL
	)
	didResolver = &resolver
}

// getDIDSigningKey resolves a DID and extracts its public signing key.
// It leverages indigo's identity package which handles multibase decoding and key parsing.
func getDIDSigningKey(ctx context.Context, did string) (crypto.PublicKey, error) {
	atID, err := syntax.ParseAtIdentifier(did)
	if err != nil {
		return nil, fmt.Errorf("invalid DID syntax: %w", err)
	}

	// Use Lookup for bi-directional verification (handle -> DID -> handle).
	// The `Lookup` method returns an `Identity` struct which contains `PublicKey()` method
	// to get the signing key.
	identity, err := didResolver.Lookup(ctx, *atID)
	if err != nil {
		return nil, fmt.Errorf("DID resolution failed for %s: %w", did, err)
	}
	if identity == nil || identity.DID.String() == "" {
		return nil, fmt.Errorf("DID resolution returned empty identity for %s", did)
	}

	// Get the public key using the PublicKey() method from the Identity struct.
	publicKey, err := identity.PublicKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get signing key for DID %s: %w", did, err)
	}

	return publicKey, nil
}

// ValidateAuth validates the authorization header and returns the requester's DID.
// It requires a context.Context for DID resolution to allow for timeouts and cancellation.
func ValidateAuth(ctx context.Context, r *http.Request) (string, error) {
	authHeader := r.Header.Get(authorizationHeaderName)
	if authHeader == "" {
		return "", &AuthorizationError{Message: "Authorization header is missing"}
	}

	if !strings.HasPrefix(authHeader, authorizationHeaderValuePrefix) {
		return "", &AuthorizationError{Message: "Invalid authorization header format"}
	}

	jwtString := strings.TrimPrefix(authHeader, authorizationHeaderValuePrefix)
	jwtString = strings.TrimSpace(jwtString)

	// Define a custom claims struct if needed, otherwise use jwt.RegisteredClaims.
	claims := jwt.RegisteredClaims{}

	// Keyfunc callback to dynamically fetch the public key based on the JWT's issuer (iss).
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		// Assert claims to RegisteredClaims to get the issuer (iss).
		regClaims, ok := token.Claims.(*jwt.RegisteredClaims)
		if !ok {
			return nil, fmt.Errorf("invalid JWT claims type")
		}

		issuerDID := regClaims.Issuer
		if issuerDID == "" {
			return nil, fmt.Errorf("JWT 'iss' claim is missing")
		}

		// Resolve the DID and get the public signing key.
		publicKey, err := getDIDSigningKey(ctx, issuerDID)
		if err != nil {
			return nil, fmt.Errorf("failed to get signing key for DID %s: %w", issuerDID, err)
		}

		return publicKey, nil
	}

	token, err := jwt.ParseWithClaims(jwtString, &claims, keyFunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return "", &AuthorizationError{Message: "Invalid signature", Err: err}
		}
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", &AuthorizationError{Message: "Token expired", Err: err}
		}
		if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return "", &AuthorizationError{Message: "Token not valid yet", Err: err}
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return "", &AuthorizationError{Message: "Malformed token", Err: err}
		}
		// Catch other generic parsing or validation errors.
		return "", &AuthorizationError{Message: "Failed to parse or validate JWT", Err: err}
	}

	if !token.Valid {
		return "", &AuthorizationError{Message: "Token is invalid"}
	}

	// Return the issuer's DID.
	return claims.Issuer, nil
}
