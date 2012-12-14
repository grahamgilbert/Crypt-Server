from django.db import models

# Create your models here.
class Computer(models.Model):
    recovery_key = models.CharField(max_length=200, verbose_name="Recovery Key")
    serial = models.CharField(max_length=200, verbose_name="Serial Number")
    username = models.CharField(max_length=200, verbose_name="User Name")
    last_checkin = models.DateTimeField(blank=True,null=True)
    def __unicode__(self):
        return self.serial
    class Meta:
        ordering = ['serial']