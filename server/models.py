from django.db import models
from django.contrib.auth.models import User, Permission
from django.contrib.contenttypes.models import ContentType
from django_extensions.db.fields.encrypted import EncryptedCharField

# Create your models here.
class Computer(models.Model):
    #recovery_key = models.CharField(max_length=200, verbose_name="Recovery Key")
    serial = models.CharField(max_length=200, verbose_name="Serial Number")
    username = models.CharField(max_length=200, verbose_name="User Name")
    computername = models.CharField(max_length=200, verbose_name="Computer Name")
    last_checkin = models.DateTimeField(blank=True,null=True)
    def __unicode__(self):
        return self.computername
    class Meta:
        ordering = ['serial']
        permissions = (
                    ('can_approve', (u'Can approve requests to see encryption keys')),
                )

SECRET_TYPES = (('recovery_key', 'Recovery Key'),
                ('password', 'Password'))

class Secret(models.Model):
    computer = models.ForeignKey(Computer)
    secret = EncryptedCharField(max_length=256)
    secret_type =  models.CharField(max_length=256, choices=SECRET_TYPES, default='recovery_key')
    date_escrowed = models.DateTimeField(auto_now_add=True)
    def __unicode__(self):
        return self.secret
    class Meta:
        ordering = ['-date_escrowed']

class Request(models.Model):
    secret = models.ForeignKey(Secret)
    #computer = models.ForeignKey(Computer, null=True, related_name='computers')
    requesting_user = models.ForeignKey(User, related_name='requesting_user')
    approved = models.NullBooleanField(verbose_name="Approved?")
    auth_user = models.ForeignKey(User, null=True, related_name='auth_user')
    reason_for_request = models.TextField()
    reason_for_approval = models.TextField(blank=True,null=True, verbose_name="Approval Notes")
    date_requested = models.DateTimeField(auto_now_add=True)
    date_approved = models.DateTimeField(blank=True,null=True)
    current = models.BooleanField(default=True)

    def __unicode__(self):
        return "%s - %s" % (self.secret, self.requesting_user)
