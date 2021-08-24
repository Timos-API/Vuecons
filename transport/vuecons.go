package transport

import (
	"Timos-API/Vuecons/service"
	"encoding/json"
	"net/http"

	"github.com/Timos-API/authenticator"
	"github.com/gorilla/mux"
)

type VueconsTransport struct {
	s *service.VueconsService
}

func NewVueconsTransport(s *service.VueconsService) *VueconsTransport {
	return &VueconsTransport{s}
}

func (t *VueconsTransport) RegisterVueconsRoutes(router *mux.Router) {
	router.HandleFunc("/vuecons", t.getAllVuecons).Methods("GET")
	router.HandleFunc("/vuecons", authenticator.Middleware(t.uploadVuecon, authenticator.Guard().G("admin").P("vuecons.create"))).Methods("POST")
	router.HandleFunc("/vuecons/{id}", authenticator.Middleware(t.deleteVuecon, authenticator.Guard().G("admin").P("vuecons.delete"))).Methods("DELETE")
}

func (t *VueconsTransport) getAllVuecons(w http.ResponseWriter, req *http.Request) {
	vuecons, err := t.s.GetAllVuecons(req.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	err = json.NewEncoder(w).Encode(vuecons)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (t *VueconsTransport) uploadVuecon(w http.ResponseWriter, req *http.Request) {
	maxSize := int64(1024000)

	err := req.ParseMultipartForm(maxSize)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		return
	}

	file, fileHeader, err := req.FormFile("vuecon")
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		return
	}

	defer file.Close()

	vuecon, err := t.s.Upload(req.Context(), file, fileHeader)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(vuecon)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (t *VueconsTransport) deleteVuecon(w http.ResponseWriter, req *http.Request) {
	id, ok := mux.Vars(req)["id"]

	if !ok {
		http.Error(w, "Missing param: id", http.StatusBadRequest)
		return
	}

	success, err := t.s.Delete(req.Context(), id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !success {
		http.Error(w, "Couldn't delete message", http.StatusInternalServerError)
	}

}
