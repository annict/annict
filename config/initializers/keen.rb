# frozen_string_literal: true

# Disable `Keen.publish` if `KEEN_PROJECT_ID` is not defined.
if Rails.env.development? && ENV["KEEN_PROJECT_ID"].blank?
  Keen.class_eval do
    def self.publish(event_collection, properties); end
  end
end
