# frozen_string_literal: true

module Publishable
  extend ActiveSupport::Concern

  included do
    scope :published, -> { where.not(unpublished_at: nil) }

    def published?
      unpublished_at.nil?
    end

    def unpublish
      touch :unpublished_at
    end
  end
end
