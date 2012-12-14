from django.conf.urls.defaults import *

urlpatterns = patterns('server.views',
    #front. page
    url(r'^$', 'index', name='home'),
    #checkin
    url(r'^checkin/', 'checkin', name='checkin'),
)
