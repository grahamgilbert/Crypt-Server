# -*- coding: utf-8 -*-
import datetime
from south.db import db
from south.v2 import SchemaMigration
from django.db import models


class Migration(SchemaMigration):

    def forwards(self, orm):
        # Removing unique constraint on 'Computer', fields ['serial']
        db.delete_unique('server_computer', ['serial'])


    def backwards(self, orm):
        # Adding unique constraint on 'Computer', fields ['serial']
        db.create_unique('server_computer', ['serial'])


    models = {
        'server.computer': {
            'Meta': {'ordering': "['serial']", 'object_name': 'Computer'},
            'id': ('django.db.models.fields.AutoField', [], {'primary_key': 'True'}),
            'last_checkin': ('django.db.models.fields.DateTimeField', [], {'null': 'True', 'blank': 'True'}),
            'recovery_key': ('django.db.models.fields.CharField', [], {'max_length': '200'}),
            'serial': ('django.db.models.fields.CharField', [], {'max_length': '200'})
        }
    }

    complete_apps = ['server']