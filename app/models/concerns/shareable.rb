# frozen_string_literal: true

module Shareable
  extend ActiveSupport::Concern

  included do
    def share_url_with_query(provider)
      utm = {
        utm_source: provider,
        utm_medium: "#{self.class.name.underscore}_share",
        utm_campaign: user.username
      }

      "#{share_url}?#{utm.to_query}"
    end

    def setup_shared_sns(user)
      self.share_to_twitter =
        user.twitter.present? &&
        user.authorized_to?(:twitter, shareable: true) &&
        user.share_record_to_twitter?
    end
  end
end
