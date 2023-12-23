# savingsquadsbackend Capstone

## Makefile Usage

The Makefile included in this project simplifies some of the tasks related to building and running the application. Here's how to use it:

- `make build`: This command compiles the Go application and outputs the binary to the `bin/` directory.

- `make run`: This command first builds the application using the `build` target, then runs the compiled binary.

- `make docker-build`: This command builds a Docker image of your application. It uses the Dockerfile in the current directory and tags the image with the name `savingsquadsbackend`.

- `make docker-run`: This command first builds the Docker image using the `docker-build` target, then runs the Docker image as a container. The container listens on port 3000.

- `make docker-stop`: This command stops the running Docker container.

- `make docker-clean`: This command first stops the running Docker container using the `docker-stop` target, then removes the Docker container.

To use these commands, navigate to the project directory in your terminal and type the command you want to run. For example, to build and run the Docker container, you would type `make docker-run`.

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
1. Calculation service
2. User - get best voucher 
3. Touch up on admin (if necessary)
4. ?Migration for mongodb
6. ?Testing