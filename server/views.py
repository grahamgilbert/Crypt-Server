import logging
from .models import *
from django.contrib.auth.decorators import login_required, permission_required
from django.template import RequestContext, Template, Context
import json
import pytz
import copy
from django.views.decorators.csrf import csrf_exempt, csrf_protect
from django.http import HttpResponse, Http404, JsonResponse
from django.contrib.auth.models import Permission, User
from django.conf import settings
from django.template.context_processors import csrf
from django.shortcuts import render, get_object_or_404, redirect
from datetime import datetime, timedelta
from django.db.models import Q
from .forms import *
from django.views.defaults import server_error
from django.core.mail import send_mail
from django.conf import settings
from django.urls import reverse
from django.utils.html import escape

# Create your views here.
logger = logging.getLogger(__name__)

##clean up old requests
def cleanup():
    how_many_days = 7
    the_requests = Request.objects.filter(
        date_approved__lte=datetime.now() - timedelta(days=how_many_days)
    ).filter(current=True)
    for the_req in the_requests:
        the_req.current = False
        the_req.save()


def get_server_version():
    current_dir = os.path.dirname(os.path.realpath(__file__))
    version = plistlib.readPlist(
        os.path.join(os.path.dirname(current_dir), "fvserver", "version.plist")
    )
    return version["version"]


##index view
@login_required
def index(request):
    cleanup()
    # show table with all the keys
    computers = Computer.objects.none()

    if hasattr(settings, "ALL_APPROVE"):
        if settings.ALL_APPROVE == True:
            permissions = Permission.objects.all()
            permission = Permission.objects.get(codename="can_approve")
            if request.user.has_perm("server.can_approve") == False:
                request.user.user_permissions.add(permission)
                request.user.save()
    ##get the number of oustanding requests - approved equals null

    outstanding = Request.objects.filter(approved__isnull=True)
    if hasattr(settings, "APPROVE_OWN"):
        if settings.APPROVE_OWN == False:
            outstanding = outstanding.filter(~Q(requesting_user=request.user))
    c = {"user": request.user, "computers": computers, "outstanding": outstanding}
    return render(request, "server/index.html", c)


@login_required
def tableajax(request):
    """Table ajax for dataTables"""
    # Pull our variables out of the GET request
    get_data = request.GET["args"]
    get_data = json.loads(get_data)
    draw = get_data.get("draw", 0)
    start = int(get_data.get("start", 0))
    length = int(get_data.get("length", 0))
    search_value = ""
    if "search" in get_data:
        if "value" in get_data["search"]:
            search_value = get_data["search"]["value"]

    # default ordering
    order_column = 2
    order_direction = "desc"
    order_name = ""
    if "order" in get_data:
        order_column = get_data["order"][0]["column"]
        order_direction = get_data["order"][0]["dir"]
    for column in get_data.get("columns", None):
        if column["data"] == order_column:
            order_name = column["name"]
            break

    machines = Computer.objects.all().values(
        "id", "serial", "username", "computername", "last_checkin"
    )

    order_string = None
    if len(order_name) != 0:
        if order_direction == "desc":
            order_string = "-%s" % order_name
        else:
            order_string = "%s" % order_name

    if len(search_value) != 0:
        searched_machines = machines.filter(
            Q(serial__icontains=search_value)
            | Q(username__icontains=search_value)
            | Q(computername__icontains=search_value)
            | Q(last_checkin__icontains=search_value)
        )

    else:
        searched_machines = machines

    if order_name != "info_button":
        searched_machines = searched_machines.order_by(order_string)

    limited_machines = searched_machines[start : (start + length)]

    return_data = {}
    return_data["draw"] = int(draw)
    return_data["recordsTotal"] = machines.count()
    return_data["recordsFiltered"] = return_data["recordsTotal"]

    return_data["data"] = []
    settings_time_zone = None
    try:
        settings_time_zone = pytz.timezone(settings.TIME_ZONE)
    except Exception:
        pass

    for machine in limited_machines:
        if machine["last_checkin"]:
            # formatted_date = pytz.utc.localize(machine.last_checkin)
            if settings_time_zone:
                formatted_date = (
                    machine["last_checkin"]
                    .astimezone(settings_time_zone)
                    .strftime("%Y-%m-%d %H:%M %Z")
                )
            else:
                formatted_date = machine["last_checkin"].strftime("%Y-%m-%d %H:%M")
        else:
            formatted_date = ""

        serial_link = '<a href="%s">%s</a>' % (
            reverse("server:computer_info", args=[machine["id"]]),
            escape(machine["serial"]),
        )

        computername_link = '<a href="%s">%s</a>' % (
            reverse("server:computer_info", args=[machine["id"]]),
            escape(machine["computername"]),
        )

        info_button = '<a class="btn btn-info btn-xs" href="%s">Info</a>' % (
            reverse("server:computer_info", args=[machine["id"]])
        )

        list_data = [
            serial_link,
            computername_link,
            escape(machine["username"]),
            formatted_date,
            info_button,
        ]
        return_data["data"].append(list_data)

    return JsonResponse(return_data)


##view to see computer info
@login_required
def computer_info(request, computer_id=None):
    cleanup()
    try:
        computer = get_object_or_404(Computer, pk=computer_id)
    except:
        computer = get_object_or_404(Computer, serial=computer_id)
    can_request = None
    approved = None

    # Get the secrets, annotated with whethere there are approvals for them
    secrets = computer.secret_set.all().prefetch_related()

    for secret in secrets:
        secret.approved = (
            Request.objects.filter(requesting_user=request.user)
            .filter(approved=True)
            .filter(current=True)
            .filter(secret=secret)
        )
        secret.pending = (
            Request.objects.filter(requesting_user=request.user)
            .filter(approved__isnull=True)
            .filter(secret=secret)
        )

    c = {"user": request.user, "computer": computer, "secrets": secrets}

    return render(request, "server/computer_info.html", c)


@login_required
def secret_info(request, secret_id):
    cleanup()

    secret = get_object_or_404(Secret, pk=secret_id)

    computer = secret.computer

    ##check if the user has outstanding request for this
    pending = secret.request_set.filter(requesting_user=request.user).filter(
        approved__isnull=True
    )
    if pending.count() == 0:
        can_request = True
    else:
        can_request = False
    ##if it's been approved, we'll show a link to retrieve the key
    approved = (
        secret.request_set.filter(requesting_user=request.user)
        .filter(approved=True)
        .filter(current=True)
    )
    requests = secret.request_set.all()

    c = {
        "user": request.user,
        "computer": computer,
        "can_request": can_request,
        "approved": approved,
        "secret": secret,
        "requests": requests,
    }
    if approved.count() != 0:
        return render(request, "server/secret_approved_button.html", c)
    else:
        return render(request, "server/secret_request_button.html", c)


##request key view
@login_required
def request(request, secret_id):
    ##we will auto approve this if the user has the right perms
    secret = get_object_or_404(Secret, pk=secret_id)
    approver = False
    if request.user.has_perm("server.can_approve"):
        approver = True

    if approver == True:
        if hasattr(settings, "APPROVE_OWN"):
            if settings.APPROVE_OWN == False:
                approver = False
    c = {}
    c.update(csrf(request))
    if request.method == "POST":
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
                perm = Permission.objects.get(codename="can_approve")
                users = User.objects.filter(
                    Q(is_superuser=True)
                    | Q(groups__permissions=perm)
                    | Q(user_permissions=perm)
                ).distinct()

                if hasattr(settings, "HOST_NAME"):
                    server_name = settings.HOST_NAME.rstrip("/")
                else:
                    server_name = "http://crypt"
                if hasattr(settings, "SEND_EMAIL"):
                    if settings.SEND_EMAIL == True:
                        for user in users:
                            if user.email:
                                email_message = """ There has been a new key request by %s. You can review this request at %s%s
                                """ % (
                                    request.user.username,
                                    server_name,
                                    reverse("server:approve", args=[new_request.id]),
                                )
                                if hasattr(settings, "EMAIL_SENDER"):
                                    email_sender = settings.EMAIL_SENDER
                                else:
                                    email_sender = (
                                        "requests@%s" % request.META["SERVER_NAME"]
                                    )

                                logger.info(
                                    "[*] Sending request email to {} from {}".format(
                                        user.email, email_sender
                                    )
                                )
                                if settings.EMAIL_USER and settings.EMAIL_PASSWORD:

                                    authing_user = settings.EMAIL_USER
                                    authing_password = settings.EMAIL_PASSWORD
                                    logger.info(
                                        "[*] Authing to mail server as {}".format(
                                            authing_user
                                        )
                                    )

                                    send_mail(
                                        "Crypt Key Request",
                                        email_message,
                                        email_sender,
                                        [user.email],
                                        fail_silently=True,
                                        auth_user=authing_user,
                                        auth_password=authing_password,
                                    )
                                else:
                                    send_mail(
                                        "Crypt Key Request",
                                        email_message,
                                        email_sender,
                                        [user.email],
                                        fail_silently=True,
                                    )

            ##if we're an approver, we'll redirect to the retrieve view
            if approver:
                return redirect("server:retrieve", new_request.id)
            else:
                return redirect("server:secret_info", secret.id)
    else:
        form = RequestForm()
    c = {"form": form, "secret": secret}
    return render(request, "server/request.html", c)


##retrieve key view
@login_required
def retrieve(request, request_id):
    cleanup()
    the_request = get_object_or_404(Request, pk=request_id)
    if the_request.approved == True and the_request.current == True:
        if hasattr(settings, "ROTATE_VIEWED_SECRETS"):
            if settings.ROTATE_VIEWED_SECRETS:
                the_request.secret.rotation_required = True
                the_request.secret.save()
        c = {"user": request.user, "the_request": the_request}
        return render(request, "server/retrieve.html", c)
    else:
        raise Http404


## approve key view
@permission_required("server.can_approve", login_url="/login/")
def approve(request, request_id):
    the_request = get_object_or_404(Request, pk=request_id)
    c = {}
    c.update(csrf(request))
    if request.method == "POST":
        form = ApproveForm(request.POST, instance=the_request)
        if form.is_valid():
            new_request = form.save(commit=False)
            new_request.auth_user = request.user
            new_request.date_approved = datetime.now()
            new_request.save()

            # Send an email to the requester with a link to retrieve (or not)
            if hasattr(settings, "HOST_NAME"):
                server_name = settings.HOST_NAME.rstrip("/")
            else:
                server_name = "http://crypt"
            if new_request.approved == True:
                request_status = "approved"
            elif new_request.approved == False:
                request_status = "denied"
            if hasattr(settings, "SEND_EMAIL"):
                if settings.SEND_EMAIL == True:
                    if new_request.requesting_user.email:
                        email_message = """ Your key request has been %s by %s. %s%s
                        """ % (
                            request_status,
                            request.user.username,
                            server_name,
                            reverse("server:secret_info", args=[new_request.secret.id]),
                        )
                        if hasattr(settings, "EMAIL_SENDER"):
                            email_sender = settings.EMAIL_SENDER
                        else:
                            email_sender = "requests@%s" % request.META["SERVER_NAME"]

                        logger.info(
                            "[*] Sending approved/denied email to {} from {}".format(
                                new_request.requesting_user.email, email_sender
                            )
                        )
                        if settings.EMAIL_USER and settings.EMAIL_PASSWORD:

                            authing_user = settings.EMAIL_USER
                            authing_password = settings.EMAIL_PASSWORD
                            logger.info(
                                "[*] Authing to mail server as {}".format(authing_user)
                            )

                            send_mail(
                                "Crypt Key Request",
                                email_message,
                                email_sender,
                                [new_request.requesting_user.email],
                                fail_silently=True,
                                auth_user=authing_user,
                                auth_password=authing_password,
                            )
                        else:
                            send_mail(
                                "Crypt Key Request",
                                email_message,
                                email_sender,
                                [new_request.requesting_user.email],
                                fail_silently=True,
                            )
            return redirect("server:managerequests")
    else:
        form = ApproveForm(instance=the_request)
    c = {"form": form, "user": request.user, "the_request": the_request}
    return render(request, "server/approve.html", c)


##manage requests
@permission_required("server.can_approve", login_url="/login/")
def managerequests(request):
    requests = Request.objects.filter(approved__isnull=True)
    if hasattr(settings, "APPROVE_OWN"):
        if settings.APPROVE_OWN == False:
            requests = requests.filter(~Q(requesting_user=request.user))
    c = {"user": request.user, "requests": requests}
    return render(request, "server/manage_requests.html", c)


# Add new manual computer
@login_required
def new_computer(request):
    c = {}
    c.update(csrf(request))
    if request.method == "POST":
        form = ComputerForm(request.POST)
        if form.is_valid():
            new_computer = form.save(commit=False)
            new_computer.save()
            form.save_m2m()
            return redirect("server:computer_info", new_computer.id)
    else:
        form = ComputerForm()
    c = {"form": form}
    return render(request, "server/new_computer_form.html", c)


@login_required
def new_secret(request, computer_id):
    c = {}
    c.update(csrf(request))
    computer = get_object_or_404(Computer, pk=computer_id)
    if request.method == "POST":
        form_data = copy.copy(request.POST)
        form_data["computer"] = computer.id
        form = SecretForm(data=form_data)
        if form.is_valid():
            new_secret = form.save(commit=False)
            new_secret.computer = computer
            new_secret.date_escrowed = datetime.now()
            new_secret.save()
            # form.save_m2m()
            return redirect("server:computer_info", computer.id)
    else:
        form = SecretForm()

    c = {"form": form, "computer": computer}
    return render(request, "server/new_secret_form.html", c)


# Verify key escrow
@csrf_exempt
def verify(request, serial, secret_type):
    computer = get_object_or_404(Computer, serial=serial)
    try:
        secret = Secret.objects.filter(
            computer=computer, secret_type=secret_type
        ).latest("date_escrowed")
        output = {"escrowed": True, "date_escrowed": secret.date_escrowed}
    except Secret.DoesNotExist:
        output = {"escrowed": False}
    return JsonResponse(output)


##checkin view
@csrf_exempt
def checkin(request):
    try:
        serial_num = request.POST["serial"]
    except:
        return HttpResponse(status=500)
    try:
        recovery_pass = request.POST["recovery_password"]
    except:
        return HttpResponse(status=500)

    try:
        macname = request.POST["macname"]
    except:
        macname = serial_num

    try:
        user_name = request.POST["username"]
    except:
        return HttpResponse(status=500)

    try:
        secret_type = request.POST["secret_type"]
    except:
        secret_type = "recovery_key"

    try:
        computer = Computer.objects.get(serial=serial_num)
    except Computer.DoesNotExist:
        computer = Computer(serial=serial_num)
    # computer = Computer(recovery_key=recovery_pass, serial=serial_num, last_checkin = datetime.now(), username=user_name, computername=macname)
    computer.last_checkin = datetime.now()
    computer.username = user_name
    computer.computername = macname
    computer.secret_type = secret_type
    computer.save()

    try:
        secret = Secret(
            computer=computer,
            secret=recovery_pass,
            secret_type=secret_type,
            date_escrowed=datetime.now(),
        )
        secret.save()
    except ValidationError:
        pass

    latest_secret = (
        Secret.objects.filter(secret_type=secret_type)
        .filter(computer_id=computer.id)
        .latest("date_escrowed")
    )
    rotation_required = latest_secret.rotation_required

    c = {
        "serial": computer.serial,
        "username": computer.username,
        "rotation_required": rotation_required,
    }
    return HttpResponse(json.dumps(c), content_type="application/json")
