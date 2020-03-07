# frozen_string_literal: true

module Publishable
  extend ActiveSupport::Concern

  included do
    scope :published, -> { appeared.without_deleted }
  end
end
