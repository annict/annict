# typed: false
# frozen_string_literal: true

module SystemSpecHelper
  # ユーザーとしてログインする
  def sign_in_as(user)
    visit new_user_session_path
    fill_in "user[email]", with: user.email
    fill_in "user[password]", with: user.password
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
