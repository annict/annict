# typed: false

include Warden::Test::Helpers

Warden.test_mode!

RSpec.configure do |config|
  config.include Devise::Test::ControllerHelpers, type: :controller
end
