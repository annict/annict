RSpec.configure do |config|
  config.before :suite do
    DatabaseRewinder.clean_all
    # or
    # DatabaseRewinder.clean_with :any_arg_that_would_be_actually_ignored_anyway
  end

  config.after :each do
    DatabaseRewinder.clean
  end
end
