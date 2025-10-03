module auth-service

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/golang-jwt/jwt/v5 v5.0.0
    github.com/lib/pq v1.10.9
    golang.org/x/crypto v0.14.0
    shared-security v0.0.0
)

replace shared-security => ./shared-security
