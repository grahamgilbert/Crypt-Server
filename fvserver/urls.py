# from django.conf.urls import include, url

# Uncomment the next two lines to enable the admin:
from django.contrib import admin

# admin.autodiscover()
import django.contrib.auth.views as auth_views
import django.contrib.admindocs.urls as admindocs_urls
from django.urls import path, include

app_name = "fvserver"

urlpatterns = [
    path("login/", auth_views.LoginView.as_view(), name="login"),
    path("logout/", auth_views.logout_then_login, name="logout"),
    path(
        "changepassword/",
        auth_views.PasswordChangeView.as_view(),
        name="password_change",
    ),
    path(
        "changepassword/done/",
        auth_views.PasswordChangeDoneView.as_view(),
        name="password_change_done",
    ),
    path("", include("server.urls")),
    # Uncomment the admin/doc line below to enable admin documentation:
    path("admin/doc/", include(admindocs_urls)),
    # Uncomment the next line to enable the admin:
    path("admin/", admin.site.urls),
]
