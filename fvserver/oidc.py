from mozilla_django_oidc.auth import OIDCAuthenticationBackend


def update_user(user, claims):
	user.username = claims.get("email").split("@")[0]
	user.first_name = claims.get("given_name")
	user.last_name = claims.get("family_name")
	user.save()

	return user


class CustomOIDC(OIDCAuthenticationBackend):
	def create_user(self, claims):
		user = super().create_user(claims)
		return update_user(user, claims)

	def update_user(self, user, claims):
		user = super().update_user(user, claims)
		return update_user(user, claims)
