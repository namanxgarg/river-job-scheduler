package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/yourusername/golang-job-scheduler/internal/jobs"
)

func AttachSchedulerAPI(r *mux.Router, js *jobs.JobService) {
	r.HandleFunc("/jobs", createJobHandler(js)).Methods("POST")
	r.HandleFunc("/jobs/{id}", getJobHandler(js)).Methods("GET")
}

func createJobHandler(js *jobs.JobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Payload    string `json:"payload"`
			MaxRetries int    `json:"maxRetries"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		jobID, err := js.ScheduleJob(req.Payload)
		if err != nil {
			log.Error().Err(err).Msg("Failed to schedule job")
			http.Error(w, "Failed to schedule job", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"jobID": jobID})
	}
}

func getJobHandler(js *jobs.JobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		jobID := vars["id"]
		job, err := js.GetJob(jobID)
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(job)
	}
}
