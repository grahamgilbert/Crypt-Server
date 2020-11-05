from django.test import TestCase, Client
from django.contrib.auth.models import User
from datetime import datetime
from server.models import Computer, Secret, Request


class RequestProcess(TestCase):
    def test_request_passes_correct_data_to_template(self):
        """
        Correct request login request.

        Args:
            self: (todo): write your description
        """
        admin = User.objects.create_superuser("admin", "a@a.com", "sekrit")
        tech = User.objects.create_user("tech", "a@a.com", "password")
        tech.save()
        tech_test_computer = Computer(
            serial="TECHSERIAL", username="Daft Tech", computername="compy587"
        )
        tech_test_computer.save()
        test_secret = Secret(
            computer=tech_test_computer,
            secret="SHHH-DONT-TELL",
            date_escrowed=datetime.now(),
        )
        test_secret.save()
        secret_request = Request(secret=test_secret, requesting_user=tech)
        secret_request.save()
        client = Client()
        login_response = self.client.post(
            "/login/", {"username": "admin", "password": "sekrit"}, follow=True
        )
        response = self.client.get("/manage-requests/", follow=True)
        print(response)
        self.assertTrue(response.context["user"].is_authenticated)
