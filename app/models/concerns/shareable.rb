# frozen_string_literal: true

module Shareable
  extend ActiveSupport::Concern

  included do
    def share_url_with_query(provider)
      utm = {
        utm_source: provider,
        utm_medium: "#{self.class.name.downcase}_share",
        utm_campaign: user.username
      }

      "#{share_url}?#{utm.to_query}"
    end
  end
end
