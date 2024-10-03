from django.db import models
from django.contrib.auth.models import User, Permission
from django.contrib.contenttypes.models import ContentType
from encrypted_model_fields.fields import EncryptedCharField

from django.core.exceptions import ValidationError


# Create your models here.
class Computer(models.Model):
    # recovery_key = models.CharField(max_length=200, verbose_name="Recovery Key")
    serial = models.CharField(max_length=200, verbose_name="Serial Number", unique=True)
    username = models.CharField(max_length=200, verbose_name="User Name")
    computername = models.CharField(max_length=200, verbose_name="Computer Name")
    last_checkin = models.DateTimeField(blank=True, null=True)

    def __str__(self):
        return self.computername

    class Meta:
        ordering = ["serial"]
        permissions = (
            ("can_approve", ("Can approve requests to see encryption keys")),
        )


SECRET_TYPES = (
    ("recovery_key", "Recovery Key"),
    ("password", "Password"),
    ("unlock_pin", "Unlock PIN"),
)


class Secret(models.Model):
    computer = models.ForeignKey(Computer, on_delete=models.CASCADE)
    secret = EncryptedCharField(max_length=256)
    secret_type = models.CharField(
        max_length=256, choices=SECRET_TYPES, default="recovery_key"
    )
    date_escrowed = models.DateTimeField(auto_now_add=True)
    rotation_required = models.BooleanField(default=False)

    def validate_unique(self, *args, **kwargs):
        if (
            self.secret
            in [
                str(s)
                for s in self.__class__.objects.filter(
                    secret_type=self.secret_type, computer=self.computer
                )
            ]
            and not self.rotation_required
        ):
            raise ValidationError("already used")
        super(Secret, self).validate_unique(*args, **kwargs)

    def save(self, *args, **kwargs):
        self.validate_unique()
        super(Secret, self).save(*args, **kwargs)

    def __str__(self):
        return self.secret

    class Meta:
        ordering = ["-date_escrowed"]


class Request(models.Model):
    secret = models.ForeignKey(Secret, on_delete=models.PROTECT)
    # computer = models.ForeignKey(Computer, null=True, related_name='computers')
    requesting_user = models.ForeignKey(
        User, related_name="requesting_user", on_delete=models.CASCADE
    )
    approved = models.BooleanField(verbose_name="Approved?", null=True)
    auth_user = models.ForeignKey(
        User, null=True, related_name="auth_user", on_delete=models.PROTECT
    )
    reason_for_request = models.TextField()
    reason_for_approval = models.TextField(
        blank=True, null=True, verbose_name="Approval Notes"
    )
    date_requested = models.DateTimeField(auto_now_add=True)
    date_approved = models.DateTimeField(blank=True, null=True)
    current = models.BooleanField(default=True)

    def __str__(self):
        return "%s - %s" % (self.secret, self.requesting_user)
