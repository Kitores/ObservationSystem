package getHosts

import (
	"fmt"
	"github.com/Kitores/ObservationSystem/internal/models/entity"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

type HostsGetter interface {
	GetHostErrorStats(hours int) ([]entity.HostErrorStats, error)
}

func New(hostsGetter HostsGetter) echo.HandlerFunc {
	return func(c echo.Context) error {
		hours := c.QueryParam("hours")
		intHours, err := strconv.Atoi(hours)
		hosts, err := hostsGetter.GetHostErrorStats(intHours)
		if err != nil {
			fmt.Println(err)
			return c.JSON(http.StatusInternalServerError, err)
		}
		return c.JSON(http.StatusOK, hosts)
	}
}
