RSpec.configure do |config|
  config.before :each do
    Keen.stub(:publish)
  end
end
