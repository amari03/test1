package main

import (
    "fmt"
    "net/http"
	"errors"

    "github.com/amari03/test1/internal/data"
    "github.com/amari03/test1/internal/validator"
	"github.com/julienschmidt/httprouter"
)

func (app *application) createAttendanceHandler(w http.ResponseWriter, r *http.Request) {
    var input struct {
        OfficerID     string  `json:"officer_id"`
        SessionID     string  `json:"session_id"`
        Status        string  `json:"status"`
        CreditedHours float64 `json:"credited_hours"`
    }

    err := app.readJSON(w, r, &input)
    if err != nil {
        app.badRequestResponse(w, r, err)
        return
    }

    attendance := &data.Attendance{
        OfficerID:     input.OfficerID,
        SessionID:     input.SessionID,
        Status:        input.Status,
        CreditedHours: input.CreditedHours,
    }

    v := validator.New()
    if data.ValidateAttendance(v, attendance); !v.Valid() {
        app.failedValidationResponse(w, r, v.Errors)
        return
    }

    err = app.models.Attendance.Insert(attendance)
    if err != nil {
        app.serverErrorResponse(w, r, err)
        return
    }

    headers := make(http.Header)
    headers.Set("Location", fmt.Sprintf("/v1/attendance/%s", attendance.ID))

    err = app.writeJSON(w, http.StatusCreated, envelope{"attendance": attendance}, headers)
    if err != nil {
        app.serverErrorResponse(w, r, err)
    }
}

func (app *application) deleteAttendanceHandler(w http.ResponseWriter, r *http.Request) {
    params := httprouter.ParamsFromContext(r.Context())
    id := params.ByName("id")

    err := app.models.Attendance.Delete(id)
    if err != nil {
        switch {
        case errors.Is(err, data.ErrRecordNotFound):
            app.notFoundResponse(w, r)
        default:
            app.serverErrorResponse(w, r, err)
        }
        return
    }

    err = app.writeJSON(w, http.StatusOK, envelope{"message": "attendance record successfully deleted"}, nil)
    if err != nil {
        app.serverErrorResponse(w, r, err)
    }
}