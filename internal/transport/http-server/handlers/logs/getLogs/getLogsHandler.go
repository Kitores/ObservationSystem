package getLogs

import (
	"encoding/json"
	"fmt"
	"github.com/Kitores/ObservationSystem/internal/models/entity"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

type LogsGetter interface {
	GetRecentLogsByInterval(filterType string, filterValue interface{}, value int, unit string) ([]entity.LogEntity, error)
}

type request struct {
	FilterType  string          `json:"filter_type"`
	FilterValue json.RawMessage `json:"filter_value"`
	Value       int             `json:"value"`
	Unit        string          `json:"unit"`
}

func New(logsGetter LogsGetter) echo.HandlerFunc {
	return func(c echo.Context) error {

		// В дальнейшем можно будет запрашивать со стороны клиента не ID а IP хоста и строить запрос в бд посложнее
		req := new(request)
		if err := c.Bind(req); err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}

		// Парсим FilterValue в зависимости от FilterType
		var parsedValue interface{}
		var err error

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

		logs, err := logsGetter.GetRecentLogsByInterval(req.FilterType, parsedValue, req.Value, req.Unit)
		fmt.Println(logs)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}
		return c.JSON(http.StatusOK, logs)
	}
}
