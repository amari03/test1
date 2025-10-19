package main

import (
    "net/http"
    "github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
    router := httprouter.New()

    router.NotFound = http.HandlerFunc(app.notFoundResponse)
    router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

    router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
    router.HandlerFunc(http.MethodPost, "/v1/users", app.createUserHandler)
    router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
     router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)
    router.HandlerFunc(http.MethodGet, "/v1/users/:id", app.getUserHandler)
    router.HandlerFunc(http.MethodPatch, "/v1/users/:id", app.updateUserHandler)
    router.HandlerFunc(http.MethodDelete, "/v1/users/:id", app.deleteUserHandler)
    router.HandlerFunc(http.MethodGet, "/v1/users", app.listUsersHandler)
    router.HandlerFunc(http.MethodPost, "/v1/officers", app.createOfficerHandler)
    router.HandlerFunc(http.MethodGet, "/v1/officers/:id", app.getOfficerHandler)
    router.HandlerFunc(http.MethodPatch, "/v1/officers/:id", app.updateOfficerHandler)
    router.HandlerFunc(http.MethodDelete, "/v1/officers/:id", app.deleteOfficerHandler)
    router.HandlerFunc(http.MethodGet, "/v1/officers", app.listOfficersHandler)
    router.HandlerFunc(http.MethodPost, "/v1/courses", app.createCourseHandler)
    router.HandlerFunc(http.MethodGet, "/v1/courses/:id", app.getCourseHandler)
    router.HandlerFunc(http.MethodPatch, "/v1/courses/:id", app.updateCourseHandler)
    router.HandlerFunc(http.MethodDelete, "/v1/courses/:id", app.deleteCourseHandler)
    router.HandlerFunc(http.MethodGet, "/v1/courses", app.listCoursesHandler)
    router.HandlerFunc(http.MethodPost, "/v1/sessions", app.createSessionHandler)
    router.HandlerFunc(http.MethodGet, "/v1/sessions/:id", app.getSessionHandler)
    router.HandlerFunc(http.MethodPatch, "/v1/sessions/:id", app.updateSessionHandler)
    router.HandlerFunc(http.MethodDelete, "/v1/sessions/:id", app.deleteSessionHandler)
    router.HandlerFunc(http.MethodGet, "/v1/sessions", app.listSessionsHandler)
    router.HandlerFunc(http.MethodPost, "/v1/facilitators", app.createFacilitatorHandler)
    router.HandlerFunc(http.MethodGet, "/v1/facilitators/:id", app.getFacilitatorHandler)
    router.HandlerFunc(http.MethodPatch, "/v1/facilitators/:id", app.updateFacilitatorHandler)
    router.HandlerFunc(http.MethodDelete, "/v1/facilitators/:id", app.deleteFacilitatorHandler)
    router.HandlerFunc(http.MethodGet, "/v1/facilitators", app.listFacilitatorsHandler)
    router.HandlerFunc(http.MethodPost, "/v1/attendance", app.createAttendanceHandler)
    router.HandlerFunc(http.MethodDelete, "/v1/attendance/:id", app.deleteAttendanceHandler)
    router.HandlerFunc(http.MethodGet, "/v1/attendance/:id", app.getAttendanceHandler)
    router.HandlerFunc(http.MethodPatch, "/v1/attendance/:id", app.updateAttendanceHandler)
    router.HandlerFunc(http.MethodGet, "/v1/attendance", app.listAttendanceHandler)
    
    return app.recoverPanic(app.rateLimit(app.authenticate(router)))
}