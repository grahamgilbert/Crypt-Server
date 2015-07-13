# -*- coding: utf-8 -*-
from __future__ import unicode_literals
from django.shortcuts import get_object_or_404
from django.db import models, migrations

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

class Migration(migrations.Migration):

    dependencies = [
        ('server', '0001_initial'),
    ]

    operations = [
        migrations.CreateModel(
            name='Secret',
            fields=[
                ('id', models.AutoField(verbose_name='ID', serialize=False, auto_created=True, primary_key=True)),
                ('secret', models.CharField(max_length=256)),
                ('secret_type', models.CharField(default=b'recovery_key', max_length=256, choices=[(b'recovery_key', b'Recovery Key'), (b'password', b'Password')])),
                ('date_escrowed', models.DateTimeField(auto_now_add=True)),
            ],
        ),
        migrations.AddField(
            model_name='secret',
            name='computer',
            field=models.ForeignKey(to='server.Computer'),
        ),
        migrations.AddField(
            model_name='request',
            name='secret',
            field=models.ForeignKey(default=1, to='server.Secret'),
            preserve_default=False,
        ),
        migrations.RunPython(move_keys_and_requests),
        migrations.RemoveField(
            model_name='computer',
            name='recovery_key',
        ),
        migrations.RemoveField(
            model_name='request',
            name='computer',
        ),
    ]
