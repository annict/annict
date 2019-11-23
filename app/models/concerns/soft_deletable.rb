# frozen_string_literal: true

module SoftDeletable
  extend ActiveSupport::Concern

  included do
    scope :without_deleted, -> { where(deleted_at: nil) }
    scope :deleted, -> { where.not(deleted_at: nil) }

    def soft_delete
      touch :deleted_at
    end

    def not_deleted?
      deleted_at.nil?
    end
  end
end
