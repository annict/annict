# typed: false
# frozen_string_literal: true

module SystemSpecHelper
  include Warden::Test::Helpers

  def sign_in(user:, password: nil)
    login_as(user, scope: :user)
    visit root_path
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
