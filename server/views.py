from models import *
from django.contrib.auth.decorators import login_required, permission_required
from django.template import RequestContext, Template, Context
from django.utils import simplejson
from django.views.decorators.csrf import csrf_exempt, csrf_protect
from django.http import HttpResponse, Http404
from django.contrib.auth.models import Permission, User
from django.conf import settings
from django.core.context_processors import csrf
from django.shortcuts import render_to_response, get_object_or_404, redirect
from datetime import datetime, timedelta
from forms import *
# Create your views here.

##clean up old requests
def cleanup():
    how_many_days = 7
    the_requests = Request.objects.filter(date_approved__lte=datetime.now()-timedelta(days=how_many_days)).filter(current=True)
    for the_req in the_requests:
        logger.debug(the_req.current)
        the_req.current = False
        the_req.save()

##index view
@login_required 
def index(request):
    cleanup()
    #show table with all the keys
    computers = Computer.objects.all()
    
    ##get the number of oustanding requests - approved equals null
    outstanding = Request.objects.filter(approved__isnull=True)
    c = {'user': request.user, 'computers':computers, 'outstanding':outstanding, }
    return render_to_response('server/index.html', c, context_instance=RequestContext(request)) 
    
##view to see computer info
def computer_info(request, computer_id):
    cleanup()
    computer = get_object_or_404(Computer, pk=computer_id)
    ##check if the user has outstanding request for this
    pending = Request.objects.filter(requesting_user=request.user).filter(approved__isnull=True).filter(computer=computer)
    if pending.count() == 0:
        can_request = True
    else:
        can_request = False
    ##if it's been approved, we'll show a link to retrieve the key
    approved = Request.objects.filter(requesting_user=request.user).filter(approved=True).filter(current=True).filter(computer=computer)
    c = {'user': request.user, 'computer':computer, 'can_request':can_request, 'approved':approved, }
    if approved.count() != 0:
        return render_to_response('server/computer_approved_button.html', c, context_instance=RequestContext(request))
    else:
        return render_to_response('server/computer_request_button.html', c, context_instance=RequestContext(request))
    

##request key view
def request(request, computer_id):
    ##we will auto approve this if the user has the right perms
    computer = get_object_or_404(Computer, pk=computer_id)
    if request.user.has_perm('server.can_approve'):
        approver = True
    else:
        approver = False
    c = {}
    c.update(csrf(request))
    if request.method == 'POST':
        form = RequestForm(request.POST)
        if form.is_valid():
            new_request = form.save(commit=False)
            new_request.requesting_user = request.user
            new_request.computer = computer
            if approver:
                new_request.auth_user = request.user
                new_request.approved = True
                new_request.date_approved = datetime.now()
            new_request.save()
            ##if we're an approver, we'll redirect to the retrieve view
            if approver:
                return redirect('server.views.retrieve', new_request.id)
            else:
                return redirect('server.views.computer_info', computer.id)
    else:
        form = RequestForm()
    c = {'form': form, 'computer':computer, }
    return render_to_response('server/request.html', c, context_instance=RequestContext(request))

    
    
##retrieve key view
def retrieve(request, request_id):
    cleanup()
    the_request = get_object_or_404(Request, pk=request_id)
    if the_request.approved == True and the_request.current==True:
        c = {'user': request.user, 'the_request':the_request, }
        return render_to_response('server/retrieve.html', c, context_instance=RequestContext(request))
    else:
        raise Http404

##approve key view
@permission_required('server.can_approve', login_url='/login/')
def approve(request, request_id):
    the_request = get_object_or_404(Request, pk=request_id)
    c = {}
    c.update(csrf(request))
    if request.method == 'POST':
        form = ApproveForm(request.POST, instance=the_request)
        if form.is_valid():
            new_request = form.save(commit=False)
            new_request.auth_user = request.user
            new_request.date_approved = datetime.now()
            new_request.save()
            return redirect('server.views.managerequests')
    else:
        form = ApproveForm(instance=the_request)
    c = {'form':form, 'user': request.user, 'the_request':the_request, }
    return render_to_response('server/approve.html', c, context_instance=RequestContext(request))

##manage requests
@permission_required('server.can_approve', login_url='/login/')
def managerequests(request):
    requests = Request.objects.filter(approved__isnull=True)
    c = {'user': request.user, 'requests':requests, }
    return render_to_response('server/manage_requests.html', c, context_instance=RequestContext(request))
    

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
        macname = request.POST['macname']
    except:
        raise Http500
        
    try:
        user_name = request.POST['username']
    except:
        raise Http500
    computer = Computer(recovery_key=recovery_pass, serial=serial_num, last_checkin = datetime.now(), username=user_name, computername=macname)
    computer.save()
    
    c ={'revovery_password':computer.recovery_key, 'serial':computer.serial, 'username':computer.username, }
    return HttpResponse(simplejson.dumps(c), mimetype="application/json")
        