**Officer Endpoint:**

* Creating an officer:

BODY_OFFICER='{

&nbsp;   "regulation_number": "PC700",

&nbsp;   "first_name": "Rachel",

&nbsp;   "last_name": "Smith",

&nbsp;   "sex": "female",

&nbsp;   "rank_code": "PC"

}'



curl -i -X POST -H "Content-Type: application/json" -d "$BODY_OFFICER" http://localhost:4000/v1/officers



* Read (get) officer:
curl -i http://localhost:4000/v1/officers/$OFFICER_ID



* Update officer:

UPDATE_BODY_OFFICER='{"rank_code": "SGT"}'


curl -i -X PATCH -H "Content-Type: application/json" -d "$UPDATE_BODY_OFFICER" http://localhost:4000/v1/officers/$OFFICER_ID



* Delete officer:

curl -i -X DELETE http://localhost:4000/v1/officers/$OFFICER_ID

-------------------------------------------------------------------------------------

COURSES ENDPOINT 
-------------------
C R E A T E:
BODY_COURSE='{
    "title": "Defensive Driving",
    "category": "elective",
    "default_credit_hours": 16.0,
    "description": "Advanced techniques for vehicle control."
}'

curl -i -X POST -H "Content-Type: application/json" -d "$BODY_COURSE" http://localhost:4000/v1/courses


R E A D:
# Replace with the actual ID from the create step
COURSE_ID="a5eabf0d-8ddb-4bd2-a857-3782d3928296"

curl -i http://localhost:4000/v1/courses/$COURSE_ID


U P D A T E:
UPDATE_BODY_COURSE='{
    "title": "Advanced Defensive Driving"
}'

curl -i -X PATCH -H "Content-Type: application/json" -d "$UPDATE_BODY_COURSE" http://localhost:4000/v1/courses/$COURSE_ID


D E L E T E:
curl -i -X DELETE http://localhost:4000/v1/courses/$COURSE_ID

L I S T   A L L   C O U R S E S:
curl -i http://localhost:4000/v1/courses

--------------------------------------------------------------------------------------

FACILITATOR ENDPOINTS
----------------------

C R E A T E :
BODY_FACILITATOR='{
    "first_name": "Goofy",
    "last_name": "Goof",
    "notes": "Specializes in practical exercises."
}'

curl -i -X POST -H "Content-Type: application/json" -d "$BODY_FACILITATOR" http://localhost:4000/v1/facilitators


R E A D:
# Replace with the actual ID from the create step
FACILITATOR_ID="25eb8d72-bb19-4884-b8b3-0d83f4c806cf"

curl -i http://localhost:4000/v1/facilitators/$FACILITATOR_ID

U P D A T E:
UPDATE_BODY_FACILITATOR='{
    "notes": "Lead instructor for driving courses."
}'

curl -i -X PATCH -H "Content-Type: application/json" -d "$UPDATE_BODY_FACILITATOR" http://localhost:4000/v1/facilitators/$FACILITATOR_ID

D E L E T E:
curl -i -X DELETE http://localhost:4000/v1/facilitators/$FACILITATOR_ID

L I S T  A L L   F A C I L I T A T O R:
curl -i http://localhost:4000/v1/facilitators