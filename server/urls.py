from django.urls import path
from . import views

app_name = 'server'

urlpatterns = [
    #front. page
    path('', views.index, name='home'),

    # Add computer
    path('new/computer/', views.new_computer, name='new_computer'),
    # Add secret
    path('new/secret/<int:computer_id>/', views.new_secret, name='new_secret'),

    # secret info
    path('info/secret/<int:secret_id>/', views.secret_info, name='secret_info'),

    #computerinfo
    path('info/<int:computer_id>/', views.computer_info, name='computer_info'),
    path('info/<str:serial>', views.computer_info, name='computer_info'),

    #request
    path('request/<int:secret_id>/', views.request, name='request'),
    #retrieve
    path('retrieve/<int:request_id>/', views.retrieve, name='retrieve'),
    #approve
    path('approve/<int:request_id>/', views.approve, name='approve'),
    # verify
    path('verify/<str:serial>/<str:secret_type>/', views.verify, name='verify'),
    #checkin
    path('checkin/', views.checkin, name='checkin'),
    #manage
    path('manage-requests/', views.managerequests, name='managerequests'),
]
