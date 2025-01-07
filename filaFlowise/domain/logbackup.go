package domain

import "time"

// FailedRequestLog estrutura para logar requisições com falha no MongoDB
type FailedRequestLog struct {
	UserNs       string                 `bson:"userNs"`
	Request      map[string]interface{} `bson:"request"`
	ResponseData map[string]interface{} `bson:"responseData"`
	ErrorMsg     string                 `bson:"errorMsg"`
	Timestamp    time.Time              `bson:"timestamp"`
}

// LogBackup interface para backup de logs de falha
type LogBackup interface {
	SaveFailedRequest(logData FailedRequestLog) error
}
