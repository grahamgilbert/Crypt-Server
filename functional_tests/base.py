from django.contrib.staticfiles.testing import StaticLiveServerTestCase
from django.contrib.auth.models import User
from selenium import webdriver
from server.models import Computer, Secret
from datetime import datetime
class FunctionalTest(StaticLiveServerTestCase):

    def setUp(self):
        self.browser = webdriver.Firefox()
        User.objects.create_superuser('admin', 'a@a.com', 'sekrit')
        User.objects._create_user('tech', 't@a.com', 'password', is_staff=True, is_superuser=False)
        tech_test_computer = Computer(serial = 'TECHSERIAL')
        tech_test_computer.username = 'Daft Tech'
        tech_test_computer.computername ='compy587'
        tech_test_computer.save()
        secret = Secret(computer = tech_test_computer, secret = 'SHHH-DONT-TELL', date_escrowed = datetime.now())
        secret.save()


    def tearDown(self):
        self.browser.quit()

    # currently doesn't work to find entered elements
    # def check_for_row_in_list_table(self, row_value):
    #     table = self.browser.find_element_by_id('id_list_table')
    #     rows = table.find_elements_by_tag_name('tr')
    #     value = rows.find_elements_by_tag_name('td')
    #     self.assertIn(row_value, [value.text for row in rows])
