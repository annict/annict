# frozen_string_literal: true

module Appearable
  extend ActiveSupport::Concern

  included do
    scope :appeared, -> { where(disappeared_at: nil) }
    scope :disappeared, -> { where.not(disappeared_at: nil) }

    def disappear
      touch :disappeared_at
    end
  end
end
