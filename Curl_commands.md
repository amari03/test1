# Phase 1: Setup - Becoming an Authorized User
## Step 1: Register a New User
```Bash
curl -i -X POST -H "Content-Type: application/json" \
-d '{"email": "testing123@example.com", "password": "password123", "role": "admin"}' \
http://localhost:4000/v1/users
```

## Step 2: Activate the User
-Go to your Mailtrap inbox.  
-Open the welcome email sent to test@example.com.  
-Copy the 26-character activation token from the email body.  

#### Replace YOUR_ACTIVATION_TOKEN with the token you copied from the email
```Bash
curl -i -X PUT -H "Content-Type: application/json" \
-d '{"token": "YOUR_ACTIVATION_TOKEN"}' \
http://localhost:4000/v1/users/activated
```

## Step 3: Authenticate and Get a Bearer Token
```Bash
curl -i -X POST -H "Content-Type: application/json" -d '{"email": "testing123@example.com", "password": "password123"}' http://localhost:4000/v1/tokens/authentication
```

## Step 4: Store the Token in a Shell Variable
#### Replace YOUR_BEARER_TOKEN with the token from the step above
```Bash
export TOKEN="YOUR_BEARER_TOKEN"
```

#### You can verify it's stored by running:
```Bash
echo $TOKEN
```

# Phase 2: Testing Officer Endpoints
## Step 1: Create Officer (POST)
```Bash
curl -i -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d '{"regulation_number": "PC750", "first_name": "Marcus", "last_name": "Moore", "sex": "male", "rank_code": "PC"}' http://localhost:4000/v1/officers
```
```Bash
export OFFICER_ID="<the-id-you-just-copied>"
```

## Step 2: Get All Officers (GET)
```Bash
curl -i -H "Authorization: Bearer $TOKEN" http://localhost:4000/v1/officers
```

## Step 3: Get One Officer (GET by ID)
```Bash
curl -i -H "Authorization: Bearer $TOKEN" http://localhost:4000/v1/officers/$OFFICER_ID
```

## Pagination
```Bash
curl -i -H "Authorization: Bearer $TOKEN" "http://localhost:4000/v1/officers?page=1&page_size=2"
curl -i -H "Authorization: Bearer $TOKEN" "http://localhost:4000/v1/officers?content=PC"
```

## Step 4: Update Officer (PATCH)
```Bash
curl -i -X PATCH \
-H "Content-Type: application/json" \
-H "Authorization: Bearer $TOKEN" \
-d '{"rank_code": "SGT"}' \
http://localhost:4000/v1/officers/$OFFICER_ID
```

## Step 5: Delete Officer (DELETE)
```Bash
curl -i -X DELETE -H "Authorization: Bearer $TOKEN" http://localhost:4000/v1/officers/$OFFICER_ID
```

-------------------------------------------------------------------------------------

# Phase 3: Testing Course Endpoints
## Step 1: Create Course (POST)
```Bash
curl -i -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
-d '{"title": "Advanced Interview and Interrogation", "category": "elective", "default_credit_hours": 40.0, "description": "Techniques for ethical and effective information gathering from victims, witnesses, and suspects."}' \
http://localhost:4000/v1/courses
```
```Bash
export COURSE_ID="<the-course-id-you-just-copied>"
```

## Step 2: Get All Courses (GET)
```Bash
curl -i -H "Authorization: Bearer $TOKEN" http://localhost:4000/v1/courses
```

## Step 3: Get One Course (GET by ID)
```Bash
curl -i -H "Authorization: Bearer $TOKEN" http://localhost:4000/v1/courses/$COURSE_ID
```

## Pagination
```Bash
curl -i -H "Authorization: Bearer $TOKEN" "http://localhost:4000/v1/courses?page=1&page_size=2"
```

## Step 4: Update Course (PATCH)
```Bash
curl -i -X PATCH -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
-d '{"category": "mandatory"}' \
http://localhost:4000/v1/courses/$COURSE_ID
```

## Step 5: Delete Course (DELETE)
```Bash
curl -i -X DELETE -H "Authorization: Bearer $TOKEN" http://localhost:4000/v1/courses/$COURSE_ID
```

--------------------------------------------------------------------------------------

# Phase 4: Testing Facilitator Endpoints
## Step 1: Create Facilitator (POST)
```Bash
curl -i -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
-d '{"first_name": "Jane", "last_name": "Smith", "notes": "Specializes in practical exercises."}' \
http://localhost:4000/v1/facilitators
```
```Bash
export FACILITATOR_ID="<the-facilitator-id-you-just-copied>"
```

## Step 2: GET, Update, Delete Facilitators...
Again, follow the same pattern for Get All, Get One, Update, and Delete for facilitators, using the `$FACILITATOR_ID`.

1. GET all:
```Bash
curl -i -H "Authorization: Bearer $TOKEN" http://localhost:4000/v1/facilitators
```

2. GET by ID:
```Bash
curl -i -H "Authorization: Bearer $TOKEN" http://localhost:4000/v1/facilitators/$FACILITAOR_ID
```

## Pagination
```Bash
curl -i -H "Authorization: Bearer $TOKEN" "http://localhost:4000/v1/facilitators?page=1&page_size=2"
```

3. Update:
```Bash
curl -i -X PATCH -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
-d '{"notes": "Lead instructor for driving courses."}' \
http://localhost:4000/v1/facilitators/$FACILITATOR_ID
```

4. Delete:
```Bash
curl -i -X DELETE -H "Authorization: Bearer $TOKEN" http://localhost:4000/v1/facilitators/$FACILITATOR_ID
```

---------------------------------------------------------------------------------------------------

# Phase 5: Testing Session Endpoints
## Step 1: Create Session (POST)
Sessions depend on a course, so make sure you have a `COURSE_ID` from the previous step.
```Bash
curl -i -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
-d '{"course_id": "'$COURSE_ID'", "start_datetime": "2025-11-10T09:00:00Z", "end_datetime": "2025-11-14T17:00:00Z", "location_text": "Training Room 1"}' \
http://localhost:4000/v1/sessions
```
 ```Bash 
export SESSION_ID="<the-session-id-you-just-copied>"
```

## Step 2: Get, Update, Delete Sessions...
You can follow the exact same pattern as above for Get All, Get One, Update, and Delete for sessions, using the `$SESSION_ID`.

1. GET all:
```Bash
curl -i -H "Authorization: Bearer $TOKEN" http://localhost:4000/v1/sessions
```

2. Get by ID:
```Bash
curl -i -H "Authorization: Bearer $TOKEN" http://localhost:4000/v1/sessions/$SESSION_ID
```

3. Update:
```Bash
curl -i -X PATCH -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
-d '{"location_text": "conference room 2"}' \
http://localhost:4000/v1/sessions/$SESSION_ID
```

4. Delete:
```Bash
curl -i -X DELETE -H "Authorization: Bearer $TOKEN" http://localhost:4000/v1/sessions/$SESSION_ID
```

**WRAP HANDLERS THAT NEED PROTECTION.**  

After you do this, if you try to run any of the curl commands from above without the `-H "Authorization: Bearer $TOKEN"` header, you will correctly receive a 401 Unauthorized error.

**example:** creating an officer

```Bash
curl -i -X POST -H "Content-Type: application/json" -d '{"regulation_number": "PC800", "first_name": "Jonathan", "last_name": "Doe", "sex": "male", "rank_code": "PC"}' http://localhost:4000/v1/officers
```

**UNIT TESTS**
```Bash
go test -v ./internal/data/...
```