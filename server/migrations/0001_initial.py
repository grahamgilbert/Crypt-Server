# -*- coding: utf-8 -*-
from __future__ import unicode_literals

from django.db import models, migrations
from django.conf import settings


class Migration(migrations.Migration):

    dependencies = [migrations.swappable_dependency(settings.AUTH_USER_MODEL)]

    operations = [
        migrations.CreateModel(
            name="Computer",
            fields=[
                (
                    "id",
                    models.AutoField(
                        verbose_name="ID",
                        serialize=False,
                        auto_created=True,
                        primary_key=True,
                    ),
                ),
                (
                    "recovery_key",
                    models.CharField(max_length=200, verbose_name=b"Recovery Key"),
                ),
                (
                    "serial",
                    models.CharField(max_length=200, verbose_name=b"Serial Number"),
                ),
                (
                    "username",
                    models.CharField(max_length=200, verbose_name=b"User Name"),
                ),
                (
                    "computername",
                    models.CharField(max_length=200, verbose_name=b"Computer Name"),
                ),
                ("last_checkin", models.DateTimeField(null=True, blank=True)),
            ],
            options={
                "ordering": ["serial"],
                "permissions": (
                    ("can_approve", "Can approve requests to see encryption keys"),
                ),
            },
        ),
        migrations.CreateModel(
            name="Request",
            fields=[
                (
                    "id",
                    models.AutoField(
                        verbose_name="ID",
                        serialize=False,
                        auto_created=True,
                        primary_key=True,
                    ),
                ),
                ("approved", models.NullBooleanField(verbose_name=b"Approved?")),
                ("reason_for_request", models.TextField()),
                (
                    "reason_for_approval",
                    models.TextField(
                        null=True, verbose_name=b"Approval Notes", blank=True
                    ),
                ),
                ("date_requested", models.DateTimeField(auto_now_add=True)),
                ("date_approved", models.DateTimeField(null=True, blank=True)),
                ("current", models.BooleanField(default=True)),
                (
                    "auth_user",
                    models.ForeignKey(
                        related_name="auth_user",
                        to=settings.AUTH_USER_MODEL,
                        null=True,
                        on_delete=models.CASCADE,
                    ),
                ),
                (
                    "computer",
                    models.ForeignKey(to="server.Computer", on_delete=models.CASCADE),
                ),
                (
                    "requesting_user",
                    models.ForeignKey(
                        related_name="requesting_user",
                        to=settings.AUTH_USER_MODEL,
                        on_delete=models.CASCADE,
                    ),
                ),
            ],
        ),
    ]
