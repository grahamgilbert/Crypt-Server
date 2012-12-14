# -*- coding: utf-8 -*-
import datetime
from south.db import db
from south.v2 import SchemaMigration
from django.db import models


class Migration(SchemaMigration):

    def forwards(self, orm):
        # Adding model 'Computer'
        db.create_table('server_computer', (
            ('id', self.gf('django.db.models.fields.AutoField')(primary_key=True)),
            ('recovery_key', self.gf('django.db.models.fields.CharField')(max_length=200)),
            ('serial', self.gf('django.db.models.fields.CharField')(unique=True, max_length=200)),
            ('last_checkin', self.gf('django.db.models.fields.DateTimeField')(null=True, blank=True)),
        ))
        db.send_create_signal('server', ['Computer'])


    def backwards(self, orm):
        # Deleting model 'Computer'
        db.delete_table('server_computer')


    models = {
        'server.computer': {
            'Meta': {'ordering': "['serial']", 'object_name': 'Computer'},
            'id': ('django.db.models.fields.AutoField', [], {'primary_key': 'True'}),
            'last_checkin': ('django.db.models.fields.DateTimeField', [], {'null': 'True', 'blank': 'True'}),
            'recovery_key': ('django.db.models.fields.CharField', [], {'max_length': '200'}),
            'serial': ('django.db.models.fields.CharField', [], {'unique': 'True', 'max_length': '200'})
        }
    }

    complete_apps = ['server']