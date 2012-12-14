from models import *
from django.contrib.auth.decorators import login_required, permission_required
from django.template import RequestContext, Template, Context
from django.utils import simplejson
from django.views.decorators.csrf import csrf_exempt, csrf_protect
from django.http import HttpResponse, Http404
from django.contrib.auth.models import Permission, User
from django.conf import settings
from django.shortcuts import render_to_response
from datetime import datetime
# Create your views here.

##index view
@login_required 
def index(request):
    #show table with all the keys
    computers = Computer.objects.all()
    c = {'user': request.user, 'computers':computers, }
    return render_to_response('server/index.html', c, context_instance=RequestContext(request)) 

##checkin view
@csrf_exempt
def checkin(request):
    try:
        serial_num = request.POST['serial']
    except:
        raise Http500
    try:
        recovery_pass = request.POST['recovery_password']
    except:
        raise Http500
        
    try:
        user_name = request.POST['username']
    except:
        raise Http500
    computer = Computer(recovery_key=recovery_pass, serial=serial_num, last_checkin = datetime.now(), username=user_name)
    computer.save()
    
    c ={'revovery_password':computer.recovery_key, 'serial':computer.serial, 'username':computer.username, }
    return HttpResponse(simplejson.dumps(c), mimetype="application/json")
        