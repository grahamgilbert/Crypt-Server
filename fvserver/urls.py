# from django.conf.urls import include, url

# Uncomment the next two lines to enable the admin:
from django.contrib import admin
# admin.autodiscover()
import django.contrib.auth.views as auth_views
import django.contrib.admindocs.urls as admindocs_urls
from django.urls import path, include

app_name = 'fvserver'

urlpatterns = [
    # Examples:

    # url(r'^macnamer/', include('macnamer.foo.urls')),
    url(r'^login/$', auth_views.LoginView.as_view(), name='login'),
    url(r'^logout/$', auth_views.logout_then_login, name='logout'),
    url(r'^changepassword/$', auth_views.PasswordChangeView.as_view(), name='password_change'),
    url(r'^changepassword/done/$', auth_views.PasswordChangeDoneView.as_view(), name='password_change_done'),
   	url(r'^', include('server.urls')),
    # Uncomment the admin/doc line below to enable admin documentation:
    path('admin/doc/', include(admindocs_urls)),

    # Uncomment the next line to enable the admin:
    url(r'^admin/', admin.site.urls),
    #url(r'^$', 'namer.views.index', name='home'),

]
