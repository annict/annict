# typed: false
# frozen_string_literal: true

module Unpublishable
  extend ActiveSupport::Concern

  include SoftDeletable

  included do
    scope :published, -> { where(unpublished_at: nil) }
    scope :unpublished, -> { where.not(unpublished_at: nil) }
    scope :without_deleted, -> { where(deleted_at: nil) }

    def self.only_kept
      without_deleted.published
    end

    def publish
      update(unpublished_at: nil)
    end

    def unpublish
      update(unpublished_at: Time.zone.now)
    end

    def published?
      unpublished_at.nil?
    end
  end
end
