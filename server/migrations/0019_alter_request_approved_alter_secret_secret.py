# Generated by Django 4.1.2 on 2022-11-10 17:28

from django.db import migrations, models
import encrypted_model_fields.fields


class Migration(migrations.Migration):

    dependencies = [
        ("server", "0018_auto_20201029_2134"),
    ]

    operations = [
        migrations.AlterField(
            model_name="request",
            name="approved",
            field=models.BooleanField(null=True, verbose_name="Approved?"),
        ),
        migrations.AlterField(
            model_name="secret",
            name="secret",
            field=encrypted_model_fields.fields.EncryptedCharField(),
        ),
    ]
