from django.conf.urls.defaults import *

urlpatterns = patterns('server.views',
    #front. page
    url(r'^$', 'index', name='home'),
    #computerinfo
    url(r'^info/(?P<computer_id>.+)/', 'computer_info', name='computer_info'),
    #request
    url(r'^request/(?P<computer_id>.+)/', 'request', name='request'),
    #retrieve
    url(r'^retrieve/(?P<request_id>.+)/', 'retrieve', name='retrieve'),
    #approve
    url(r'^approve/(?P<request_id>.+)/', 'approve', name='approve'),
    #checkin
    url(r'^checkin/', 'checkin', name='checkin'),
    #manage
    url(r'^manage-requests/', 'managerequests', name='managerequests'),
)
