package getServices

import (
	"encoding/json"
	"fmt"
	"github.com/Kitores/ObservationSystem/internal/models/entity"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

type ServivicesGetter interface {
	GetServicesByHost(filterType string, filterValue interface{}) ([]entity.Service, error)
}

type request struct {
	FilterType  string          `json:"filter_type"`
	FilterValue json.RawMessage `json:"filter_value"`
}

func New(servivicesGetter ServivicesGetter) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := new(request)
		if err := c.Bind(req); err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}

		// Парсим FilterValue в зависимости от FilterType
		var parsedValue interface{}

		switch req.FilterType {
		case "ip", "host_name":
			var strVal string
			if err := json.Unmarshal(req.FilterValue, &strVal); err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"error": "Invalid filter value for type " + req.FilterType,
				})
			}
			parsedValue = strVal

		case "host_id":
			var numVal int64
			// Пробуем распарсить как число
			if err := json.Unmarshal(req.FilterValue, &numVal); err != nil {
				// Если не число, пробуем как строку (на случай если пришло как строка)
				var strVal string
				if err := json.Unmarshal(req.FilterValue, &strVal); err != nil {
					return c.JSON(http.StatusBadRequest, map[string]string{
						"error": "Invalid host_id value",
					})
				}
				// Конвертируем строку в число
				numVal, err = strconv.ParseInt(strVal, 10, 64)
				if err != nil {
					return c.JSON(http.StatusBadRequest, map[string]string{
						"error": "Host ID must be a number",
					})
				}
			}
			parsedValue = numVal

		default:
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid filter type. Use 'ip', 'host_id', or 'host_name'",
			})
		}

		fmt.Println(req.FilterType, parsedValue)

		services, err := servivicesGetter.GetServicesByHost(req.FilterType, parsedValue)
		if err != nil {
			fmt.Println(err)
			return c.JSON(http.StatusInternalServerError, err)
		}
		return c.JSON(http.StatusOK, services)
	}
}
