package main

import (
	"encoding/json"
	"net/http"
)

func telemetryHandler(pub Publisher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "metodo nao permitido", http.StatusMethodNotAllowed)
			return
		}

		var telemetry Telemetry
		if err := json.NewDecoder(r.Body).Decode(&telemetry); err != nil {
			http.Error(w, "json invalido", http.StatusBadRequest)
			return
		}

		body, err := json.Marshal(telemetry)
		if err != nil {
			http.Error(w, "erro ao serializar payload", http.StatusInternalServerError)
			return
		}

		err = pub.Publish(body)
		if err != nil {
			http.Error(w, "erro ao publicar mensagem", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("mensagem enfileirada com sucesso"))
	}
}