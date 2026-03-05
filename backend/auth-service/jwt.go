package main

import (
    "time"
    
    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
    "github.com/sin-engine/auth-service/models"
)

var jwtSecret = []byte("your-secret-key-change-this") // Change in production

type Claims struct {
    UserID   string `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}

func GenerateTokens(user *models.User) (accessToken, refreshToken string, err error) {
    // Access token (15 minutes)
    accessClaims := &Claims{
        UserID:   user.ID,
        Username: user.Username,
        Role:     user.Role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            ID:        uuid.New().String(),
        },
    }
    
    accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
    accessToken, err = accessTokenObj.SignedString(jwtSecret)
    if err != nil {
        return "", "", err
    }
    
    // Refresh token (7 days)
    refreshClaims := &Claims{
        UserID: user.ID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            ID:        uuid.New().String(),
        },
    }
    
    refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
    refreshToken, err = refreshTokenObj.SignedString(jwtSecret)
    if err != nil {
        return "", "", err
    }
    
    return accessToken, refreshToken, nil
}

func ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return jwtSecret, nil
    })
    
    if err != nil {
        return nil, err
    }
    
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    
    return nil, jwt.ErrSignatureInvalid
}