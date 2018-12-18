# -*- coding: utf-8 -*-
from __future__ import unicode_literals
from django.shortcuts import get_object_or_404
from django.db import models, migrations


def encrypt_secrets(apps, schema_editor):

    Secret = apps.get_model("server", "Secret")

    for secret in Secret.objects.all():
        secret.save()


class Migration(migrations.Migration):

    dependencies = [("server", "0006_auto_20150714_0821")]

    operations = [migrations.RunPython(encrypt_secrets)]
