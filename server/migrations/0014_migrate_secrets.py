# -*- coding: utf-8 -*-
from __future__ import unicode_literals

from server.models import *
from django.db import migrations, models
from django.conf import settings


def migrate_secrets(apps, schema_editor):
    """
    Migrate the secrets to the new field
    """

    if not hasattr(settings, 'FIELD_ENCRYPTION_KEY'):
        raise Exception('FIELD_ENCRYPTION_KEY not set in settings.py')
    
    if settings.FIELD_ENCRYPTION_KEY == '':
        raise Exception('FIELD_ENCRYPTION_KEY not configured correctly in settings.py')

    Secret = apps.get_model("server", "Secret")
    secrets_to_update = Secret.objects.all()
    for secret_to_update in secrets_to_update:
        secret_to_update.new_secret = secret_to_update.secret
        secret_to_update.save()


class Migration(migrations.Migration):

    dependencies = [
        ('server', '0013_secret_new_secret'),
    ]

    operations = [
        migrations.RunPython(migrate_secrets),
    ]
