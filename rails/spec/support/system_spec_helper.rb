# typed: false
# frozen_string_literal: true

module SystemSpecHelper
  def sign_in(user:, password: "passw0rd")
    visit "/legacy/sign_in"
    fill_in "user[email_username]", with: user.email
    fill_in "user[password]", with: password
    click_button "ログイン"
  end

  # ページの読み込みを待つ
  def wait_for_turbo
    has_css?("[data-turbo-visit-control]", wait: 0)
  end

  # JavaScriptが実行されるまで待つ
  def wait_for_javascript
    Timeout.timeout(Capybara.default_max_wait_time) do
      loop until page.evaluate_script("document.readyState") == "complete"
    end
  end
end

RSpec.configure do |config|
  config.include SystemSpecHelper, type: :system
end
