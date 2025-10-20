Phase 1: Setup - Becoming an Authorized User
Step 1: Register a New User
curl -i -X POST -H "Content-Type: application/json" \
-d '{"email": "testing@example.com", "password": "password123", "role": "admin"}' \ http://localhost:4000/v1/users

Step 2: Activate the User
-Go to your Mailtrap inbox.
-Open the welcome email sent to test@example.com.
-Copy the 26-character activation token from the email body.
# Replace YOUR_ACTIVATION_TOKEN with the token you copied from the email
curl -i -X PUT -H "Content-Type: application/json" \ -d '{"token": "YOUR_ACTIVATION_TOKEN"}' \ http://localhost:4000/v1/users/activated

Step 3: Authenticate and Get a Bearer Token
curl -i -X POST -H "Content-Type: application/json" \ -d '{"email": "testing@example.com", "password": "password123"}' \ http://localhost:4000/v1/tokens/authentication

Step 4: Store the Token in a Shell Variable
# Replace YOUR_BEARER_TOKEN with the token from the step above
export TOKEN="YOUR_BEARER_TOKEN"

# You can verify it's stored by running:
echo $TOKEN


**Officer Endpoint:**
Phase 2: Testing Officer Endpoints
1. Create Officer (POST)
curl -i -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d '{"regulation_number": "PC700", "first_name": "Andy", "last_name": "Doe", "sex": "male", "rank_code": "PC"}' http://localhost:4000/v1/officers

export OFFICER_ID="<the-id-you-just-copied>"

2. Get All Officers (GET)
curl -i -H "Authorization: Bearer $TOKEN" http://localhost:4000/v1/officers

3. Get One Officer (GET by ID)
curl -i -H "Authorization: Bearer $TOKEN" http://localhost:4000/v1/officers/$OFFICER_ID

4. Update Officer (PATCH)
curl -i -X PATCH -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" −d′"rankcode":"CPL"′ http://localhost:4000/v1/officers/TOKEN" −d′"rankc​ode":"CPL"′ http://localhost:4000/v1/officers/OFFICER_ID

5. Delete Officer (DELETE)
curl -i -X DELETE -H "Authorization: Bearer $TOKEN" http://localhost:4000/v1/officers/$OFFICER_ID

-------------------------------------------------------------------------------------

COURSES ENDPOINT 
-------------------
Phase 3: Testing Course Endpoints
1. Create Course (POST)
curl -i -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
-d '{"title": "Defensive Driving", "category": "elective", "default_credit_hours": 25.0, "description": "Advanced techniques for vehicle control"}' \
http://localhost:4000/v1/courses

export COURSE_ID="<the-course-id-you-just-copied>"


2. Get All Courses (GET)
curl -i -H "Authorization: Bearer $TOKEN" http://localhost:4000/v1/courses

3. Get One Course (GET by ID)
curl -i -H "Authorization: Bearer $TOKEN" http://localhost:4000/v1/courses/$COURSE_ID

4. Update Course (PATCH)
curl -i -X PATCH -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
-d '{"category": "mandatory"}' \
http://localhost:4000/v1/courses/$COURSE_ID

5. Delete Course (DELETE)
curl -i -X DELETE -H "Authorization: Bearer $TOKEN" http://localhost:4000/v1/courses/$COURSE_ID

--------------------------------------------------------------------------------------

FACILITATOR ENDPOINTS
----------------------
Phase 4: Testing Facilitator Endpoints
1. Create Facilitator (POST)
curl -i -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
-d '{"first_name": "Jane", "last_name": "Smith", "notes": "Specializes in practical exercises."}' \
http://localhost:4000/v1/facilitators

export FACILITATOR_ID="<the-facilitator-id-you-just-copied>"

2. GET, Update, Delete Facilitators...
Again, follow the same pattern for Get All, Get One, Update, and Delete for facilitators, using the `$FACILITATOR_ID`.

GET all:
curl -i -H "Authorization: Bearer $TOKEN" http://localhost:4000/v1/facilitators

GET by ID:
curl -i -H "Authorization: Bearer $TOKEN" http://localhost:4000/v1/facilitators/$FACILITAOR_ID

Update:
curl -i -X PATCH -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
-d '{"notes": "Lead instructor for driving courses."}' \
http://localhost:4000/v1/facilitators/$FACILITATOR_ID

Delete:
curl -i -X DELETE -H "Authorization: Bearer $TOKEN" http://localhost:4000/v1/facilitators/$FACILITATOR_ID

---------------------------------------------------------------------------------------------------

SESSION ENDPOINTS
------------------
Phase 5: Testing Session Endpoints
1. Create Session (POST)
Sessions depend on a course, so make sure you have a COURSE_ID from the previous step.

curl -i -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
-d '{"course_id": "'$COURSE_ID'", "start_datetime": "2025-11-10T09:00:00Z", "end_datetime": "2025-11-14T17:00:00Z", "location_text": "Training Room 1"}' \
http://localhost:4000/v1/sessions
  
export SESSION_ID="<the-session-id-you-just-copied>"

2. Get, Update, Delete Sessions...
You can follow the exact same pattern as above for Get All, Get One, Update, and Delete for sessions, using the `$SESSION_ID`.

GET all:
curl -i -H "Authorization: Bearer $TOKEN" http://localhost:4000/v1/sessions

Get by ID:
curl -i -H "Authorization: Bearer $TOKEN" http://localhost:4000/v1/sessions/$SESSION_ID

Update:
curl -i -X PATCH -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
-d '{"location_text": "conference room 2"}' \
http://localhost:4000/v1/sessions/$SESSION_ID

Delete:
curl -i -X DELETE -H "Authorization: Bearer $TOKEN" http://localhost:4000/v1/sessions/$SESSION_ID

**WRAP HANDLERS THAT NEED PROTECTION.**
eg. Create, get, update, delete, list handlers (try for officers first)

After you do this, if you try to run any of the curl commands from above without the -H "Authorization: Bearer $TOKEN" header, you will correctly receive a 401 Unauthorized error.