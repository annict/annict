# frozen_string_literal: true

RSpec.configure do |config|
  config.before :each do
    allow(Keen).to receive(:publish)
  end
end
