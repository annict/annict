# typed: false
# frozen_string_literal: true

module RootResourceCommon
  extend ActiveSupport::Concern

  included do
    def root_resource?
      true
    end
  end
end
