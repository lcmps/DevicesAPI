# DevicesAPI

- Go v1.25+
- Docker
- Docker Compose

## Proposal
develop a REST API capable of persisting and managing device resources.

### Device Domain
In PostgreSQL

- ID `UUID`
- Name `varchar(250)`
- Brand `varchar(250)`
- CreatedAt `TIMESTAMPTZ`
- State: `ENUM`
    - Available
    - In-Use
    - Inactive

### Supported Functionalities
- Create a new device. `POST`
- Fully and/or partially update an existing device. `PUT`
- Fetch a single device. `GET`
- Fetch all devices. `GET`
- Fetch devices by brand. `GET`
- Fetch devices by state. `GET`
- Delete a single device. `DELETE`

### Domain Validations
- Creation time cannot be updated.
- Name and brand properties cannot be updated if the device is in use.
- In use devices cannot be deleted.