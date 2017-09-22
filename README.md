Crypt-Server
============
__[Crypt][1]__ is a system for centrally storing FileVault 2 recovery keys. It is made up of a client app, and a Django web app for storing the keys.

This Docker image contains the fully configured Crypt Django web app. A default admin user has been preconfigured, use admin/password to login.
If you intend on using the server for anything semi-serious it is a good idea to change the password or add a new admin user and delete the default one.

__Changes in this version__
=================

- 10.7 is no longer supported.
- Improved logging on errors.
- Improved user feedback during long operations (such as enabling FileVault).

__Client__
====
The client is written in Pyobjc, and makes use of the built in fdesetup on OS X 10.8 and higher. An example login hook is provided to see how this could be implemented in your organisation.

__Features__
=======
- If escrow fails for some reason, the recovery key is stored on disk and a Launch Daemon will attempt to escrow the key periodically.
- If the app cannot contact the server, it can optionally quit.
- If FileVault is already enabled, the app will quit.


  [1]: https://github.com/grahamgilbert/Crypt

## Installation instructions
It is recommended that you use [Docker](https://github.com/grahamgilbert/Crypt-Server/blob/master/docs/Docker.md) to run this, but if you wish to run directly on a host, installation instructions are over on the in the [docs directory](https://github.com/grahamgilbert/Crypt-Server/blob/master/docs/Installation_on_Ubuntu_1404.md)

## Settings

* ``SEND_EMAIL`` - Crypt Server can send email notifcations when secrets are requested and approved. Set ``SEND_EMAIL`` to True, and set ``HOST_NAME`` to your server's host and URL scheme (e.g. ``https://crypt.example.com``). For configuring your email settings, see the [Django documentation](https://docs.djangoproject.com/en/1.9/ref/settings/#std:setting-EMAIL_HOST).

* ``APPROVE_OWN`` - By default, users with approval permissons can approve their own key requests. By setting this to False in settings.py (or by using the `DOCKER_CRYPT_APPROVE_OWN` environment variable with Docker), users cannot approve their own requests.

* ``ALL_APPROVE`` - By default, users need to be explicitly given approval permissions to approve key retrieval requests. By setting this to True in settings.py (or by using the `DOCKER_CRYPT_ALL_APPROVE` environment variable with Docker), all users are given this permission when they log in.

## New features in latest release
- Records Bonjour Name of Macs submitting keys
- Introduces the can_approve permission - users must have this permission to authorise key retrieval
- Key retrievals are logged

## Todo
- Email user when their request is approved or denied
- Move 7 day allowance into settings.py so it can be changed


## Screenshots
Main Page:
![Crypt Main Page](https://raw.github.com/grahamgilbert/Crypt-Server/master/docs/images/home.png)

User Computer Info:
![User computer info](https://raw.github.com/grahamgilbert/Crypt-Server/master/docs/images/user_computer_info.png)

Admin Computer Info:
![Admin computer info](https://raw.github.com/grahamgilbert/Crypt-Server/master/docs/images/admin_computer_info.png)

User Key Request:
![Userkey request](https://raw.github.com/grahamgilbert/Crypt-Server/master/docs/images/user_key_request.png)

Manage Requests:
![Manage Requests](https://raw.github.com/grahamgilbert/Crypt-Server/master/docs/images/manage_requests.png)

Approve Request:
![Approve Request](https://raw.github.com/grahamgilbert/Crypt-Server/master/docs/images/approve_request.png)

Key Retrieval:
![Key Retrieval](https://raw.github.com/grahamgilbert/Crypt-Server/master/docs/images/key_retrieval.png)
