package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gachal/mossbase/backend/internal/domain/repository"
	"github.com/gachal/mossbase/backend/pkg/response"
)

func SpaceAuth(spaceMemberRepo repository.SpaceMemberRepository, requireAdmin bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := response.GetUserID(c)
		spaceIDStr := c.Param("spaceId")
		spaceID, err := strconv.ParseUint(spaceIDStr, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid space id")
			c.Abort()
			return
		}

		member, err := spaceMemberRepo.FindBySpaceAndUser(c.Request.Context(), spaceID, userID)
		if err != nil {
			response.Error(c, http.StatusForbidden, "not a space member")
			c.Abort()
			return
		}

		if requireAdmin && !member.IsAdmin() {
			response.Error(c, http.StatusForbidden, "space admin required")
			c.Abort()
			return
		}

		c.Set("spaceID", spaceID)
		c.Set("memberRole", string(member.Role))
		c.Next()
	}
}
