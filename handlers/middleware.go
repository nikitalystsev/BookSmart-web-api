package handlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

const (
	authorizationHeader = "Authorization"
	ID                  = "ID"
	Role                = "role"
)

func (h *Handler) readerIdentity(c *gin.Context) {
	id, role, err := h.parseAuthHeader(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, err.Error())
	}

	c.Set(ID, id)
	c.Set(Role, role)
}

func (h *Handler) parseAuthHeader(c *gin.Context) (string, string, error) {
	header := c.GetHeader(authorizationHeader)
	if header == "" {
		return "", "", errors.New("empty auth header")
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return "", "", errors.New("invalid auth header")
	}

	if len(headerParts[1]) == 0 {
		return "", "", errors.New("token is empty")
	}

	return h.tokenManager.Parse(headerParts[1])
}

func getReaderData(c *gin.Context) (string, string, error) {
	id, ok := c.Get(ID)
	if !ok {
		return "", "", errors.New("user id not found")
	}

	idStr, ok := id.(string)
	if !ok {
		return "", "", errors.New("user id is of invalid type")
	}

	role, ok := c.Get(Role)
	if !ok {
		return "", "", errors.New("user role not found")
	}

	roleStr, ok := role.(string)
	if !ok {
		return "", "", errors.New("user role is of invalid type")
	}

	return idStr, roleStr, nil
}
