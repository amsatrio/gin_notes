package response

import "time"

type Response struct {
	Path      string      `json:"path" example:"/v1/m_biodata"`
	Timestamp JSONTime    `json:"timestamp" example:"2024-02-16 10:33:10" swaggertype:"string"`
	Status    int         `json:"status" example:"200"`
	Message   string      `json:"message" example:"success"`
	Data      interface{} `json:"data" swaggertype:"string"`
}

type ResponseTimestamp struct {
	Path      string      `json:"path" example:"/v1/m_biodata"`
	Timestamp time.Time   `json:"timestamp" example:"2024-02-16 10:33:10" swaggertype:"string" time_format:"2024-02-16 10:33:10"`
	Status    int         `json:"status" example:"200"`
	Message   string      `json:"message" example:"success"`
	Data      interface{} `json:"data" swaggertype:"string"`
}
