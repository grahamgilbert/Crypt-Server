from models import *
from django.contrib.auth.decorators import login_required, permission_required
from django.template import RequestContext, Template, Context
import json
from django.views.decorators.csrf import csrf_exempt, csrf_protect
from django.http import HttpResponse, Http404
from django.contrib.auth.models import Permission, User
from django.conf import settings
from django.core.context_processors import csrf
from django.shortcuts import render_to_response, get_object_or_404, redirect
from datetime import datetime, timedelta
from django.db.models import Q
from forms import *
# Create your views here.

##clean up old requests
def cleanup():
    how_many_days = 7
    the_requests = Request.objects.filter(date_approved__lte=datetime.now()-timedelta(days=how_many_days)).filter(current=True)
    for the_req in the_requests:
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
@login_required
def computer_info(request, computer_id):
    cleanup()
    computer = get_object_or_404(Computer, pk=computer_id)
    can_request = None
    approved = None

    # Get the secrets, annotated with whethere there are approvals for them
    secrets = computer.secret_set.all().prefetch_related()

    for secret in secrets:
        secret.approved = Request.objects.filter(requesting_user=request.user).filter(approved=True).filter(current=True).filter(secret=secret)
        secret.pending = Request.objects.filter(requesting_user=request.user).filter(approved__isnull=True).filter(secret=secret)

    print vars(secrets)
    c = {'user': request.user, 'computer':computer, 'secrets':secrets }

    return render_to_response('server/computer_info.html', c, context_instance=RequestContext(request))

@login_required
def secret_info(request, secret_id):
    cleanup()

    secret = get_object_or_404(Secret, pk=secret_id)

    computer = secret.computer

    ##check if the user has outstanding request for this
    pending = secret.request_set.filter(requesting_user=request.user).filter(approved__isnull=True)
    if pending.count() == 0:
        can_request = True
    else:
        can_request = False
    ##if it's been approved, we'll show a link to retrieve the key
    approved = secret.request_set.filter(requesting_user=request.user).filter(approved=True).filter(current=True)

    c = {'user': request.user, 'computer':computer, 'can_request':can_request, 'approved':approved, 'secret':secret}
    if approved.count() != 0:
        return render_to_response('server/secret_approved_button.html', c, context_instance=RequestContext(request))
    else:
        return render_to_response('server/secret_request_button.html', c, context_instance=RequestContext(request))

##request key view
@login_required
def request(request, secret_id):
    ##we will auto approve this if the user has the right perms
    secret = get_object_or_404(Secret, pk=secret_id)
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
            new_request.secret = secret
            if approver:
                new_request.auth_user = request.user
                new_request.approved = True
                new_request.date_approved = datetime.now()
            else:
                # User isn't an approver, send an email to all of the approvers
                perm = Permission.objects.get(codename='can_approve')
                users = User.objects.filter(Q(groups__permissions=perm) | Q(user_permissions=perm) ).distinct()
                print users
                for user in users:
                    if user.email:
                        print user.email
            new_request.save()
            ##if we're an approver, we'll redirect to the retrieve view
            if approver:
                return redirect('server.views.retrieve', new_request.id)
            else:
                return redirect('server.views.secret_info', secret.id)
    else:
        form = RequestForm()
    c = {'form': form, 'secret':secret, }
    return render_to_response('server/request.html', c, context_instance=RequestContext(request))



##retrieve key view
@login_required
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

# Add new manual computer
@login_required
def new_computer(request):
    c = {}
    c.update(csrf(request))
    if request.method == 'POST':
        form = ComputerForm(request.POST)
        if form.is_valid():
            new_computer = form.save(commit=False)
            new_computer.save()
            form.save_m2m()
            return redirect('computer_info', new_computer.id)
    else:
        form = ComputerForm()
    c = {'form': form}
    return render_to_response('server/new_computer_form.html', c, context_instance=RequestContext(request))


@login_required
def new_secret(request, computer_id):
    c = {}
    c.update(csrf(request))
    computer = get_object_or_404(Computer, pk=computer_id)
    if request.method == 'POST':
        form = SecretForm(request.POST)
        if form.is_valid():
            new_secret = form.save(commit=False)
            new_secret.computer = computer
            new_secret.date_escrowed = datetime.now()
            new_secret.save()
            #form.save_m2m()
            return redirect('computer_info', computer.id)
    else:
        form = SecretForm()

    c = {'form': form, 'computer': computer, }
    return render_to_response('server/new_secret_form.html', c, context_instance=RequestContext(request))


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

    try:
        secret_type = request.POST['secret_type']
    except:
        secret_type = 'recovery_key'

    try:
        computer = Computer.objects.get(serial=serial)
    except Computer.DoesNotExist:
        computer = Computer(serial=serial)
    #computer = Computer(recovery_key=recovery_pass, serial=serial_num, last_checkin = datetime.now(), username=user_name, computername=macname)
    computer.last_checkin = datetime.now()
    computer.username=user_name
    computer.computername = macname
    computer.secret_type = secret_type
    computer.save()

    secret = Secret(computer=computer, secret=recovery_pass, date_escrowed=datetime.now())
    secret.save()

    c ={'revovery_password':secret.secret, 'serial':computer.serial, 'username':computer.username, }
    return HttpResponse(json.dumps(c), mimetype="application/json")
