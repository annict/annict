module RegistrationStep
  rspec

  def click_signin_with_twitter_link
    visit '/'
    find('.welcome').click_link('Twitterアカウントでログイン')
  end
end