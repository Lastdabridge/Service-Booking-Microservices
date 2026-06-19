package kafka
 
import "time"
 
const (
	SuspiciousThreshold int64 = 5
	SuspiciousWindow = 5 * time.Minute
)