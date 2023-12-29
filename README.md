# savingsquadsbackend Capstone

## Routes

### Admin Routes

- `GET /v1/vouchers`: Fetch all vouchers. Requires authentication.
- `POST /v1/vouchers`: Create a new voucher. Requires authentication.
- `GET /v1/vouchers/{id}`: Fetch a voucher by its ID. Requires authentication.
- `DELETE /v1/vouchers/{id}`: Delete a voucher by its ID. Requires authentication.

### User Routes

- `POST /v1/users/register`: Register a new user.
- `POST /v1/users/login`: Login a user.
- `POST /v1/users/logout`: Logout a user. Requires authentication.
- `GET /v1/users/vouchers`: Get all vouchers of a user. Requires authentication.
- `PUT /v1/users/vouchers/{id}/redeem`: Redeem a voucher for a user. Requires authentication.
- `PUT /v1/users/vouchers/{id}/use`: Use a voucher for a user. Requires authentication.
- `GET /v1/users/points`: Get the points of a user. Requires authentication.
- `PUT /v1/users/points`: Add points to a user. Requires authentication.
- `POST /v1/users/points/redeem`: Redeem points for a voucher. Requires authentication.
- `GET /v1/users/vouchers/best`: (TODO) Get the best voucher for a user. Requires authentication.

## To do list
1. Integrate calculation service
2. User - get best voucher 
3. Touch up on admin (if necessary)
4. To complete unit test for handler, unit test for data layer, unit test for middleware, integration and end-to-end testing - using docker and docker-compose
5. Explore additional features
6. Explore go swagger for documenting API