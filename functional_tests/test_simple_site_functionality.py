import time
from .base import FunctionalTest
from selenium.webdriver.common.keys import Keys

class LoginAndBasicFunctionality(FunctionalTest):
    def test_admin_can_create_and_browse(self):
        #Admin goes to fv2 key mgmt site, sees it's named Crypt post-redirect to a login
        self.browser.get(self.live_server_url)
        self.assertIn('Crypt', self.browser.title)
        username_box = self.browser.find_element_by_id('id_username')
        password_box = self.browser.find_element_by_id('id_password')
        username_box.send_keys('admin')
        password_box.send_keys('sekrit')
        password_box.send_keys(Keys.ENTER)
        time.sleep(1)
        #After putting in creds, admin can create a computer from the hamburger menu, and is redirected to details
        self.browser.find_element_by_id("dLabel").click()
        self.browser.find_element_by_link_text('New computer').click()
        inputbox = self.browser.find_element_by_id('id_serial')
        self.assertEqual(
            inputbox.get_attribute('placeholder'),
            'Serial Number'
        )
        inputbox.send_keys('MYSERIAL')
        username = self.browser.find_element_by_id('id_username')
        username.send_keys('Mr. Admin')
        computername = self.browser.find_element_by_id('id_computername')
        computername.send_keys('compy486')
        self.browser.find_element_by_css_selector('button.btn.btn-primary').click()
        detail_url = self.browser.current_url
        self.assertRegexpMatches(detail_url, '/info/.+')
        #When viewing details of computer, admin can create a secret for it
        self.browser.find_element_by_class_name('dropdown-toggle').click()
        self.browser.find_element_by_css_selector('span.glyphicon.glyphicon-plus').click()
        secretbox = self.browser.find_element_by_name('secret')
        self.assertEqual(
            secretbox.get_attribute('placeholder'),
            'Secret'
        )
        secretbox.send_keys('LICE-NSEP-LATE')
        self.browser.find_element_by_css_selector('button.btn.btn-primary').click()
        #The newly created secret shows up on the page, and you can click info
        self.browser.find_element_by_css_selector('a.btn.btn-info.btn-xs').click()
        #You're taken to the secret's info page, and you can start a request and provide a reason
        self.browser.find_element_by_css_selector('a.btn.btn-large.btn-info').click()
        requestbox = self.browser.find_element_by_name('reason_for_request')
        self.assertEqual(
            requestbox.get_attribute('placeholder'),
            'Reason for request'
        )
        requestbox.send_keys('Pretty Please Gimme')
        self.browser.find_element_by_css_selector('button.btn.primary.btn-default').click()
        #As the admin is all-powerful, they are automatically approved and can find the secret in the page text
        key = self.browser.find_element_by_tag_name('code').text
        self.assertEqual(key, 'LICE-NSEP-LATE')

    def test_standard_user_can_request_and_admin_can_approve(self):
        #Standard tech user can log in and finds previously-created computer+secret
        self.browser.get(self.live_server_url)
        self.assertIn('Crypt', self.browser.title)
        username_box = self.browser.find_element_by_id('id_username')
        password_box = self.browser.find_element_by_id('id_password')
        username_box.send_keys('tech')
        password_box.send_keys('password')
        password_box.send_keys(Keys.ENTER)
        self.browser.find_element_by_link_text('Info').click()
        self.browser.find_element_by_link_text('Info / Request').click()
        secret_url = self.browser.current_url
        self.assertRegexpMatches(secret_url, '/info/secret/.+')
        self.browser.find_element_by_link_text('Request Key').click()
        requestbox = self.browser.find_element_by_name('reason_for_request')
        requestbox.send_keys('With sugar on top')
        self.browser.find_element_by_css_selector('button.btn.primary.btn-default').click()
        #Standard users live in a world ruled by gravity, and must wait for approval
        disabled_button = self.browser.find_element_by_css_selector('button.btn.btn-disabled.btn-info').text
        self.assertEqual(disabled_button, 'Request Pending')
        #Let's log out and let the admin do their approval magic
        self.browser.find_element_by_id("dLabel").click()
        self.browser.find_element_by_link_text('Log out').click()
        username_box = self.browser.find_element_by_id('id_username')
        password_box = self.browser.find_element_by_id('id_password')
        username_box.send_keys('admin')
        password_box.send_keys('sekrit')
        password_box.send_keys(Keys.ENTER)
        self.browser.find_element_by_link_text('Approve requests').click()
        #This should fail, as per https://github.com/grahamgilbert/Crypt-Server/issues/12
        self.browser.find_element_by_link_text('Manage').click()


