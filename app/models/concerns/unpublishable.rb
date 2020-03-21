# frozen_string_literal: true

module Unpublishable
  extend ActiveSupport::Concern

  include SoftDeletable

  included do
    scope :published, -> { where(unpublished_at: nil) }
    scope :unpublished, -> { where.not(unpublished_at: nil) }
    scope :only_kept, -> { without_deleted.published }

    def unpublish
      touch :unpublished_at
    end
  end
end
