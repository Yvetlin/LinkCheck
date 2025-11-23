package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/linkcheck/internal/pdf"
	"github.com/linkcheck/internal/queue"
	"github.com/linkcheck/internal/service"
	"github.com/linkcheck/internal/storage"
	"github.com/linkcheck/pkg/types"
)

// Handler обрабатывает HTTP запросы
type Handler struct {
	storage *storage.Storage
	queue   *queue.Queue
	checker *service.Checker
	pdfGen  *pdf.Generator
}

// NewHandler создает новый обработчик HTTP запросов
func NewHandler(storage *storage.Storage, queue *queue.Queue, checker *service.Checker) *Handler {
	return &Handler{
		storage: storage,
		queue:   queue,
		checker: checker,
		pdfGen:  pdf.NewGenerator(),
	}
}

// SubmitLinks обрабатывает запрос на проверку ссылок
func (h *Handler) SubmitLinks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	var req types.SubmitLinksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Неверный запрос: %v", err), http.StatusBadRequest)
		return
	}

	if len(req.Links) == 0 {
		http.Error(w, "Список ссылок пуст", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	linksStatus := h.checker.CheckLinks(ctx, req.Links)

	linksNum := h.storage.AddLinksSet(linksStatus)

	if err := h.storage.Save(); err != nil {
		log.Printf("Ошибка при сохранении состояния: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	response := types.SubmitLinksResponse{
		Links:    make(map[string]string),
		LinksNum: linksNum,
	}

	for link, status := range linksStatus {
		response.Links[link] = string(status)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetReport обрабатывает запрос на получение PDF отчета
func (h *Handler) GetReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	var req types.GetReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Неверный запрос: %v", err), http.StatusBadRequest)
		return
	}

	if len(req.LinksList) == 0 {
		http.Error(w, "Список номеров пуст", http.StatusBadRequest)
		return
	}

	linksSets := h.storage.GetLinksSets(req.LinksList)

	if len(linksSets) == 0 {
		http.Error(w, "Наборы ссылок с указанными номерами не найдены", http.StatusNotFound)
		return
	}

	pdfData, err := h.pdfGen.GenerateReport(linksSets)
	if err != nil {
		log.Printf("Ошибка при генерации PDF: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	filename := formatLinksList(req.LinksList)

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=report_%s.pdf", filename))
	w.Header().Set("Content-Length", strconv.Itoa(len(pdfData)))

	w.Write(pdfData)
}

// formatLinksList создает строку из номеров для имени файла
func formatLinksList(nums []int) string {
	result := ""
	for i, num := range nums {
		if i > 0 {
			result += "_"
		}
		result += strconv.Itoa(num)
	}
	return result
}

// RegisterRoutes регистрирует маршруты для обработчика
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/submit", h.SubmitLinks)
	mux.HandleFunc("/report", h.GetReport)
}
