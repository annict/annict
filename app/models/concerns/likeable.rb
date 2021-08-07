# frozen_string_literal: true

module Likeable
  extend ActiveSupport::Concern

  included do
    def likeable?
      true
    end
  end
end
