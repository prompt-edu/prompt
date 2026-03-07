package keycloakTokenVerifier

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// this is a reduced middleware, which does not require an prompt service account
func ApplicationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := extractBearerToken(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		ctx := c.Request.Context()
		idToken, err := verifier.Verify(ctx, tokenString)
		if err != nil {
			log.Error("Failed to validate token: ", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		claims, err := extractClaims(idToken)
		if err != nil {
			log.Error("Failed to parse claims: ", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		if !checkAuthorizedParty(claims, KeycloakTokenVerifierSingleton.expectedAuthorizedParty) {
			log.Error("Token authorized party mismatch")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token authorized party mismatch"})
			return
		}

		// extract user Id
		userID, ok := claims["sub"].(string)
		if !ok {
			log.Error("Failed to extract user ID (sub) from token claims")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
			return
		}

		userEmail, ok := claims["email"].(string)
		if !ok {
			log.Error("Failed to extract user email (sub) from token claims")
		}

		matriculationNumber, ok := claims["matriculation_number"].(string)
		if !ok {
			log.Error("Failed to extract user matriculation number (sub) from token claims")
		}

		universityLogin, ok := claims["university_login"].(string)
		if !ok {
			log.Error("Failed to extract user university login (sub) from token claims")
		}

		firstName, ok := claims["given_name"].(string)
		if !ok {
			log.Error("Failed to extract user given name (sub) from token claims")
		}

		lastName, ok := claims["family_name"].(string)
		if !ok {
			log.Error("Failed to extract user family name (sub) from token claims")
		}

		// Store the extracted roles in the context
		c.Set(CtxUserID, userID)
		c.Set(CtxUserEmail, userEmail)
		c.Set(CtxMatriculationNumber, matriculationNumber)
		c.Set(CtxUniversityLogin, universityLogin)
		c.Set(CtxFirstName, firstName)
		c.Set(CtxLastName, lastName)
		c.Next()
	}
}
