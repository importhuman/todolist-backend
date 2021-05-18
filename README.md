This is the backend repository for the to-do list web application [deployed here.](https://mighty-fjord-07080.herokuapp.com/)

## Stack

- Frontend: React
- Backend: Go
- Database: PostgreSQL
- Authentication: Auth0
- Deployment: Heroku

## How it works

- The user signs in to their account. The interface is provided by React, the authentication is handled via Auth0 and Go.
- Through the Auth0 pipeline, a custom field is added to the JSON web token (JWT) used for authentication.
- The user is added to the PostGreSQL database, thus creating their account.
- The user can now access the API endpoints to view their tasks, mark them as done, add, modify, or delete them. These API endpoints are protected, that is, they cannot be accessed without the appropriate JWT.
- The changes made in the user interface are reflected in the database via the APIs, ensuring the data is stored even when the user signs out.


