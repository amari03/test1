package main

import (
    "fmt"
    "net/http"

    "github.com/amari03/test1/internal/data"
    "github.com/amari03/test1/internal/validator"
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