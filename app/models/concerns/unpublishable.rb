# frozen_string_literal: true

module Unpublishable
  extend ActiveSupport::Concern

  include SoftDeletable

  included do
    # Unpublishable.only_kept を定義するため、SoftDeletable.only_kept の定義を取り消す
    class << self; undef :only_kept; end

    scope :published, -> { where(unpublished_at: nil) }
    scope :unpublished, -> { where.not(unpublished_at: nil) }
    scope :only_kept, -> { without_deleted.published }

    def unpublish
      touch :unpublished_at
    end
  end
end
