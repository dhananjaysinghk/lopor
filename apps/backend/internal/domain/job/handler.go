package job

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lopor-ai/lopor/pkg/jobqueue"
	"github.com/lopor-ai/lopor/pkg/response"
)

type Handler struct {
	queue *jobqueue.Queue
}

func NewHandler(queue *jobqueue.Queue) *Handler {
	return &Handler{queue: queue}
}

func (h *Handler) EnqueueJob(c *fiber.Ctx) error {
	type EnqueueRequest struct {
		Type    string      `json:"type"` // "document_pdf_export", "rag_reindex", "cron_summary"
		Payload interface{} `json:"payload"`
	}

	var req EnqueueRequest
	if err := c.BodyParser(&req); err != nil || req.Type == "" {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Job type is required", nil)
	}

	job, err := h.queue.EnqueueJob(c.Context(), req.Type, req.Payload)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "ENQUEUE_FAILED", err.Error(), nil)
	}

	return response.Success(c, fiber.StatusAccepted, "Async job enqueued onto Redis worker pool", job)
}

func (h *Handler) GetJobStatus(c *fiber.Ctx) error {
	jobID := c.Params("jobId")
	if jobID == "" {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_JOB_ID", "Job ID is required", nil)
	}

	job, err := h.queue.GetJobStatus(c.Context(), jobID)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "JOB_NOT_FOUND", "Job status not found", nil)
	}

	return response.Success(c, fiber.StatusOK, "Job status retrieved", job)
}
