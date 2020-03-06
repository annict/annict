# frozen_string_literal: true
# https://github.com/plataformatec/devise/wiki/How-To:-Redirect-to-a-specific-page-when-the-user-can-not-be-authenticated

class Annict::Devise::CustomFailure < Devise::FailureApp
  def redirect_url
    root_url
  end
end
