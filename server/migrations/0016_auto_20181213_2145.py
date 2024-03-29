# Generated by Django 2.1.4 on 2018-12-13 21:45

from django.conf import settings
from django.db import migrations, models
import django.db.models.deletion


class Migration(migrations.Migration):

    dependencies = [("server", "0012_auto_20181128_2038")]

    operations = [
        migrations.AlterField(
            model_name="request",
            name="auth_user",
            field=models.ForeignKey(
                null=True,
                on_delete=django.db.models.deletion.PROTECT,
                related_name="auth_user",
                to=settings.AUTH_USER_MODEL,
            ),
        ),
        migrations.AlterField(
            model_name="request",
            name="secret",
            field=models.ForeignKey(
                on_delete=django.db.models.deletion.PROTECT, to="server.Secret"
            ),
        ),
    ]
