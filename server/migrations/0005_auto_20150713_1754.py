# -*- coding: utf-8 -*-
from __future__ import unicode_literals

from django.db import models, migrations


class Migration(migrations.Migration):

    dependencies = [
        ('server', '0004_auto_20150713_1216'),
    ]

    operations = [
        migrations.AlterField(
            model_name='request',
            name='secret',
            field=models.ForeignKey(to='server.Secret', on_delete=models.CASCADE),
        ),
    ]
