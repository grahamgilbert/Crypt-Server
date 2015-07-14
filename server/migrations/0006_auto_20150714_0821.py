# -*- coding: utf-8 -*-
from __future__ import unicode_literals

from django.db import models, migrations
import django_extensions.db.fields.encrypted


class Migration(migrations.Migration):

    dependencies = [
        ('server', '0005_auto_20150713_1754'),
    ]

    operations = [
        migrations.AlterField(
            model_name='secret',
            name='secret',
            field=django_extensions.db.fields.encrypted.EncryptedCharField(max_length=256),
        ),
    ]
