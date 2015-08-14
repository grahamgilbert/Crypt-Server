# -*- coding: utf-8 -*-
from __future__ import unicode_literals

from django.db import models, migrations


class Migration(migrations.Migration):

    dependencies = [
        ('server', '0007_auto_20150714_0822'),
    ]

    operations = [
        migrations.AlterModelOptions(
            name='secret',
            options={'ordering': ['-date_escrowed']},
        ),
    ]
