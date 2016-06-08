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
from django.views.defaults import server_error
from django.core.mail import send_mail
from django.conf import settings
from django.core.urlresolvers import reverse
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

    if hasattr(settings, 'ALL_APPROVE'):
        if settings.ALL_APPROVE == True:
            permissions = Permission.objects.all()
            permission = Permission.objects.get(codename='can_approve')
            if request.user.has_perm('server.can_approve') == False:
                request.user.user_permissions.add(permission)
                request.user.save()
    ##get the number of oustanding requests - approved equals null

    outstanding = Request.objects.filter(approved__isnull=True)
    if hasattr(settings, 'APPROVE_OWN'):
        if settings.APPROVE_OWN == False:
            outstanding = outstanding.filter(~Q(requesting_user=request.user))
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
    requests = secret.request_set.all()

    c = {'user': request.user, 'computer':computer, 'can_request':can_request, 'approved':approved, 'secret':secret, 'requests':requests}
    if approved.count() != 0:
        return render_to_response('server/secret_approved_button.html', c, context_instance=RequestContext(request))
    else:
        return render_to_response('server/secret_request_button.html', c, context_instance=RequestContext(request))

##request key view
@login_required
def request(request, secret_id):
    ##we will auto approve this if the user has the right perms
    secret = get_object_or_404(Secret, pk=secret_id)
    approver = False
    if request.user.has_perm('server.can_approve'):
        approver = True

    if approver == True:
        if hasattr(settings, 'APPROVE_OWN'):
            if settings.APPROVE_OWN == False:
                approver = False
    c = {}
    c.update(csrf(request))
    if request.method == 'POST':
        form = RequestForm(request.POST)
        if form.is_valid():
            new_request = form.save(commit=False)
            new_request.requesting_user = request.user
            new_request.secret = secret
            new_request.save()
            if approver:
                new_request.auth_user = request.user
                new_request.approved = True
                new_request.date_approved = datetime.now()
                new_request.save()
            else:
                # User isn't an approver, send an email to all of the approvers
                perm = Permission.objects.get(codename='can_approve')
                users = User.objects.filter(Q(is_superuser=True) | Q(groups__permissions=perm) | Q(user_permissions=perm) ).distinct()

                if hasattr(settings, 'HOST_NAME'):
                    server_name = settings.HOST_NAME.rstrip('/')
                else:
                    server_name = 'http://crypt'
                if hasattr(settings, 'SEND_EMAIL'):
                    if settings.SEND_EMAIL == True:
                        for user in users:
                            if user.email:
                                email_message = """ There has been a new key request by %s. You can review this request at %s%s
                                """ % (request.user.username, server_name, reverse('server.views.approve', args=[new_request.id]))
                                email_sender = 'requests@%s' % request.META['SERVER_NAME']
                                send_mail('Crypt Key Request', email_message, email_sender,
        [user.email], fail_silently=True)

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

## approve key view
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

            # Send an email to the requester with a link to retrieve (or not)
            if hasattr(settings, 'HOST_NAME'):
                server_name = settings.HOST_NAME.rstrip('/')
            else:
                server_name = 'http://crypt'
            if new_request.approved == True:
                request_status = 'approved'
            elif new_request.approved == False:
                request_status = 'denied'
            if hasattr(settings, 'SEND_EMAIL'):
                if settings.SEND_EMAIL == True:
                    if new_request.requesting_user.email:
                        email_message = """ Your key request has been %s by %s. %s%s
                        """ % (request_status, request.user.username, server_name, reverse('server.views.secret_info', args=[new_request.id]))
                        email_sender = 'requests@%s' % request.META['SERVER_NAME']
                        send_mail('Crypt Key Request', email_message, email_sender,
    [new_request.requesting_user.email], fail_silently=True)
            return redirect('server.views.managerequests')
    else:
        form = ApproveForm(instance=the_request)
    c = {'form':form, 'user': request.user, 'the_request':the_request, }
    return render_to_response('server/approve.html', c, context_instance=RequestContext(request))

##manage requests
@permission_required('server.can_approve', login_url='/login/')
def managerequests(request):
    requests = Request.objects.filter(approved__isnull=True)
    if hasattr(settings, 'APPROVE_OWN'):
        if settings.APPROVE_OWN == False:
            requests = requests.filter(~Q(requesting_user=request.user))
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
        return HttpResponse(status=500)
    try:
        recovery_pass = request.POST['recovery_password']
    except:
        return HttpResponse(status=500)

    try:
        macname = request.POST['macname']
    except:
        macname = serial_num

    try:
        user_name = request.POST['username']
    except:
        return HttpResponse(status=500)

    try:
        secret_type = request.POST['secret_type']
    except:
        secret_type = 'recovery_key'

    try:
        computer = Computer.objects.get(serial=serial_num)
    except Computer.DoesNotExist:
        computer = Computer(serial=serial_num)
    #computer = Computer(recovery_key=recovery_pass, serial=serial_num, last_checkin = datetime.now(), username=user_name, computername=macname)
    computer.last_checkin = datetime.now()
    computer.username=user_name
    computer.computername = macname
    computer.secret_type = secret_type
    computer.save()

    try:
        secret = Secret(computer=computer, secret=recovery_pass, secret_type=secret_type, date_escrowed=datetime.now())
        secret.save()
    except ValidationError:
        pass

    c ={'revovery_password':secret.secret, 'serial':computer.serial, 'username':computer.username, }
    return HttpResponse(json.dumps(c), content_type="application/json")
