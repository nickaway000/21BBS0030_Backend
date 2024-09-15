tests-
Register-
curl -X POST http://localhost:9080/register \
     -H "Content-Type: application/json" \
     -d '{
           "email": "newuser@example.com",
           "password": "password123"
         }'

Login-
curl -X POST http://localhost:9080/login \
     -H "Content-Type: application/json" \
     -d '{
           "email": "newuser@example.com",
           "password": "password123"
         }'

curl -X POST "http://localhost:9080/upload" \
-H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im5ld3VzZXJAZXhhbXBsZS5jb20iLCJleHAiOjE3MjY0NzEwMDUsInVzZXJfaWQiOjF9.5JVHdvS6I8HTuvHvZu2J5FuZGYkHTbrpqvMqXflRCJU" \
-F "file=@/Users/nikhil//Downloads/prac.txt"

retrieve-
curl -X GET "http://localhost:9080/retrieve?file_id=FILE_ID" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"


Share-
curl -X GET "http://localhost:9080/share?file_id=FILE_ID" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

Search-
curl -X GET "http://localhost:9080/search?filename=IMG_3123" \
-H "Authorization: Bearer YOUR_JWT_TOKEN"
