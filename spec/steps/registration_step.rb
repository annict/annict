# frozen_string_literal: true

module RegistrationStep
  rspec

  def click_signin_with_twitter_link
    visit "/"
    find(".ann-navbar .sign-up").click
    find("#signup-modal").click_link("Twitterアカウントで登録")
  end
end
