module RegistrationStep
  rspec

  def click_signin_with_twitter_link
    visit '/'
    click_link 'Twitterアカウントでログイン'
  end
end