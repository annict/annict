# frozen_string_literal: true

# Disable Keen IO when the project ID has not been specified.
if Rails.env.development? && !ENV.key?("KEEN_PROJECT_ID")
  Keen.class_eval do
    def self.publish(_, _); end
  end
end
