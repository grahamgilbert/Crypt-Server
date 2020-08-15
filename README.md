Crypt-Server
============
__[Crypt][1]__ is a tool for securely storing secrets such as FileVault 2 recovery keys. It is made up of a client app, and a Django web app for storing the keys.

This Docker image contains the fully configured Crypt Django web app. A default admin user has been preconfigured, use admin/password to login.
If you intend on using the server for anything semi-serious it is a good idea to change the password or add a new admin user and delete the default one.

__Features__
=======
- Secrets are encrypted in the database
- All access is audited - all reasons for retrieval and approval are logged along side the users performing the actions
- Two step approval for retrieval of secrets is enabled by default
- Approval permission can be given to all users (so just any two users need to approve the retrieval) or a specific group of users


  [1]: https://github.com/grahamgilbert/Crypt

## Installation instructions
It is recommended that you use [Docker](https://github.com/grahamgilbert/Crypt-Server/blob/master/docs/Docker.md) to run this, but if you wish to run directly on a host, installation instructions are over in the [docs directory](https://github.com/grahamgilbert/Crypt-Server/blob/master/docs/Installation_on_Ubuntu_1404.md)

## Settings

All settings that would be entered into `settings.py` can also be passed into the Docker container as environment variables.

* ``FIELD_ENCRYPTION_KEY`` - The key to use when encrypting the secrets. This is required.

* ``SEND_EMAIL`` - Crypt Server can send email notifcations when secrets are requested and approved. Set ``SEND_EMAIL`` to True, and set ``HOST_NAME`` to your server's host and URL scheme (e.g. ``https://crypt.example.com``). For configuring your email settings, see the [Django documentation](https://docs.djangoproject.com/en/1.9/ref/settings/#std:setting-EMAIL_HOST).

* ``APPROVE_OWN`` - By default, users with approval permissons can approve their own key requests. By setting this to False in settings.py (or by using the `APPROVE_OWN` environment variable with Docker), users cannot approve their own requests.

* ``ALL_APPROVE`` - By default, users need to be explicitly given approval permissions to approve key retrieval requests. By setting this to True in `settings.py`, all users are given this permission when they log in.

* ``ROTATE_VIEWED_SECRETS`` - With a compatible client (such as Crypt 3.2.0 and greater), Crypt Server can instruct the client to rotate the secret and re-escrow it when the secret has been viewed. Enable by setting this to `True` or by using `ROTATE_VIEWED_SECRETS` and setting to `true`.


## Screenshots
Main Page:
![Crypt Main Page](https://raw.github.com/grahamgilbert/Crypt-Server/master/docs/images/home.png)

Computer Info:
![Computer info](https://raw.github.com/grahamgilbert/Crypt-Server/master/docs/images/admin_computer_info.png)

User Key Request:
![Userkey request](https://raw.github.com/grahamgilbert/Crypt-Server/master/docs/images/user_key_request.png)

Manage Requests:
![Manage Requests](https://raw.github.com/grahamgilbert/Crypt-Server/master/docs/images/manage_requests.png)

Approve Request:
![Approve Request](https://raw.github.com/grahamgilbert/Crypt-Server/master/docs/images/approve_request.png)

Key Retrieval:
![Key Retrieval](https://raw.github.com/grahamgilbert/Crypt-Server/master/docs/images/key_retrieval.png)
