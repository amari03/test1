//File in internal/data folder
package data

import (
	"database/sql"
)

// Models struct holds all data models for the application.
type Models struct {
	Users    UserModel
	Officers OfficerModel
	Courses  CourseModel
	Sessions SessionModel
	Facilitators FacilitatorModel 
	Attendance   AttendanceModel
	Tokens TokenModel 
	SessionFacilitators SessionFacilitatorModel
	SessionFeedback     SessionFeedbackModel
	ImportJobs          ImportJobModel
}

// NewModels initializes and returns a Models struct.
func NewModels(db *sql.DB) Models {
	return Models{
		Users:    UserModel{DB: db},
		Officers: OfficerModel{DB: db},
		Courses:  CourseModel{DB: db},
		Sessions: SessionModel{DB: db},
		Facilitators: FacilitatorModel{DB: db},
		Attendance:   AttendanceModel{DB: db},
		Tokens:  TokenModel{DB: db},
		SessionFacilitators: SessionFacilitatorModel{DB: db},
		SessionFeedback:     SessionFeedbackModel{DB: db},
		ImportJobs:          ImportJobModel{DB: db},
	}
}