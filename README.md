Crypt-Server
============
This is the server component for [Crypt](https://github.com/grahamgilbert/Crypt).

##Installation instructions
Installation instructions are over on the in the [docs directory](https://github.com/grahamgilbert/Crypt-Server/blob/master/docs/Installation_on_Ubuntu_12.md)

##Acknowledgements
Many thanks to my lovely employers at [pebble.it](http://pebbleit.com) for letting me release this.

##New features in latest release
- Records Bonjour Name of Macs submitting keys
- Introduces the can_approve permission - users must have this permission to authorise key retrieval
- Key retrievals are logged

##Todo
- Email approvers when a new request is submitted
- Email user when their request is approved or denied
- Move 7 day allowance into settings.py so it can be changed

##Screenshots
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

