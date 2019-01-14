from django import forms
from .models import *


class RequestForm(forms.ModelForm):
    class Meta:
        model = Request
        fields = ("reason_for_request",)


class ApproveForm(forms.ModelForm):
    # approved = forms.BooleanField()
    approved = forms.TypedChoiceField(
        coerce=lambda x: bool(int(x)),
        choices=((1, "Approved"), (0, "Denied")),
        widget=forms.RadioSelect,
        label="Approved?",
    )

    class Meta:
        model = Request
        fields = ("approved", "reason_for_approval")


class ComputerForm(forms.ModelForm):
    class Meta:
        model = Computer
        fields = ("serial", "username", "computername")


class SecretForm(forms.ModelForm):
    class Meta:
        model = Secret
        fields = ("secret_type", "secret", "computer")
        widgets = {"computer": forms.HiddenInput()}
