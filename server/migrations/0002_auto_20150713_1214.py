# -*- coding: utf-8 -*-
from __future__ import unicode_literals
from django.shortcuts import get_object_or_404
from django.db import models, migrations

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
        migrations.AlterField(
            model_name='request',
            name='computer',
            field=models.ForeignKey(related_name='computers', to='server.Computer'),
        ),
        migrations.AddField(
            model_name='request',
            name='secret',
            field=models.ForeignKey(null=True, related_name='secrets', to='server.Secret'),
            preserve_default=False,
        ),
    ]
