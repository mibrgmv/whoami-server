# gateway
recovery, logging, CORS and authorization middlewares are user, in that order from a request's prerspective. auth middleware also puts `user_id` into Metadata which is later extracted by an interceptor
# services
all services share logging and recovery interceptors, and some handles are excempt from authorization via a selector interceptor
## quiz and questions service
- all handles except fetching quizzes require authorization
- request to get questions for a specific quiz is cached
## user and authorization service
- all auth handles and create new user require do not require authorization
- request to get users list is cached
## history service
- all handles require authorization
