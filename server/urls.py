from django.conf.urls import *
from server.views import *
urlpatterns = [
    #front. page
    url(r'^$', index, name='home'),

    # Add computer
    url(r'^new/computer/', new_computer, name='new_computer'),
    # Add secret
    url(r'^new/secret/(?P<computer_id>.+)/', new_secret, name='new_secret'),

    # secret info
    url(r'^info/secret/(?P<secret_id>.+)/', secret_info, name='secret_info'),

    #computerinfo
    url(r'^info/(?P<computer_id>.+)/', computer_info, name='computer_info'),

    #request
    url(r'^request/(?P<secret_id>.+)/', request, name='request'),
    #retrieve
    url(r'^retrieve/(?P<request_id>.+)/', retrieve, name='retrieve'),
    #approve
    url(r'^approve/(?P<request_id>.+)/', approve, name='approve'),
    # verify
    url(r'^verify/(?P<serial>.+)/(?P<secret_type>.+)/', verify, name='verify'),
    #checkin
    url(r'^checkin/', checkin, name='checkin'),
    #manage
    url(r'^manage-requests/', managerequests, name='managerequests'),
]
