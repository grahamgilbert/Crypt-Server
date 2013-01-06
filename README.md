Crypt-Server
============
This is the server component for [Crypt](https://github.com/grahamgilbert/Crypt).

__It is not considered production quality code yet as it has not undergone extensive testing. Please use with caution. It has however, been used at pebble.it internally with no issues.__

##Installation instructions
Installation instructions are over on the [Wiki](https://github.com/grahamgilbert/Crypt-Server/wiki)

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
- Write upgrade documentaton (essentially, install new client, pull new code, run python manage.py migrate when in the virtualenv)