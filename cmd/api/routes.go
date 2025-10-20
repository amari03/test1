package main

import (
    "net/http"
    "github.com/julienschmidt/httprouter"
)

/*You should apply the app.requireActivatedUser wrapper to any handlers in your 
application that require a logged-in, activated user. 
*/

func (app *application) routes() http.Handler {
    router := httprouter.New()

    router.NotFound = http.HandlerFunc(app.notFoundResponse)
    router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

    router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

    router.HandlerFunc(http.MethodPost, "/v1/users", app.createUserHandler)
    router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
    router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)
    router.HandlerFunc(http.MethodPost, "/v1/tokens/password-reset", app.createPasswordResetTokenHandler)
    router.HandlerFunc(http.MethodPut, "/v1/users/password", app.updatePasswordHandler)
    router.HandlerFunc(http.MethodGet, "/v1/users/:id", app.getUserHandler)
    router.HandlerFunc(http.MethodPatch, "/v1/users/:id", app.updateUserHandler)
    router.HandlerFunc(http.MethodDelete, "/v1/users/:id", app.deleteUserHandler)
    //this is how the requireActivated function will look like
    router.Handler(http.MethodGet, "/v1/users", app.requireActivatedUser(http.HandlerFunc(app.listUsersHandler)))

    router.Handler(http.MethodPost, "/v1/officers", app.requireActivatedUser(http.HandlerFunc(app.createOfficerHandler)))
    router.HandlerFunc(http.MethodGet, "/v1/officers/:id", app.getOfficerHandler)
    router.Handler(http.MethodPatch, "/v1/officers/:id", app.requireActivatedUser(http.HandlerFunc(app.updateOfficerHandler)))
    router.Handler(http.MethodDelete, "/v1/officers/:id", app.requireActivatedUser(http.HandlerFunc(app.deleteOfficerHandler)))
    router.HandlerFunc(http.MethodGet, "/v1/officers", app.listOfficersHandler)

    router.Handler(http.MethodPost, "/v1/courses", app.requireActivatedUser(http.HandlerFunc(app.createCourseHandler)))
    router.HandlerFunc(http.MethodGet, "/v1/courses/:id", app.getCourseHandler)
    router.Handler(http.MethodPatch, "/v1/courses/:id", app.requireActivatedUser(http.HandlerFunc(app.updateCourseHandler)))
    router.Handler(http.MethodDelete, "/v1/courses/:id", app.requireActivatedUser(http.HandlerFunc(app.deleteCourseHandler)))
    router.HandlerFunc(http.MethodGet, "/v1/courses", app.listCoursesHandler)

    router.Handler(http.MethodPost, "/v1/sessions", app.requireActivatedUser(http.HandlerFunc(app.createSessionHandler)))
    router.HandlerFunc(http.MethodGet, "/v1/sessions/:id", app.getSessionHandler)
    router.Handler(http.MethodPatch, "/v1/sessions/:id", app.requireActivatedUser(http.HandlerFunc(app.updateSessionHandler)))
    router.Handler(http.MethodDelete, "/v1/sessions/:id", app.requireActivatedUser(http.HandlerFunc(app.deleteSessionHandler)))
    router.HandlerFunc(http.MethodGet, "/v1/sessions", app.listSessionsHandler)

    router.Handler(http.MethodPost, "/v1/facilitators", app.requireActivatedUser(http.HandlerFunc(app.createFacilitatorHandler)))
    router.HandlerFunc(http.MethodGet, "/v1/facilitators/:id", app.getFacilitatorHandler)
    router.Handler(http.MethodPatch, "/v1/facilitators/:id", app.requireActivatedUser(http.HandlerFunc(app.updateFacilitatorHandler)))
    router.Handler(http.MethodDelete, "/v1/facilitators/:id", app.requireActivatedUser(http.HandlerFunc(app.deleteFacilitatorHandler)))
    router.HandlerFunc(http.MethodGet, "/v1/facilitators", app.listFacilitatorsHandler)

    router.Handler(http.MethodPost, "/v1/attendance", app.requireActivatedUser(http.HandlerFunc(app.createAttendanceHandler)))
    router.Handler(http.MethodDelete, "/v1/attendance/:id", app.requireActivatedUser(http.HandlerFunc(app.deleteAttendanceHandler)))
    router.HandlerFunc(http.MethodGet, "/v1/attendance/:id", app.getAttendanceHandler)
    router.Handler(http.MethodPatch, "/v1/attendance/:id", app.requireActivatedUser(http.HandlerFunc(app.updateAttendanceHandler)))
    router.HandlerFunc(http.MethodGet, "/v1/attendance", app.listAttendanceHandler)


    // Session Facilitators
    router.Handler(http.MethodPost, "/v1/session-facilitators", app.requireActivatedUser(http.HandlerFunc(app.createSessionFacilitatorHandler)))
    router.Handler(http.MethodDelete, "/v1/session-facilitators/:id", app.requireActivatedUser(http.HandlerFunc(app.deleteSessionFacilitatorHandler)))
    router.HandlerFunc(http.MethodGet, "/v1/session-facilitators", app.listSessionFacilitatorsHandler)

    // Session Feedback
    router.Handler(http.MethodPost, "/v1/session-feedback", app.requireActivatedUser(http.HandlerFunc(app.createSessionFeedbackHandler)))
    router.HandlerFunc(http.MethodGet, "/v1/session-feedback", app.listSessionFeedbackHandler)

    // Import Jobs
    router.Handler(http.MethodPost, "/v1/import-jobs", app.requireActivatedUser(http.HandlerFunc(app.createImportJobHandler)))
    router.HandlerFunc(http.MethodGet, "/v1/import-jobs", app.listImportJobsHandler)
    router.HandlerFunc(http.MethodGet, "/v1/import-jobs/:id", app.getImportJobHandler)

    
    return app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router))))
}