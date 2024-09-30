package handlers

import (
	"context"
	httpModels "gfsloader/cmd/restserver/models"
	appModels "gfsloader/internal/models"
	"gfsloader/internal/storage/postgres/models"
	"net/http"

	"time"

	"github.com/gin-gonic/gin"
)

type ForecastProvider interface {
	GetForecastBySegments(ctx context.Context, segments []appModels.WKTRequestItem) ([]models.PGResponse, error)
}

type WKTHandler struct {
	forecastProvider ForecastProvider
}

func NewWKTHandler(
	forecastProvider ForecastProvider,
) *WKTHandler {
	return &WKTHandler{
		forecastProvider: forecastProvider,
	}
}

var (
	duration_3h, _ = time.ParseDuration("3h")
)

func (h *WKTHandler) HandlerByWKT(c *gin.Context) {
	var body httpModels.WKTRequestBody
	err := c.BindJSON(&body)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, "Bad request")
		return
	}

	q := make([]appModels.WKTRequestItem, 0, len(body.Shapes))

	for _, item := range body.Shapes {

		to := item.To
		if to != nil {
			_to := *item.To
			_to = _to.Round(duration_3h)
			to = &_to
		}

		q = append(q, appModels.WKTRequestItem{
			WKT:  item.WKT,
			From: item.From.Round(duration_3h),
			To:   to,
		})
	}

	res, err := h.forecastProvider.GetForecastBySegments(c.Request.Context(), q)

	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "Some error")
		return
	}

	response := make([]httpModels.ForecastResponse, 0, len(res))
	shapeGroup := make(map[string][]models.PGResponse, len(res))
	for _, item := range res {
		if l, ok := shapeGroup[item.Sec]; ok {
			shapeGroup[item.Sec] = append(l, item)
		} else {
			shapeGroup[item.Sec] = []models.PGResponse{
				item,
			}
		}
	}

	for shape, items := range shapeGroup {

		fcst := httpModels.ForecastResponse{
			Shape:    shape,
			Forecast: make([]httpModels.ForecastDetail, 0, len(items)),
		}

		for _, item := range items {
			fcst.Forecast = append(fcst.Forecast,
				httpModels.ForecastDetail{
					DateTime:    item.Date,
					Temperature: &item.Temperature,
					Pressure:    &item.Pressure,
					RHumidity:   &item.RHumidity,
					CRain:       &item.CRain,
					Wind: &httpModels.WindInfo{
						U: item.UWind,
						V: item.VWind,
					},
				})
		}

		response = append(response, fcst)
	}

	c.IndentedJSON(http.StatusOK, response)
}
