# frozen_string_literal: true

module SoftDeletable
  extend ActiveSupport::Concern

  included do
    scope :deleted, -> { where.not(deleted_at: nil) }

    def self.only_kept
      where(deleted_at: nil)
    end

    def soft_delete
      touch :deleted_at
    end

    def not_deleted?
      deleted_at.nil?
    end

    def deleted?
      !not_deleted?
    end
  end
end
