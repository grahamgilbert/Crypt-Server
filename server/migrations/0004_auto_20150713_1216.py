# -*- coding: utf-8 -*-
from __future__ import unicode_literals
from django.shortcuts import get_object_or_404
from django.db import models, migrations
import django.db.models.deletion


class Migration(migrations.Migration):

    dependencies = [
        ('server', '0003_auto_20150713_1215'),
    ]

    operations = [
        migrations.RemoveField(
            model_name='computer',
            name='recovery_key',
        ),
        migrations.RemoveField(
            model_name='request',
            name='computer',
        ),
    ]
