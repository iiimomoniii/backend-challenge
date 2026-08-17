package handler

import (
	"encoding/json"
	"net/http"

	appuser "github.com/iiimomoniii/backend-challenge/internal/application/user"
	domainuser "github.com/iiimomoniii/backend-challenge/internal/domain/user"
	"github.com/iiimomoniii/backend-challenge/internal/interface/http/request"
	"github.com/iiimomoniii/backend-challenge/internal/interface/http/response"
	"github.com/iiimomoniii/backend-challenge/pkg/code"
)

// UserHandler ทำหน้าที่จัดการ HTTP request/response
// สำหรับ user api
type UserHandler struct {
	create    *appuser.CreateUseCase
	search    *appuser.SearchUseCase
	searchAll *appuser.SearchAllUseCase
	update    *appuser.UpdateUseCase
	delete    *appuser.DeleteUseCase
}

// NewUserHandler สร้าง UserHandler
// โดยรับ use case ที่เกี่ยวข้องกับ user เข้ามาใช้งาน
func NewUserHandler(create *appuser.CreateUseCase, search *appuser.SearchUseCase,
	searchAll *appuser.SearchAllUseCase, update *appuser.UpdateUseCase, del *appuser.DeleteUseCase) *UserHandler {
	return &UserHandler{create: create, search: search, searchAll: searchAll, update: update, delete: del}
}

// Create รับ HTTP request สำหรับสร้าง user
// และเรียก CreateUseCase เพื่อสร้าง user
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req request.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, domainuser.CodeInvalidInput)
		return
	}
	// แปลง HTTP request เป็น use case input
	// แล้วส่งให้ CreateUseCase
	u, err := h.create.Execute(r.Context(), req.ToUseCase())
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	// แปลง domain user เป็น HTTP response
	// และreturn status 201 Created
	writeJSON(w, http.StatusCreated, response.FromUser(u))
}

// Search รับ HTTP request สำหรับค้นหา user ด้วย ID
func (h *UserHandler) Search(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	// เรียก SearchUseCase เพื่อค้นหา user
	u, err := h.search.Execute(r.Context(), id)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	// แปลง domain user เป็น HTTP response
	writeJSON(w, http.StatusOK, response.FromUser(u))
}

// SearchAll รับ HTTP request สำหรับค้นหา user ทั้งหมด
func (h *UserHandler) SearchAll(w http.ResponseWriter, r *http.Request) {
	// เรียก SearchAllUseCase เพื่อค้นหา user ทั้งหมด
	users, err := h.searchAll.Execute(r.Context())
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	// แปลง domain user หลายรายการเป็น HTTP response
	writeJSON(w, http.StatusOK, response.FromUsers(users))
}

// Update รับ HTTP request สำหรับแก้ไขข้อมูล user
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req request.UpdateRequest
	// decode JSON จาก request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, domainuser.CodeInvalidInput)
		return
	}

	// แปลง HTTP Request เป็น use case input
	// แล้วส่งให้ UpdateUseCase
	u, err := h.update.Execute(r.Context(), id, req.ToUseCase())
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	// แปลง domain user เป็น HTTP response
	writeJSON(w, http.StatusOK, response.FromUser(u))
}

// Delete รับ HTTP request สำหรับลบ user
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	// เรียก DeleteUseCase เพื่อลบ user
	if err := h.delete.Execute(r.Context(), id); err != nil {
		writeDomainError(w, r, err)
		return
	}
	// ลบสำเร็จและไม่มีข้อมูลที่ต้องreturn
	w.WriteHeader(http.StatusNoContent)
}

// writeJSON ใช้สำหรับเขียน HTTP JSON response
// พร้อมกำหนด HTTP status code และ content-type
func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// writeDomainError แปลง domain error
// เป็น HTTP error response
// ถ้าเป็น domain error ที่ map ไว้ก็จะส่งกลับตามที่ map
func writeDomainError(w http.ResponseWriter, r *http.Request, err error) {
	if de, ok := domainuser.AsDomainError(err); ok {
		writeError(w, r, de.Code)
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
}

// writeError ใช้สำหรับสร้าง HTTP error response
// จาก ErrorCode และ Language ที่ client ต้องการ
func writeError(w http.ResponseWriter, r *http.Request, errCode string) {
	lang := code.ParseLang(r.Header.Get("Accept-Language"))
	writeJSON(w, response.StatusForCode(errCode), response.FromErrorCode(errCode, lang))
}
