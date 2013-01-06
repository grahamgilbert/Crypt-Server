# -*- coding: utf-8 -*-
import datetime
from south.db import db
from south.v2 import SchemaMigration
from django.db import models


class Migration(SchemaMigration):

    def forwards(self, orm):
        # Adding field 'Computer.computername'
        db.add_column('server_computer', 'computername',
                      self.gf('django.db.models.fields.CharField')(default='Mac', max_length=200),
                      keep_default=False)


    def backwards(self, orm):
        # Deleting field 'Computer.computername'
        db.delete_column('server_computer', 'computername')


    models = {
        'server.computer': {
            'Meta': {'ordering': "['serial']", 'object_name': 'Computer'},
            'computername': ('django.db.models.fields.CharField', [], {'max_length': '200'}),
            'id': ('django.db.models.fields.AutoField', [], {'primary_key': 'True'}),
            'last_checkin': ('django.db.models.fields.DateTimeField', [], {'null': 'True', 'blank': 'True'}),
            'recovery_key': ('django.db.models.fields.CharField', [], {'max_length': '200'}),
            'serial': ('django.db.models.fields.CharField', [], {'max_length': '200'}),
            'username': ('django.db.models.fields.CharField', [], {'max_length': '200'})
        }
    }

    complete_apps = ['server']