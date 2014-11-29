include Warden::Test::Helpers

Warden.test_mode!

RSpec.configure do |config|
  config.include Devise::TestHelpers, type: :controller
end
