# -*- coding: utf-8 -*-
from __future__ import unicode_literals
from django.shortcuts import get_object_or_404
from django.db import models, migrations
import django.db.models.deletion

def move_keys_and_requests(apps, schema_editor):
    seen_serials = [('dummy_serial', 'dummy_id')]
    Computer = apps.get_model("server", "Computer")
    Secret = apps.get_model("server", "Secret")
    Request = apps.get_model("server", "Request")
    for computer in Computer.objects.all():
        # if we've seen the serial before, get the computer that we saw before
        target_id = None
        for serial, id in seen_serials:
            if computer.serial == serial:
                target_id = id
                break
        if target_id == None:
            target_id = computer.id

        target_computer = get_object_or_404(Computer, pk=target_id)
        # create a new secret
        secret = Secret(computer=target_computer, secret=computer.recovery_key, date_escrowed=computer.last_checkin)
        secret.save()

        requests = Request.objects.filter(computer=computer)
        for request in requests:
            request.secret = secret
            request.save()

        if target_computer.id != computer.id:
            #Dupe computer, bin it
            computer.delete()

class Migration(migrations.Migration):

    dependencies = [
        ('server', '0002_auto_20150713_1214'),
    ]

    operations = [
        migrations.RunPython(move_keys_and_requests),
    ]
