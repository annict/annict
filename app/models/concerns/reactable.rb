# frozen_string_literal: true

module Reactable
  extend ActiveSupport::Concern

  included do
    def reactable?
      true
    end
  end
end
