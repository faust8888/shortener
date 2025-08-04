package handler

import (
	"encoding/json"
	"github.com/faust8888/shortener/internal/app/model"
	"github.com/faust8888/shortener/internal/app/security"
	"net/http"
)

// Stat — обработчик HTTP-запросов для получения статистики.
// Хранит ссылку на сервисный слой (Collector) для получения данных и
// строку доверенной подсети trustedSubnet для проверки административного доступа.
type Stat struct {
	service       Collector
	trustedSubnet string
}

// Collector — интерфейс сервисного слоя для получения статистики по сокращённым ссылкам.
//
// Определяет метод Collect, который возвращает структуру статистики или ошибку.
type Collector interface {
	Collect() (*model.Statistic, error)
}

// CollectStatistic обрабатывает HTTP-запрос на получение статистики.
//
// Метод выполняет проверку IP клиента на принадлежность к доверенной подсети через security.IsAllowedTrustedIP.
// Если проверка не пройдена, возвращает HTTP 403 Forbidden и прерывает обработку.
//
// Если IP доверенный, метод вызывает сервисный слой для получения статистики,
// сериализует её в JSON и отправляет клиенту с HTTP статусом 200 OK.
//
// В случае ошибок при сборе статистики или сериализации возвращается HTTP 500 Internal Server Error.
//
// Параметры:
//   - res: http.ResponseWriter для формирования ответа клиенту.
//   - req: *http.Request с информацией о входящем запросе.
func (handler *Stat) CollectStatistic(res http.ResponseWriter, req *http.Request) {
	if !security.IsAllowedTrustedIP(req, res, handler.trustedSubnet) {
		return
	}
	stat, err := handler.service.Collect()
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	resp, err := json.Marshal(&stat)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	_, err = res.Write(resp)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}
